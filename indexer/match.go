package indexer

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/observe-fi/indexer/db"
	"github.com/observe-fi/indexer/network"
	"github.com/xssnick/tonutils-go/tlb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"os"
)

type Match struct {
	*db.Collection
	data []MatchCondition
}

type MatchType uint32

const (
	CodeMatch MatchType = iota
	TxMatch
	AddressMatch
)

type MatchCondition struct {
	ID          primitive.ObjectID `bson:"_id"`
	Type        MatchType          `bson:"type"`
	TargetValue string             `bson:"target-value"`
}

func (p *Provider) MatchCollection() *Match {
	c := p.db.LoadCollection(fmt.Sprintf("indexer-match-%s", os.Getenv("NETWORK")))
	return &Match{
		Collection: c,
	}
}

func (m *Match) Load() error {
	var res []MatchCondition
	e := m.ReadAll(context.Background(), bson.M{}, &res)
	if e != nil {
		return e
	}
	m.data = res
	return nil
}

func (m *MatchCondition) Matches(tx *tlb.Transaction, account *tlb.Account, addr string) bool {
	var h string

	if m.Type == CodeMatch {
		h = base64.StdEncoding.EncodeToString(account.Code.Hash())
	} else if m.Type == TxMatch {
		h = base64.StdEncoding.EncodeToString(tx.Hash)
	} else if m.Type == AddressMatch {
		h = addr
	}

	return h == m.TargetValue
}

func (m *Match) Matches(tx *tlb.Transaction, account *tlb.Account, addr string) bool {
	for _, v := range m.data {
		if v.Matches(tx, account, addr) {
			return true
		}
	}
	return false
}

func (m *Match) FilterBlock(block *network.BlockWithTx) *network.BlockWithTx {
	newBlock := &network.BlockWithTx{
		MasterBlock: block.MasterBlock,
		TxList:      make([]*tlb.Transaction, 0),
		Accounts:    make(map[string]*tlb.Account),
		TxAccounts:  make(map[string]string),
	}
	for _, tx := range block.TxList {
		hb64 := base64.StdEncoding.EncodeToString(tx.Hash)
		addr := block.TxAccounts[hb64]
		account := block.Accounts[addr]
		matches := m.Matches(tx, account, addr)
		if matches {
			newBlock.TxList = append(newBlock.TxList, tx)
			newBlock.Accounts[addr] = account
			newBlock.TxAccounts[hb64] = addr
		}
	}
	return block
}
