package db

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

var client *mongo.Client

func MustConnect(ctx context.Context) {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	uri := os.Getenv("DB_URI")
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	var err error
	client, err = mongo.Connect(ctx, opts)
	if err != nil {
		panic(err)
	}

	var result bson.M
	if err := client.Database("admin").RunCommand(ctx, bson.D{{"ping", 1}}).Decode(&result); err != nil {
		panic(err)
	}

	fmt.Println("Database is connected!")
}

func (p *Provider) Connect() (err error) {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	uri := os.Getenv("DB_URI")
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	p.client, err = mongo.Connect(p.ctx, opts)
	if err != nil {
		return
	}

	var result bson.M
	if err = p.client.Database("admin").RunCommand(p.ctx, bson.D{{"ping", 1}}).Decode(&result); err != nil {
		return
	}

	p.log.Info("DB Connected Successfully!")
	return nil
}

func (p *Provider) Disconnect() (err error) {
	return p.client.Disconnect(p.ctx)
}
