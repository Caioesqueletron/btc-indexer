package bitcoin

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
)

type Client struct {
    url, user, password string
    http *http.Client
}

func NewClient(url, user, password string) *Client {
    return &Client{url: url, user: user, password: password, http: &http.Client{}}
}

type rpcRequest struct {
    JSONRPC string `json:"jsonrpc"`
    ID      int    `json:"id"`
    Method  string `json:"method"`
    Params  []any  `json:"params"`
}

type rpcResponse struct {
    Result json.RawMessage `json:"result"`
    Error  *struct {
        Code    int    `json:"code"`
        Message string `json:"message"`
    } `json:"error"`
}

func (c *Client) Call(ctx context.Context, method string, params ...any) (json.RawMessage, error) {
    payload, err := json.Marshal(rpcRequest{JSONRPC: "1.0", ID: 1, Method: method, Params: params})
    if err != nil {
        return nil, err
    }

    req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(payload))
    if err != nil {
        return nil, err
    }

    req.SetBasicAuth(c.user, c.password)
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.http.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return nil, fmt.Errorf("bitcoin rpc http status: %s", resp.Status)
    }

    var out rpcResponse
    if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
        return nil, err
    }
    if out.Error != nil {
        return nil, fmt.Errorf("bitcoin rpc error %d: %s", out.Error.Code, out.Error.Message)
    }

    return out.Result, nil
}

func (c *Client) GetBlockCount(ctx context.Context) (int64, error) {
    raw, err := c.Call(ctx, "getblockcount")
    if err != nil {
        return 0, err
    }
    var height int64
    err = json.Unmarshal(raw, &height)
    return height, err
}

func (c *Client) GetBlockHash(ctx context.Context, height int64) (string, error) {
    raw, err := c.Call(ctx, "getblockhash", height)
    if err != nil {
        return "", err
    }
    var hash string
    err = json.Unmarshal(raw, &hash)
    return hash, err
}

func (c *Client) GetBlock(ctx context.Context, hash string) (Block, error) {
    raw, err := c.Call(ctx, "getblock", hash, 2)
    if err != nil {
        return Block{}, err
    }
    var block Block
    err = json.Unmarshal(raw, &block)
    return block, err
}

type Block struct {
    Hash         string        `json:"hash"`
    Height       int64         `json:"height"`
    PreviousHash string        `json:"previousblockhash"`
    MerkleRoot   string        `json:"merkleroot"`
    Time         int64         `json:"time"`
    Nonce        uint64        `json:"nonce"`
    Bits         string        `json:"bits"`
    Size         int64         `json:"size"`
    Tx           []Transaction `json:"tx"`
}

type Transaction struct {
    TxID string `json:"txid"`
    Hex  string `json:"hex"`
    Vin  []Vin  `json:"vin"`
    Vout []Vout `json:"vout"`
}

type Vin struct {
    TxID string `json:"txid"`
    Vout int    `json:"vout"`
    ScriptSig struct {
        Hex string `json:"hex"`
    } `json:"scriptSig"`
    Sequence uint64 `json:"sequence"`
}

type Vout struct {
    Value float64 `json:"value"`
    N     int     `json:"n"`
    ScriptPubKey struct {
        Hex       string   `json:"hex"`
        Address   string   `json:"address"`
        Addresses []string `json:"addresses"`
    } `json:"scriptPubKey"`
}

func SatsFromBTC(btc float64) int64 {
    return int64(btc * 100_000_000)
}
