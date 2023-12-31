package network

import (
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
)

type BlockWithTx struct {
	MasterBlock *ton.BlockIDExt
	TxList      []*tlb.Transaction
	TxAccounts  map[string]string
	Accounts    map[string]*tlb.Account
}
