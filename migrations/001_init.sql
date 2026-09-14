CREATE TABLE IF NOT EXISTS blocks (
    hash TEXT PRIMARY KEY,
    height BIGINT NOT NULL UNIQUE,
    previous_hash TEXT,
    merkle_root TEXT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    nonce BIGINT NOT NULL,
    bits TEXT NOT NULL,
    size BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS transactions (
    txid TEXT PRIMARY KEY,
    block_hash TEXT NOT NULL REFERENCES blocks(hash),
    block_height BIGINT NOT NULL,
    tx_index INT NOT NULL,
    raw_hex TEXT
);

CREATE TABLE IF NOT EXISTS transaction_inputs (
    id BIGSERIAL PRIMARY KEY,
    txid TEXT NOT NULL REFERENCES transactions(txid),
    vin_index INT NOT NULL,
    previous_txid TEXT NOT NULL,
    previous_vout INT NOT NULL,
    script_sig TEXT,
    sequence BIGINT NOT NULL,
    UNIQUE (txid, vin_index)
);

CREATE TABLE IF NOT EXISTS transaction_outputs (
    txid TEXT NOT NULL REFERENCES transactions(txid),
    vout INT NOT NULL,
    value_sats BIGINT NOT NULL,
    script_pub_key TEXT NOT NULL,
    address TEXT,
    spent BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (txid, vout)
);

CREATE TABLE IF NOT EXISTS sync_state (
    id BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK (id = TRUE),
    last_height BIGINT NOT NULL DEFAULT -1,
    last_hash TEXT
);

INSERT INTO sync_state (id, last_height)
VALUES (TRUE, -1)
ON CONFLICT DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_transactions_height
ON transactions(block_height);

CREATE INDEX IF NOT EXISTS idx_outputs_address
ON transaction_outputs(address);

CREATE INDEX IF NOT EXISTS idx_inputs_previous
ON transaction_inputs(previous_txid, previous_vout);
