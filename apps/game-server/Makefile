.PHONY: test build run fmt lint

test:
	go test ./... -v -race -count=1

build:
	go build -o bin/server cmd/server/main.go

run:
	go run cmd/server/main.go

fmt:
	go fmt ./...

lint:
	golangci-lint run ./...
