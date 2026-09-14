package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/caio/btc-indexer/internal/bitcoin"
    "github.com/caio/btc-indexer/internal/indexer"
    "github.com/caio/btc-indexer/internal/storage"
)

func getenv(key, fallback string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return fallback
}

func main() {
    ctx, stop := signal.NotifyContext(
        context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    bitcoinClient := bitcoin.NewClient(
        getenv("BITCOIN_RPC_URL", "http://localhost:18443"),
        getenv("BITCOIN_RPC_USER", "bitcoin"),
        getenv("BITCOIN_RPC_PASSWORD", "bitcoin"),
    )

    store, err := storage.New(ctx,
        getenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/btcindexer"))
    if err != nil {
        log.Fatal(err)
    }
    defer store.Close()

    sync := indexer.NewSynchronizer(bitcoinClient, store)
    if err := sync.Run(ctx); err != nil && ctx.Err() == nil {
        log.Fatal(err)
    }
}
