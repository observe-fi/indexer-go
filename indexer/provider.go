package indexer

import (
	"context"
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

	if blk == LastPossibleBlock {
		masterAt = p.network.CurrentMasterBlock()
	} else {
		var err error
		masterAt, err = p.network.MasterBlockAt(blk)
		if err != nil {
			return err
		}
	}

	dataChannel := make(chan *network.BlockWithTx)
	go func() {
		for {
			data := <-dataChannel
			fmt.Println("Received data:", data)
		}
	}()
	return p.network.BlockWatcher(masterAt, dataChannel)
}
