package indexer

import (
	"encoding/base64"
	"fmt"
	"github.com/observe-fi/indexer/db"
	"github.com/xssnick/tonutils-go/tlb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"os"
)

type Txs struct {
	*db.Collection
}

type Msg struct {
	Type tlb.MsgType `bson:"type"`
	Data string      `bson:"data"`
}

type Tx struct {
	ID          primitive.ObjectID `bson:"_id"`
	Address     string             `bson:"address"`
	Now         uint32             `bson:"now"`
	OrigStatus  tlb.AccountStatus  `bson:"orig-status"`
	EndStatus   tlb.AccountStatus  `bson:"end-status"`
	InMessage   *Msg               `bson:"in-msg"`
	OutMessages []Msg              `bson:"out-messages"`
	FullTx      string             `bson:"full-tx"`
	Hint        string             `bson:"hint"`
}

func (p *Provider) TxCollection() *Txs {
	c := p.db.LoadCollection(fmt.Sprintf("indexer-txs-%s", os.Getenv("NETWORK")))
	return &Txs{
		Collection: c,
	}
}

func tlbEncode(d interface{}) (string, error) {
	v, e := tlb.ToCell(d)
	if e != nil {
		return "", e
	}
	return base64.StdEncoding.EncodeToString(v.ToBOC()), nil
}

func (txs *Txs) Store(tx *tlb.Transaction, addr string) error {
	outs := make([]Msg, 0)
	if tx.IO.Out != nil {
		o, e := tx.IO.Out.ToSlice()
		if e != nil {
			return e
		}
		for _, om := range o {
			enc, e := tlbEncode(om.Msg)
			if e != nil {
				return e
			}
			outs = append(outs, Msg{
				Type: om.MsgType,
				Data: enc,
			})
		}
	}

	var in *Msg
	if tx.IO.In != nil {
		enc, e := tlbEncode(tx.IO.In.Msg)
		if e != nil {
			return e
		}
		in = &Msg{
			Type: tx.IO.In.MsgType,
			Data: enc,
		}
	}

	nTx := Tx{
		ID:          primitive.NewObjectID(),
		Address:     addr,
		Now:         tx.Now,
		OrigStatus:  tx.OrigStatus,
		EndStatus:   tx.EndStatus,
		InMessage:   in,
		OutMessages: outs,
		FullTx:      "",
		Hint:        tx.String(),
	}

	e := txs.Create(nTx)
	return e
}
