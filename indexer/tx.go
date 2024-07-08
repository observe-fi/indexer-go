package indexer

import (
	"encoding/base64"
	"fmt"
	"github.com/observe-fi/indexer/db"
	"github.com/observe-fi/indexer/util"
	"github.com/xssnick/tonutils-go/tlb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"os"
)

type Txs struct {
	*db.Collection
}

type Messages struct {
	*db.Collection
}

type Msg struct {
	Type    tlb.MsgType `bson:"type"`
	Data    string      `bson:"data"`
	OptBody string      `bson:"opt_body"`
}

type ProcessedMessage struct {
	ID        primitive.ObjectID `bson:"_id"`
	ParentTx  primitive.ObjectID `bson:"parent_tx"`
	Hash      string             `bson:"hash"`
	From      string             `bson:"from"`
	To        string             `bson:"to"`
	ValueF    float64            `bson:"value"`
	ValueN    string             `bson:"value_n"`
	OptBody   string             `bson:"opt_body"`
	CreatedAt uint32             `bson:"created_at"`
	Out       bool               `bson:"out"`
}

type Tx struct {
	ID         primitive.ObjectID `bson:"_id"`
	Address    string             `bson:"address"`
	Now        uint32             `bson:"now"`
	OrigStatus tlb.AccountStatus  `bson:"orig-status"`
	EndStatus  tlb.AccountStatus  `bson:"end-status"`

	Messages    []primitive.ObjectID `bson:"messages"`
	Hash        string               `bson:"hash"`
	Fee         string               `bson:"fee"`
	Value       float64              `bson:"value"`
	FeeF        float64              `bson:"fee_f"`
	BlockNumber uint32               `bson:"block-number"`
	LT          uint64               `bson:"lt"`
}

func (p *Provider) TxCollection() *Txs {
	c := p.db.LoadCollection(fmt.Sprintf("indexer-txs-%s", os.Getenv("NETWORK")))
	return &Txs{
		Collection: c,
	}
}

func (p *Provider) MsgCollection() *Messages {
	c := p.db.LoadCollection(fmt.Sprintf("indexer-msgs-%s", os.Getenv("NETWORK")))
	return &Messages{
		Collection: c,
	}
}

func tlbEncode(d interface{}) (string, error) {
	v, e := tlb.ToCell(d)
	if e != nil {
		return "", e
	}
	return base64.StdEncoding.EncodeToString(v.ToBOC()), nil
}

func (messages *Messages) Store(msg *ProcessedMessage) error {
	e := messages.Create(msg)
	return e
}

func constructPMFromExternalIn(om *tlb.Message) *ProcessedMessage {
	if om.MsgType != tlb.MsgTypeExternalIn {
		return nil
	}
	external := om.AsExternalIn()
	amountS := "0"
	amountF := float64(0)
	hash := ""
	c, _ := external.ToCell()
	if c != nil {
		hash = base64.StdEncoding.EncodeToString(c.Hash())
	}
	boc := make([]byte, 0)
	if external.Body != nil {
		boc = external.Body.ToBOC()
	}
	pm := &ProcessedMessage{
		ID:        primitive.NewObjectID(),
		From:      "external",
		To:        util.AddressToRaw(external.DstAddr),
		ValueN:    amountS,
		ValueF:    amountF,
		Hash:      hash,
		OptBody:   base64.StdEncoding.EncodeToString(boc),
		CreatedAt: 0,
	}
	return pm
}

func constructPMFromExternalOut(om *tlb.Message) *ProcessedMessage {
	return nil
	// TODO: Support External Out
	//if om.MsgType != tlb.MsgTypeExternalOut {
	//	return nil
	//}
	//external := om.AsExternalOut()
	//amountS := "0"
	//amountF := float64(0)
	//hash := ""
	//boc := make([]byte, 0)
	//if external.Body != nil {
	//	boc = external.Body.ToBOC()
	//}
	//pm := &ProcessedMessage{
	//	ID:        primitive.NewObjectID(),
	//	From:      "external",
	//	To:        util.AddressToRaw(external.DstAddr),
	//	ValueN:    amountS,
	//	ValueF:    amountF,
	//	Hash:      hash,
	//	OptBody:   base64.StdEncoding.EncodeToString(boc),
	//	CreatedAt: 0,
	//}
	//return pm
}

