package grpcapi

import (
    "context"

    blockchain "github.com/caio/btc-indexer/gen/blockchain"
    "github.com/caio/btc-indexer/internal/storage"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

type Server struct {
    blockchain.UnimplementedBlockchainServiceServer
    store *storage.Store
}

func NewServer(store *storage.Store) *Server {
    return &Server{store: store}
}

func (s *Server) GetBlockByHeight(ctx context.Context, req *blockchain.GetBlockByHeightRequest) (*blockchain.GetBlockResponse, error) {
    hash, err := s.store.GetBlockByHeight(ctx, req.GetHeight())
    if err != nil {
        return nil, status.Error(codes.NotFound, "block not found")
    }

    return &blockchain.GetBlockResponse{
        Hash: hash,
        Height: req.GetHeight(),
    }, nil
}

func (s *Server) GetTransaction(ctx context.Context, req *blockchain.GetTransactionRequest) (*blockchain.GetTransactionResponse, error) {
    blockHash, height, err := s.store.GetTransaction(ctx, req.GetTxid())
    if err != nil {
        return nil, status.Error(codes.NotFound, "transaction not found")
    }

    return &blockchain.GetTransactionResponse{
        Txid:       req.GetTxid(),
        BlockHash:  blockHash,
        BlockHeight: height,
    }, nil
}
