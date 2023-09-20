package indexer

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/observe-fi/indexer/db"
	"github.com/observe-fi/indexer/network"
	"github.com/xssnick/tonutils-go/ton"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"os"
	"strconv"
)

type Provider struct {
	network *network.Provider
	db      *db.Provider
	log     *zap.SugaredLogger
	ctx     context.Context
}

const LastPossibleBlock uint32 = 0xffffffff - 1

func NewProvider(life fx.Lifecycle, log *zap.Logger, db *db.Provider, net *network.Provider) *Provider {
	provider := &Provider{log: log.Sugar(), db: db, network: net}
	life.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			provider.ctx = ctx
			return nil
		},
		OnStop: func(ctx context.Context) error {
			provider.ctx = ctx
			return nil
		},
	})
	return provider
}

func (p *Provider) Begin() error {
	p.db.AwaitStart()

	p.network.AwaitStart()

	state := p.StateCollection()
	blk := state.LastBlock()
	if blk == 0 {
		stBlk := os.Getenv("START_BLOCK")
		if stBlk == "-1" {
			// We're starting from last block known
			blk = LastPossibleBlock
		} else {
			startBlk, err := strconv.ParseUint(stBlk, 10, 32)
			blk = uint32(startBlk)
			if err != nil {
				return errors.New("unable to read start block")
			}
		}
	}
	var masterAt *ton.BlockIDExt
	fmt.Println(blk)
	if blk == LastPossibleBlock {
		masterAt = p.network.CurrentMasterBlock()
	} else {
		var err error
		masterAt, err = p.network.MasterBlockAt(blk)
		if err != nil {
			return err
		}
	}

	match := p.MatchCollection()
	err := match.Load()
	if err != nil {
		return err
	}

	dataChannel := make(chan *network.BlockWithTx)
	go func() {
		txs := p.TxCollection()
		accounts := p.AccountsCollection()

		for {
			block := <-dataChannel
			fBlock := match.FilterBlock(block)

			for _, acc := range fBlock.Accounts {
				e := accounts.Store(acc)
				if e != nil {
					panic(e)
				}
			}

			for _, tx := range fBlock.TxList {
				addr := fBlock.TxAccounts[base64.StdEncoding.EncodeToString(tx.Hash)]
				e := txs.Store(tx, addr)
				if e != nil {
					panic(e)
				}
			}

			fmt.Println("Received block:", block)

			e := state.SetLastBlock(block.MasterBlock.SeqNo)
			if e != nil {
				panic(e)
			}
		}
	}()
	return p.network.BlockWatcher(masterAt, dataChannel)
}