func constructPMFromInternal(om *tlb.Message) *ProcessedMessage {
	if om.MsgType != tlb.MsgTypeInternal {
		return nil
	}
	internal := om.AsInternal()
	amountS := internal.Amount.String()
	amountF, _ := internal.Amount.Nano().Float64()

	hash := ""
	c, _ := internal.ToCell()
	if c != nil {
		hash = base64.StdEncoding.EncodeToString(c.Hash())
	}
	boc := make([]byte, 0)
	if internal.Body != nil {
		boc = internal.Body.ToBOC()
	}
	pm := &ProcessedMessage{
		ID:        primitive.NewObjectID(),
		From:      util.AddressToRaw(internal.SrcAddr),
		To:        util.AddressToRaw(internal.DstAddr),
		ValueN:    amountS,
		ValueF:    amountF,
		Hash:      hash,
		OptBody:   base64.StdEncoding.EncodeToString(boc),
		CreatedAt: internal.CreatedAt,
	}
	return pm
}

func constructPM(om *tlb.Message) *ProcessedMessage {
	if om.MsgType == tlb.MsgTypeInternal {
		return constructPMFromInternal(om)
	} else if om.MsgType == tlb.MsgTypeExternalIn {
		return constructPMFromExternalIn(om)
	} else if om.MsgType == tlb.MsgTypeExternalOut {
		return constructPMFromExternalOut(om)
	}
	return nil
}

func ExtractTxMessages(tx *tlb.Transaction) []*ProcessedMessage {
	messageList := make([]*ProcessedMessage, 0)

	if tx.IO.In != nil {
		pm := constructPM(tx.IO.In)
		if pm != nil {
			pm.Out = false
			messageList = append(messageList, pm)
		}
	}

	if tx.IO.Out != nil {
		x, e := tx.IO.Out.ToSlice()
		if e != nil {
			for _, message := range x {
				pm := constructPM(&message)
				if pm != nil {
					pm.Out = true
					messageList = append(messageList, pm)
				}
			}
		}
	}

	return messageList
}

func (txs *Txs) Store(messages *Messages, tx *tlb.Transaction, addr string, blockNo uint32) (e error) {
	// Let's process the messages first
	processedMessages := ExtractTxMessages(tx)
	ids := make([]primitive.ObjectID, 0)
	txValueFlow := float64(0)
	txId := primitive.NewObjectID()
	for _, item := range processedMessages {
		// Let's calculate value flow
		if item.Out {
			txValueFlow -= item.ValueF
		} else {
			txValueFlow += item.ValueF
		}
		item.ParentTx = txId
		e = messages.Store(item)
		if e != nil {
			return
		}
		ids = append(ids, item.ID)
	}
	feeF, _ := tx.TotalFees.Coins.Nano().Float64()
	txValueFlow -= feeF

	nTx := Tx{
		ID:          txId,
		Address:     addr,
		Now:         tx.Now,
		OrigStatus:  tx.OrigStatus,
		EndStatus:   tx.EndStatus,
		Hash:        base64.StdEncoding.EncodeToString(tx.Hash),
		Messages:    ids,
		LT:          tx.LT,
		Fee:         tx.TotalFees.Coins.String(),
		FeeF:        feeF,
		Value:       txValueFlow,
		BlockNumber: blockNo,
	}

	e = txs.Create(nTx)
	return e
}

//
//func (txs *Txs) Store(tx *tlb.Transaction, addr string) error {
//
//	outs := make([]Msg, 0)
//	if tx.IO.Out != nil {
//		o, e := tx.IO.Out.ToSlice()
//		if e != nil {
//			return e
//		}
//		for _, om := range o {
//			enc, e := tlbEncode(om.Msg)
//			if e != nil {
//				return e
//			}
//			optBody := ""
//			if om.Msg != nil {
//				p := om.Msg.Payload()
//				if p != nil {
//					optBody = base64.StdEncoding.EncodeToString(p.ToBOC())
//				}
//			}
//			outs = append(outs, Msg{
//				Type:    om.MsgType,
//				Data:    enc,
//				OptBody: optBody,
//			})
//		}
//	}
//
//	var in *Msg
//	if tx.IO.In != nil {
//		im := tx.IO.In
//		enc, e := tlbEncode(im.Msg)
//		if e != nil {
//			return e
//		}
//		optBody := ""
//		if im.Msg != nil {
//			p := im.Msg.Payload()
//			if p != nil {
//				optBody = base64.StdEncoding.EncodeToString(p.ToBOC())
//			}
//		}
//
//		in = &Msg{
//			Type:    tx.IO.In.MsgType,
//			Data:    enc,
//			OptBody: optBody,
//		}
//	}
//
//	nTx := Tx{
//		ID:          primitive.NewObjectID(),
//		Address:     addr,
//		Now:         tx.Now,
//		OrigStatus:  tx.OrigStatus,
//		EndStatus:   tx.EndStatus,
//		InMessage:   in,
//		OutMessages: outs,
//		FullTx:      "",
//		Hint:        tx.String(),
//	}
//
//	e := txs.Create(nTx)
//	return e
//}
