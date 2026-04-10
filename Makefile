.PHONY: tidy fmt test build run snapshots

tidy:
	go mod tidy

fmt:
	gofmt -w ./cmd ./internal

test:
	go test ./...

build:
	mkdir -p bin
	go build -o bin/server ./cmd/server
	go build -o bin/crawl ./cmd/crawl

run:
	go run ./cmd/server

snapshots:
	go run ./cmd/crawl --url https://cmlabs.co --out snapshots/cmlabs.co.html
	go run ./cmd/crawl --url https://sequence.day --out snapshots/sequence.day.html
	go run ./cmd/crawl --url https://example.com --out snapshots/example.com.html

