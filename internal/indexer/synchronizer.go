package indexer

import (
    "context"
    "log"
    "time"

    "github.com/caio/btc-indexer/internal/bitcoin"
    "github.com/caio/btc-indexer/internal/storage"
)

type Synchronizer struct {
    bitcoin *bitcoin.Client
    store   *storage.Store
}

func NewSynchronizer(bitcoinClient *bitcoin.Client, store *storage.Store) *Synchronizer {
    return &Synchronizer{bitcoin: bitcoinClient, store: store}
}

func (s *Synchronizer) Run(ctx context.Context) error {
    ticker := time.NewTicker(2 * time.Second)
    defer ticker.Stop()

    for {
        if err := s.syncOnce(ctx); err != nil {
            log.Printf("sync error: %v", err)
        }

        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
        }
    }
}

func (s *Synchronizer) syncOnce(ctx context.Context) error {
    chainHeight, err := s.bitcoin.GetBlockCount(ctx)
    if err != nil {
        return err
    }

    lastHeight, err := s.store.LastHeight(ctx)
    if err != nil {
        return err
    }

    for height := lastHeight + 1; height <= chainHeight; height++ {
        hash, err := s.bitcoin.GetBlockHash(ctx, height)
        if err != nil {
            return err
        }

        block, err := s.bitcoin.GetBlock(ctx, hash)
        if err != nil {
            return err
        }

        if err := s.store.SaveBlock(ctx, block); err != nil {
            return err
        }

        log.Printf("indexed block height=%d hash=%s txs=%d",
            block.Height, block.Hash, len(block.Tx))
    }

    return nil
}
