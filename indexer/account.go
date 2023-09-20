package indexer

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/observe-fi/indexer/db"
	"github.com/xssnick/tonutils-go/tlb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"os"
)

type Accounts struct {
	*db.Collection
}

type Account struct {
	ID      primitive.ObjectID `bson:"_id"`
	Active  bool               `bson:"active"`
	Address string             `bson:"address"`
	Status  tlb.AccountStatus  `bson:"status"`
	Balance string             `bson:"balance"`
	Data    string             `bson:"data"`
	Code    string             `bson:"code"`
}

func (p *Provider) AccountsCollection() *Accounts {
	c := p.db.LoadCollection(fmt.Sprintf("indexer-accounts-%s", os.Getenv("NETWORK")))
	return &Accounts{
		Collection: c,
	}
}

func (accounts *Accounts) Store(acc *tlb.Account) error {
	nAcc := Account{
		ID:      primitive.NewObjectID(),
		Active:  acc.IsActive,
		Address: acc.State.Address.String(),
		Status:  acc.State.Status,
		Balance: acc.State.Balance.Nano().String(),
		Data:    base64.StdEncoding.EncodeToString(acc.Data.ToBOC()),
		Code:    base64.StdEncoding.EncodeToString(acc.Code.ToBOC()),
	}
	var account Account
	e := accounts.ReadOne(context.Background(), &bson.M{"address": acc.State.Address.String()}, &account)
	if e != nil {
		// We don't have this account [MOST PROBABLY]
		e = accounts.Create(nAcc)
	} else {
		// We have one guy here, let's use his ID and update
		nAcc.ID = account.ID
		e = accounts.Update(context.Background(), &bson.M{"_id": account.ID}, &bson.M{
			"$set": nAcc,
		})
	}
	return e
}
