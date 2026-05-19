.PHONY: test build run proto

test:
	go test ./... -v -race -count=1

build:
	go build -o bin/bot-service cmd/bot-service/main.go

run:
	go run cmd/bot-service/main.go

proto:
	protoc --go_out=. --go-grpc_out=. proto/bot.proto
