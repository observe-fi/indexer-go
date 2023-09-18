package main

import (
	"fmt"
	"github.com/observe-fi/indexer/app"
	"github.com/observe-fi/indexer/db"
	"github.com/observe-fi/indexer/indexer"
	"github.com/observe-fi/indexer/network"
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
					fmt.Println(err)
					// TODO: Handle Error
				}
			}()
		}),
	)
	fxApp.Run()
}
