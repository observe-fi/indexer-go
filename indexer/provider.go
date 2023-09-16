package indexer

import (
	"context"
	"errors"
	"fmt"
	"github.com/observe-fi/indexer/db"
	"github.com/observe-fi/indexer/network"
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

func NewProvider(life fx.Lifecycle, log *zap.Logger, db *db.Provider, net *network.Provider) *Provider {
	provider := &Provider{log: log.Sugar(), db: db, network: net}
	life.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			provider.ctx = ctx
			state := provider.StateCollection()
			blk := state.LastBlock()
			if blk == 0 {
				stBlk := os.Getenv("START_BLOCK")
				startBlk, err := strconv.ParseUint(stBlk, 10, 32)
				blk = uint32(startBlk)
				if err != nil {
					return errors.New("unable to read start block")
				}
			}
			master, err := net.MasterBlockAt(blk)
			if err != nil {
				return err
			}

			dataChannel := make(chan *network.BlockWithTx)
			go func() {
				for {
					data := <-dataChannel
					fmt.Println("Received data:", data)
				}
			}()
			return net.BlockWatcher(master, dataChannel)
		},
		OnStop: func(ctx context.Context) error {
			provider.ctx = ctx
			return nil
		},
	})
	return provider
}
