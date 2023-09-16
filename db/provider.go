package db

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Provider struct {
	client      *mongo.Client
	db          *mongo.Database
	log         *zap.Logger
	ctx         context.Context
	collections map[string]*Collection
}

func NewProvider(life fx.Lifecycle, log *zap.Logger) *Provider {
	provider := &Provider{
		log:         log,
		collections: map[string]*Collection{},
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
