package db

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Provider struct {
	client *mongo.Client
	log    *zap.Logger
	ctx    context.Context
}

func NewProvider(life fx.Lifecycle, log *zap.Logger) *Provider {
	provider := &Provider{
		log: log,
	}
	life.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			provider.ctx = ctx
			e := provider.Connect()
			if e != nil {
				return e
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			provider.ctx = ctx
			return provider.Disconnect()
		},
	})
	return provider
}
