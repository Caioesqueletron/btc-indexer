package main

import (
    "context"
    "log"
    "net"
    "os"

    blockchain "github.com/caio/btc-indexer/gen/blockchain"
    "github.com/caio/btc-indexer/internal/grpcapi"
    "github.com/caio/btc-indexer/internal/storage"
    "google.golang.org/grpc"
)

func main() {
    ctx := context.Background()

    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        dsn = "postgres://postgres:postgres@localhost:5432/btcindexer"
    }

    store, err := storage.New(ctx, dsn)
    if err != nil {
        log.Fatal(err)
    }
    defer store.Close()

    addr := os.Getenv("GRPC_ADDR")
    if addr == "" {
        addr = ":50051"
    }

    listener, err := net.Listen("tcp", addr)
    if err != nil {
        log.Fatal(err)
    }

    server := grpc.NewServer()
    blockchain.RegisterBlockchainServiceServer(
        server, grpcapi.NewServer(store))

    log.Printf("gRPC server listening on %s", addr)
    if err := server.Serve(listener); err != nil {
        log.Fatal(err)
    }
}
