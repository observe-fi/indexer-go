package network

import (
	"context"
	"github.com/observe-fi/indexer/db"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Provider struct {
	client      *liteclient.ConnectionPool
	api         ton.APIClientWrapped
	log         *zap.SugaredLogger
	ctx         context.Context
	masterBlock *ton.BlockIDExt
	db          *db.Provider
}

func NewProvider(life fx.Lifecycle, log *zap.Logger, db *db.Provider) *Provider {
	provider := &Provider{log: log.Sugar(), db: db}
	life.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// TODO: We should manage this better
			provider.ctx = context.Background()
			return provider.Connect()
		},
		OnStop: func(ctx context.Context) error {
			provider.ctx = ctx
			provider.Disconnect()
			return nil
		},
	})
	return provider
}
