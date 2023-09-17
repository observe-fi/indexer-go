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
	started     bool
	startCh     chan bool
}

func NewProvider(life fx.Lifecycle, log *zap.Logger) *Provider {
	provider := &Provider{
		log:         log,
		collections: map[string]*Collection{},
		started:     false,
		startCh:     make(chan bool),
	}
	life.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			provider.ctx = ctx
			e := provider.Connect()
			if e != nil {
				return e
			}
			provider.started = true
			provider.startCh <- true
			return nil
		},
		OnStop: func(ctx context.Context) error {
			provider.ctx = ctx
			return provider.Disconnect()
		},
	})
	return provider
}

func (p *Provider) AwaitStart() {
	if p.started {
		return
	}
	<-p.startCh
}
