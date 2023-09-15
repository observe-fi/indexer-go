package main

import (
	"github.com/observe-fi/indexer/app"
	"github.com/observe-fi/indexer/db"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	app.MustLoadEnv()
	//db.MustConnect(context.Background())
	//
	//network.MustConnect(context.Background())
	//oldBlock := network.MasterBlockAt(context.Background(), 32600000)
	//network.BlockWatcher(context.Background(), oldBlock)

	fx.New(
		fx.Provide(db.NewProvider, zap.NewProduction),
		fx.Invoke(func(p *db.Provider) {}),
	).Run()
}
