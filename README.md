# Indexer-Go

This repository contains the initial version of our indexer that acts as a PoC which will let us build other parts of the system upon it. The final version will be heavily optimized and the choice of technologies may change. This codebase uses `fx` for dependency injection.

## Quickstart
1. Spawn a mongo instance on your system or do it on cloud.
2. Make a copy of `.example.env` file and name it `.env`
3. Change the mongo URI:
```
DB_URI=<YOUR MONGO URI>
NETWORK=testnet
START_BLOCK=-1
DB_NAME=indexer
STORAGE_STRATEGY=STORE_ALL
```
4. Run:
```shell
go mod download
go build -o /indexer-go
./indexer-go
```

***Experimental: Alternatively, you can use the docker files in the repo, they may need some adjustment***

The tool will start to monitor blocks on the specified network from the block you specified, setup with a auto-restart capable manager, either docker or sth like `pm2`.

## Adding filters
You can change `STORAGE_STRATEGY` to `OPTIMIZED` and add documents to the `indexer-match-{network}` collection. The tool will only store the transactions with satisfying conditions. Schema:

```json5
{
	"_id": "ObjectId('...')",
	"type": 0, // 0 - CodeMatch; 1 - Tx Hash Match; 2 - Address Match;
	"target-value": "expected value" // for code and tx is b64 hash, and for address is standard address
}
```

## Future
Right now, this tool is mostly intended for internal use of our project, and it will have complete documentation and an easy-to-use interface in near future. 