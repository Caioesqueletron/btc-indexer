# Bitcoin Blockchain Indexer

A production-oriented educational Bitcoin blockchain indexer written in Go.

## Architecture

```text
Bitcoin Core (regtest)
        |
        | JSON-RPC polling
        v
Block Synchronizer
        |
        v
Transactional Block Processor
        |
        v
PostgreSQL
  - blocks
  - transactions
  - inputs
  - outputs / UTXO state
  - sync checkpoint
        |
        v
gRPC API
```

The MVP uses JSON-RPC polling. ZMQ-based real-time ingestion is planned for a later phase.

## Why this project matters

This project demonstrates:

- Go application structure and concurrency
- Bitcoin blockchain data modeling
- Transactional persistence
- Idempotent processing
- Crash recovery through checkpoints
- PostgreSQL indexing and transactions
- gRPC API design
- Reorg-handling foundations
- Production-oriented architecture

## Requirements

- Go 1.24+
- Docker and Docker Compose
- PostgreSQL client tools
- `protoc`
- `protoc-gen-go`
- `protoc-gen-go-grpc`

## Run locally

```bash
make infra
```

Apply the schema:

```bash
psql "postgres://postgres:postgres@localhost:5432/btcindexer"   -f migrations/001_init.sql
```

Generate protobuf code:

```bash
make proto
```

Start the indexer:

```bash
make indexer
```

Start the API in another terminal:

```bash
make api
```

Generate regtest blocks:

```bash
docker exec -it $(docker ps -qf name=bitcoin) bitcoin-cli   -regtest -rpcuser=bitcoin -rpcpassword=bitcoin   -generate 101
```

## gRPC methods

- `GetBlockByHeight`
- `GetTransaction`

## Transactional processing

Each block is persisted inside one database transaction:

1. Insert block
2. Insert transactions
3. Insert inputs
4. Mark referenced outputs as spent
5. Insert outputs
6. Update synchronization checkpoint
7. Commit

If the process crashes before commit, PostgreSQL rolls back the block and the checkpoint remains unchanged.

## Reorganization roadmap

A complete production implementation should:

1. Detect previous-hash mismatches
2. Find the common ancestor
3. Roll back orphaned blocks
4. Restore spent outputs
5. Remove outputs created by orphaned transactions
6. Select the branch with the most cumulative work
7. Index the canonical branch atomically

The current schema is an MVP and should evolve to support competing branches and canonical-chain flags.

## Roadmap

### Phase 1 - MVP

- JSON-RPC client
- Block synchronization
- PostgreSQL schema
- Transaction/input/output indexing
- Basic UTXO state
- Checkpointing
- gRPC API
- Graceful shutdown
- Idempotent writes

### Phase 2 - Bitcoin correctness

- Canonical chain tracking
- Reorg detection and rollback
- UTXO restoration
- Chain-work fork selection
- Coinbase maturity
- Better script/address decoding
- Exact satoshi parsing

### Phase 3 - Performance

- Batch inserts
- PostgreSQL COPY
- Worker pools
- Backpressure
- Benchmarks
- CPU and memory profiling

### Phase 4 - Real-time ingestion

- ZMQ block notifications
- ZMQ transaction notifications
- At-least-once delivery
- Deduplication
- Polling fallback

### Phase 5 - Production hardening

- Prometheus metrics
- OpenTelemetry
- Structured logging
- Health/readiness endpoints
- Integration tests
- Chaos and recovery tests
- CI/CD
- Kubernetes deployment

Do not publish invented benchmark numbers. Measure actual throughput, transaction rate, database writes per second, memory usage, recovery time, and API latency.
