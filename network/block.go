package network

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/observe-fi/indexer/app"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"log"
)

var client *liteclient.ConnectionPool
var masterBlock *ton.BlockIDExt
var api ton.APIClientWrapped

func MustConnect(ctx context.Context) {
	client = liteclient.NewConnectionPool()

	var url string
	if app.IsTestnet() {
		url = "https://ton-blockchain.github.io/testnet-global.config.json"
	} else {
		url = "https://ton.org/global.config.json"
	}

	cfg, err := liteclient.GetConfigFromUrl(ctx, url)
	if err != nil {
		log.Fatalln("get config err: ", err.Error())
		return
	}

	// connect to mainnet lite servers
	err = client.AddConnectionsFromConfig(ctx, cfg)
	if err != nil {
		log.Fatalln("connection err: ", err.Error())
		return
	}

	// initialize ton api lite connection wrapper with full proof checks
	api = ton.NewAPIClient(client, ton.ProofCheckPolicySecure).WithRetry()
	api.SetTrustedBlockFromConfig(cfg)

	log.Println("checking proofs since config init block, it may take near a minute...")

	masterBlock, err = api.GetMasterchainInfo(ctx)
	if err != nil {
		log.Fatalln("get masterchain info err: ", err.Error())
		return
	}

}

func (p *Provider) Connect() (err error) {
	p.client = liteclient.NewConnectionPool()

	var url string
	if app.IsTestnet() {
		url = "https://ton-blockchain.github.io/testnet-global.config.json"
	} else {
		url = "https://ton.org/global.config.json"
	}

	cfg, err := liteclient.GetConfigFromUrl(p.ctx, url)
	if err != nil {
		return
	}

	// connect to mainnet lite servers
	err = p.client.AddConnectionsFromConfig(p.ctx, cfg)
	if err != nil {
		return
	}

	// initialize ton api lite connection wrapper with full proof checks
	p.api = ton.NewAPIClient(p.client, ton.ProofCheckPolicySecure).WithRetry()
	p.api.SetTrustedBlockFromConfig(cfg)

	p.log.Info("checking proofs since config init block, it may take near a minute...")

	masterBlock, err = p.api.GetMasterchainInfo(p.ctx)
	if err != nil {
		return
	}

	return nil
}

func (p *Provider) Disconnect() {
	p.client.Stop()
}

func CurrentMasterBlock() *ton.BlockIDExt {
	return masterBlock
}

func MasterBlockAt(ctx context.Context, seqNo uint32) *ton.BlockIDExt {
	block, err := api.LookupBlock(ctx, masterBlock.Workchain, masterBlock.Shard, seqNo)
	if err != nil {
		log.Fatalln("get masterchain info err: ", err.Error())
		return nil
	}
	return block
}

func BlockWatcher(ctx context.Context, starting *ton.BlockIDExt) {
	master := starting
	ctx = api.Client().StickyContext(ctx)
	shardLastSeqno := map[string]uint32{}

	firstShards, err := api.GetBlockShardsInfo(ctx, master)
	if err != nil {
		log.Fatalln("get shards err:", err.Error())
		return
	}
	for _, shard := range firstShards {
		shardLastSeqno[getShardID(shard)] = shard.SeqNo
	}

	for {
		log.Printf("scanning %d master block...\n", master.SeqNo)

		// getting information about other work-chains and shards of master block
		currentShards, err := api.GetBlockShardsInfo(ctx, master)
		if err != nil {
			log.Fatalln("get shards err:", err.Error())
			return
		}

		// shards in master block may have holes, e.g. shard seqno 2756461, then 2756463, and no 2756462 in master chain
		// thus we need to scan a bit back in case of discovering a hole, till last seen, to fill the misses.
		var newShards []*ton.BlockIDExt
		for _, shard := range currentShards {
			notSeen, err := getNotSeenShards(ctx, api, shard, shardLastSeqno)
			if err != nil {
				log.Fatalln("get not seen shards err:", err.Error())
				return
			}
			shardLastSeqno[getShardID(shard)] = shard.SeqNo
			newShards = append(newShards, notSeen...)
		}
		newShards = append(newShards, master)

		var txList []*tlb.Transaction

		// for each shard block getting transactions
		for _, shard := range newShards {
			log.Printf("scanning block %d of shard %x in workchain %d...", shard.SeqNo, uint64(shard.Shard), shard.Workchain)

			var fetchedIDs []ton.TransactionShortInfo
			var after *ton.TransactionID3
			var more = true

			// load all transactions in batches with 100 transactions in each while exists
			for more {
				fetchedIDs, more, err = api.WaitForBlock(master.SeqNo).GetBlockTransactionsV2(ctx, shard, 100, after)
				if err != nil {
					log.Fatalln("get tx ids err:", err.Error())
					return
				}

				if more {
					// set load offset for next query (pagination)
					after = fetchedIDs[len(fetchedIDs)-1].ID3()
				}

				for _, id := range fetchedIDs {
					// get full transaction by id
					tx, err := api.GetTransaction(ctx, shard, address.NewAddress(0, byte(shard.Workchain), id.Account), id.LT)
					if err != nil {
						log.Fatalln("get tx data err:", err.Error())
						return
					}
					txList = append(txList, tx)
				}
			}
		}

		for i, transaction := range txList {

			log.Println(i, transaction.String())
			log.Println(base64.URLEncoding.EncodeToString(transaction.Hash))
		}

		if len(txList) == 0 {
			log.Printf("no transactions in %d block\n", master.SeqNo)
		}

		master = MasterBlockAt(ctx, master.SeqNo+1)
		//master, err = MasterBlockAt(ctx, master.SeqNo + 1)
		//if err != nil {
		//	log.Fatalln("get masterchain info err: ", err.Error())
		//	return
		//}
	}
}

func getShardID(shard *ton.BlockIDExt) string {
	return fmt.Sprintf("%d|%d", shard.Workchain, shard.Shard)
}

func getNotSeenShards(ctx context.Context, api ton.APIClientWrapped, shard *ton.BlockIDExt, shardLastSeqno map[string]uint32) (ret []*ton.BlockIDExt, err error) {
	if no, ok := shardLastSeqno[getShardID(shard)]; ok && no == shard.SeqNo {
		return nil, nil
	}

	b, err := api.GetBlockData(ctx, shard)
	if err != nil {
		return nil, fmt.Errorf("get block data: %w", err)
	}

	parents, err := b.BlockInfo.GetParentBlocks()
	if err != nil {
		return nil, fmt.Errorf("get parent blocks (%d:%x:%d): %w", shard.Workchain, uint64(shard.Shard), shard.Shard, err)
	}

	for _, parent := range parents {
		ext, err := getNotSeenShards(ctx, api, parent, shardLastSeqno)
		if err != nil {
			return nil, err
		}
		ret = append(ret, ext...)
	}

	ret = append(ret, shard)
	return ret, nil
}
