.PHONY: build test vet lint check run clean

BIN := bin/terminal-portfolio

# Ensure ~/go/bin is on PATH for golangci-lint
export PATH := $(HOME)/go/bin:$(PATH)

build:
	go build -o $(BIN) ./cmd/server

test:
	go test ./... -race

vet:
	go vet ./...

lint:
	golangci-lint run

check: vet lint test

run: build
	./$(BIN)

clean:
	rm -rf bin/
