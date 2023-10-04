package main

import (
	"github.com/observe-fi/indexer/app"
	"github.com/observe-fi/indexer/db"
	"github.com/observe-fi/indexer/indexer"
	"github.com/observe-fi/indexer/network"
	"github.com/observe-fi/indexer/util"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	app.MustLoadEnv()

	fxApp := fx.New(
		fx.Provide(db.NewProvider, network.NewProvider, zap.NewProduction, indexer.NewProvider),
		fx.Invoke(func(p *indexer.Provider) {
			go func() {
				err := p.Begin()
				if err != nil {
					util.Halt(err)
					// TODO: Handle Error
				}
			}()
		}),
	)
	fxApp.Run()
}
