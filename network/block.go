package network

import (
	"encoding/base64"
	"fmt"
	"github.com/observe-fi/indexer/app"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
)

func (p *Provider) Connect() (err error) {
	p.client = liteclient.NewConnectionPool()

	var url string
	if app.IsTestnet() {
		url = "https://ton.org/testnet-global.config.json"
	} else {
		url = "https://ton.org/global.config.json"
	}

	cfg, err := liteclient.GetConfigFromUrl(p.ctx, url)
	if err != nil {
		return
	}

	// connect to main-net lite servers
	err = p.client.AddConnectionsFromConfig(p.ctx, cfg)
	if err != nil {
		return
	}

	// initialize ton api lite connection wrapper with full proof checks
	p.api = ton.NewAPIClient(p.client, ton.ProofCheckPolicySecure).WithRetry()
	// Init an unchecked proofs api
	// Why this is happening? => Because we want to get account states too which may have no proofs for specific blocks on liteservers
	// TODO: Maybe make this more secure?!
	p.uncheckedApi = ton.NewAPIClient(p.client, ton.ProofCheckPolicyUnsafe).WithRetry()

	s := p.StateCollection()
	blk := s.TrustedBlock()
	if blk != nil {
		p.api.SetTrustedBlock(blk)
	} else {
		p.api.SetTrustedBlockFromConfig(cfg)
	}

	p.log.Info("checking proofs since config init block, it may take near a minute...")

	p.masterBlock, err = p.api.GetMasterchainInfo(p.ctx)
	if err != nil {
		return
	}

	err = s.SaveTrustedBlock(p.masterBlock)
	if err != nil {
		return
	}

	return nil
}

func (p *Provider) Disconnect() {
	p.client.Stop()
}

func (p *Provider) CurrentMasterBlock() *ton.BlockIDExt {
	return p.masterBlock
}

func (p *Provider) MasterBlockAt(seqNo uint32) (blk *ton.BlockIDExt, err error) {
	blk, err = p.api.LookupBlock(p.ctx, p.masterBlock.Workchain, p.masterBlock.Shard, seqNo)
	if err != nil {
		return
	}
	return
}

func (p *Provider) BlockWatcher(starting *ton.BlockIDExt, rx chan *BlockWithTx) error {
	master := starting
	ctx := p.api.Client().StickyContext(p.ctx)
	ctxUnchecked := p.uncheckedApi.Client().StickyContext(p.ctx)

	shardLastSeqNo := map[string]uint32{}

	firstShards, err := p.api.GetBlockShardsInfo(ctx, master)
	if err != nil {
		return err
	}
	for _, shard := range firstShards {
		shardLastSeqNo[getShardID(shard)] = shard.SeqNo
	}

	for {
		p.log.Infow("Scanning Master Block", "seq-no", master.SeqNo)

		// getting information about other work-chains and shards of master block
		currentShards, err := p.api.GetBlockShardsInfo(ctx, master)
		if err != nil {
			return err
		}

		// shards in master block may have holes, e.g. shard seqno 2756461, then 2756463, and no 2756462 in master chain
		// thus we need to scan a bit back in case of discovering a hole, till last seen, to fill the misses.
		var newShards []*ton.BlockIDExt
		for _, shard := range currentShards {
			notSeen, err := p.getNotSeenShards(shard, shardLastSeqNo)
			if err != nil {
				return err
			}
			shardLastSeqNo[getShardID(shard)] = shard.SeqNo
			newShards = append(newShards, notSeen...)
		}
		newShards = append(newShards, master)

		var txList []*tlb.Transaction
		accounts := make(map[string]*tlb.Account)
		// for each shard block getting transactions
		for _, shard := range newShards {
			p.log.Infow("Scanning block", "seq-no", shard.SeqNo, "shard", uint64(shard.Shard), "workchain", shard.Workchain)

			var fetchedIDs []ton.TransactionShortInfo
			var after *ton.TransactionID3
			var more = true

			// load all transactions in batches with 100 transactions in each while exists
			for more {
				fetchedIDs, more, err = p.api.WaitForBlock(master.SeqNo).GetBlockTransactionsV2(ctx, shard, 100, after)
				if err != nil {
					return err
				}

				if more {
					// set load offset for next query (pagination)
					after = fetchedIDs[len(fetchedIDs)-1].ID3()
				}

				for _, id := range fetchedIDs {
					// get full transaction by id
					addr := address.NewAddress(0, byte(shard.Workchain), id.Account)
					tx, err := p.api.GetTransaction(ctx, shard, addr, id.LT)
					if err != nil {
						return err
					}
					// get transaction account - for account based indexing
					_, ok := accounts[addr.String()]
					if !ok {
						acc, err := p.uncheckedApi.GetAccount(ctxUnchecked, shard, addr)
						if err != nil {
							return err
						}
						accounts[addr.String()] = acc
					}
					txList = append(txList, tx)
				}
			}
		}

		for i, transaction := range txList {
			p.log.Infow("Transaction spotted", "index", i, "data", transaction.String(), "hash", base64.URLEncoding.EncodeToString(transaction.Hash))
		}

		if len(txList) == 0 {
			p.log.Infow("No Tx found in block!", "seq-no", master.SeqNo)
		}

		rx <- &BlockWithTx{MasterBlock: master, TxList: txList, Accounts: accounts}

		master, err = p.MasterBlockAt(master.SeqNo + 1)
		if err != nil {
			return err
		}
	}
}

func getShardID(shard *ton.BlockIDExt) string {
	return fmt.Sprintf("%d|%d", shard.Workchain, shard.Shard)
}

func (p *Provider) getNotSeenShards(shard *ton.BlockIDExt, shardLastSeqno map[string]uint32) (ret []*ton.BlockIDExt, err error) {
	if no, ok := shardLastSeqno[getShardID(shard)]; ok && no == shard.SeqNo {
		return nil, nil
	}

	b, err := p.api.GetBlockData(p.ctx, shard)
	if err != nil {
		return nil, fmt.Errorf("get block data: %w", err)
	}

	parents, err := b.BlockInfo.GetParentBlocks()
	if err != nil {
		return nil, fmt.Errorf("get parent blocks (%d:%x:%d): %w", shard.Workchain, uint64(shard.Shard), shard.Shard, err)
	}

	for _, parent := range parents {
		ext, err := p.getNotSeenShards(parent, shardLastSeqno)
		if err != nil {
			return nil, err
		}
		ret = append(ret, ext...)
	}

	ret = append(ret, shard)
	return ret, nil
}
