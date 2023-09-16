package db

import "go.mongodb.org/mongo-driver/mongo"

type Collection struct {
	col      *mongo.Collection
	provider *Provider
}
