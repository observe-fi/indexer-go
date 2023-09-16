package db

import (
	"github.com/observe-fi/indexer/util"
	"go.mongodb.org/mongo-driver/bson"
)

func LookupID(key string) *bson.M {
	return &bson.M{
		"hash-id": util.HashID(key),
	}
}
