package indexer

import (
	"context"
	"errors"
	"github.com/observe-fi/indexer/db"
	"github.com/observe-fi/indexer/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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
	c := p.db.LoadCollection("indexer-state")
	return &State{
		c,
	}
}

func (state *State) LastBlock() uint32 {
	var value StateValue[uint32]
	e := state.ReadID(context.Background(), "last-block", &value)
	if errors.Is(e, mongo.ErrNoDocuments) {
		return 0
	}
	return value.Value
}

func getState[V any](state *State, key string) *StateValue[V] {
	var value StateValue[V]
	e := state.ReadID(context.Background(), key, &value)
	if e != nil {
		return nil
	}
	return &value
}

func createState[V any](key string, value V) *StateValue[V] {
	return &StateValue[V]{
		ID:     primitive.NewObjectID(),
		HashID: util.HashID(key),
		Value:  value,
	}
}

func setState[V any](state *State, key string, value V) error {
	var st StateValue[V]
	stv := &st
	err := state.ReadID(context.Background(), key, stv)
	if err != nil {
		stv = createState(key, value)
		return state.Create(stv)
	} else {
		stv.Value = value
		return state.Update(context.Background(), db.LookupID(key), &bson.M{"$set": stv})
	}
}

func (state *State) SetLastBlock(seqNo uint32) error {
	return setState(state, "last-block", seqNo)
}
