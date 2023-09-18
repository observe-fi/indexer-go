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
	client       *liteclient.ConnectionPool
	api          ton.APIClientWrapped
	uncheckedApi ton.APIClientWrapped
	log          *zap.SugaredLogger
	ctx          context.Context
	masterBlock  *ton.BlockIDExt
	db           *db.Provider
	started      bool
	startCh      chan bool
}

func NewProvider(life fx.Lifecycle, log *zap.Logger, db *db.Provider) *Provider {
	provider := &Provider{log: log.Sugar(), db: db, startCh: make(chan bool), started: false}
	life.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// TODO: We should manage this better
			provider.ctx = context.Background()
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
			provider.Disconnect()
			return nil
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
