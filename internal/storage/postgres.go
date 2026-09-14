package storage

import (
    "context"
    "fmt"

    "github.com/caio/btc-indexer/internal/bitcoin"
    "github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
    db *pgxpool.Pool
}

func New(ctx context.Context, dsn string) (*Store, error) {
    pool, err := pgxpool.New(ctx, dsn)
    if err != nil {
        return nil, err
    }
    if err := pool.Ping(ctx); err != nil {
        pool.Close()
        return nil, err
    }
    return &Store{db: pool}, nil
}

func (s *Store) Close() {
    s.db.Close()
}

func (s *Store) LastHeight(ctx context.Context) (int64, error) {
    var height int64
    err := s.db.QueryRow(ctx,
        `SELECT last_height FROM sync_state WHERE id = TRUE`).Scan(&height)
    return height, err
}

func (s *Store) SaveBlock(ctx context.Context, block bitcoin.Block) error {
    tx, err := s.db.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)

    _, err = tx.Exec(ctx, `
        INSERT INTO blocks
        (hash,height,previous_hash,merkle_root,timestamp,nonce,bits,size)
        VALUES ($1,$2,$3,$4,to_timestamp($5),$6,$7,$8)
        ON CONFLICT (hash) DO NOTHING`,
        block.Hash, block.Height, block.PreviousHash, block.MerkleRoot,
        block.Time, block.Nonce, block.Bits, block.Size)
    if err != nil {
        return fmt.Errorf("insert block: %w", err)
    }

    for txIndex, t := range block.Tx {
        _, err = tx.Exec(ctx, `
            INSERT INTO transactions
            (txid,block_hash,block_height,tx_index,raw_hex)
            VALUES ($1,$2,$3,$4,$5)
            ON CONFLICT (txid) DO NOTHING`,
            t.TxID, block.Hash, block.Height, txIndex, t.Hex)
        if err != nil {
            return fmt.Errorf("insert tx: %w", err)
        }

        for vinIndex, in := range t.Vin {
            if in.TxID == "" {
                continue
            }

            _, err = tx.Exec(ctx, `
                INSERT INTO transaction_inputs
                (txid,vin_index,previous_txid,previous_vout,script_sig,sequence)
                VALUES ($1,$2,$3,$4,$5,$6)
                ON CONFLICT (txid,vin_index) DO NOTHING`,
                t.TxID, vinIndex, in.TxID, in.Vout, in.ScriptSig.Hex, in.Sequence)
            if err != nil {
                return fmt.Errorf("insert input: %w", err)
            }

            _, err = tx.Exec(ctx, `
                UPDATE transaction_outputs
                SET spent = TRUE
                WHERE txid = $1 AND vout = $2`,
                in.TxID, in.Vout)
            if err != nil {
                return fmt.Errorf("mark output spent: %w", err)
            }
        }

        for _, out := range t.Vout {
            address := out.ScriptPubKey.Address
            if address == "" && len(out.ScriptPubKey.Addresses) > 0 {
                address = out.ScriptPubKey.Addresses[0]
            }

            _, err = tx.Exec(ctx, `
                INSERT INTO transaction_outputs
                (txid,vout,value_sats,script_pub_key,address)
                VALUES ($1,$2,$3,$4,$5)
                ON CONFLICT DO NOTHING`,
                t.TxID, out.N, SatsFromBTC(out.Value), out.ScriptPubKey.Hex, address)
            if err != nil {
                return fmt.Errorf("insert output: %w", err)
            }
        }
    }

    _, err = tx.Exec(ctx, `
        UPDATE sync_state
        SET last_height = $1, last_hash = $2
        WHERE id = TRUE`, block.Height, block.Hash)
    if err != nil {
        return fmt.Errorf("update checkpoint: %w", err)
    }

    return tx.Commit(ctx)
}

func (s *Store) GetBlockByHeight(ctx context.Context, height int64) (string, error) {
    var hash string
    err := s.db.QueryRow(ctx,
        `SELECT hash FROM blocks WHERE height=$1`, height).Scan(&hash)
    return hash, err
}

func (s *Store) GetTransaction(ctx context.Context, txid string) (string, int64, error) {
    var blockHash string
    var height int64
    err := s.db.QueryRow(ctx,
        `SELECT block_hash, block_height FROM transactions WHERE txid=$1`, txid).
        Scan(&blockHash, &height)
    return blockHash, height, err
}
