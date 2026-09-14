PROTO=proto/blockchain.proto

.PHONY: proto tidy test indexer api infra

proto:
	protoc --go_out=. --go_opt=module=github.com/caio/btc-indexer \
		--go-grpc_out=. --go-grpc_opt=module=github.com/caio/btc-indexer \
		$(PROTO)

tidy:
	go mod tidy

test:
	go test ./...

indexer:
	go run ./cmd/indexer

api:
	go run ./cmd/api

infra:
	docker compose up -d
