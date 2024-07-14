# Indexer-Go

This repository contains our streaming indexer that receives the blocks actively by watching the liteservers. This codebase uses `fx` for dependency injection.

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
### Run with Docker
```sh
docker compose up -d
```

### Run without docker:
```shell
go mod download
go build -o /indexer-go
./indexer-go
```

The tool will start to monitor blocks on the specified network from the block you specified, set up with an auto-restart capable manager, either docker or sth like `pm2`.

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
Right now, this tool is mostly intended for internal use of our project.
