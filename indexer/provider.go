package indexer

import (
	"context"
	"github.com/observe-fi/indexer/db"
	"github.com/observe-fi/indexer/network"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Provider struct {
	network *network.Provider
	db      *db.Provider
	log     *zap.SugaredLogger
	ctx     context.Context
}

func NewProvider(life fx.Lifecycle, log *zap.Logger, db *db.Provider, network *network.Provider) *Provider {
	provider := &Provider{log: log.Sugar(), db: db, network: network}
	life.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			provider.ctx = ctx
			master, err := network.MasterBlockAt(32600000)
			if err != nil {
				return err
			}
			return network.BlockWatcher(master)
		},
		OnStop: func(ctx context.Context) error {
			provider.ctx = ctx
			return nil
		},
	})
	return provider
}
