package network

import (
	"context"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Provider struct {
	client *liteclient.ConnectionPool
	api    ton.APIClientWrapped
	log    *zap.Logger
	ctx    context.Context
}

func NewProvider(life fx.Lifecycle, log *zap.Logger) *Provider {
	provider := &Provider{log: log}
	life.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			provider.ctx = ctx
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
