package network

import (
	"context"
	"encoding/base64"
	"github.com/observe-fi/indexer/db"
	"github.com/observe-fi/indexer/util"
	"github.com/xssnick/tonutils-go/tl"
	"github.com/xssnick/tonutils-go/ton"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type State struct {
	*db.Collection
}

type StateValue[V any] struct {
	ID     primitive.ObjectID `bson:"_id"`
	HashID string             `bson:"hash-id"`
	Value  V                  `bson:"value"`
}

func (p *Provider) StateCollection() *State {
	c := p.db.LoadCollection("network-state")
	return &State{
		c,
	}
}

func createState[V any](key string, value V) *StateValue[V] {
	return &StateValue[V]{
		HashID: util.HashID(key),
		Value:  value,
	}
}

func setState[V any](state *State, key string, value V) error {
	return state.Upsert(db.LookupID(key), &bson.M{
		"$set": createState(key, value),
	})
}

func getState[V any](state *State, key string) *StateValue[V] {
	var value StateValue[V]
	e := state.ReadID(context.Background(), key, &value)
	if e != nil {
		return nil
	}
	return &value
}

func (state *State) SaveTrustedBlock(blk *ton.BlockIDExt) error {
	b, err := tl.Serialize(blk, true)
	if err != nil {
		return err
	}
	blockStr := base64.StdEncoding.EncodeToString(b)
	return setState(state, "trusted-block", blockStr)
}

func (state *State) TrustedBlock() *ton.BlockIDExt {
	v := getState[string](state, "trusted-block")
	if v == nil {
		return nil
	}
	b64 := v.Value
	tlBytes, e := base64.StdEncoding.DecodeString(b64)
	if e != nil {
		return nil
	}
	var master ton.BlockIDExt
	_, e = tl.Parse(&master, tlBytes, true)
	if e != nil {
		return nil
	}
	return &master
}
