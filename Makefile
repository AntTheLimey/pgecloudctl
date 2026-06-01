BINARY := pgecloudctl
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X github.com/AntTheLimey/pgecloudctl/internal/cmd.Version=$(VERSION)"

.PHONY: build test lint generate clean

build:
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/pgecloudctl

test:
	go test ./... -race -coverprofile=coverage.out

lint:
	golangci-lint run

generate:
	./scripts/generate.sh

clean:
	rm -rf bin/ coverage.out
