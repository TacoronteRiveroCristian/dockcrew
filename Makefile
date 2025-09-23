BINARY := dockcrew
VERSION ?= 0.0.0-dev
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo nogit)
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X main.version=$(VERSION)+$(COMMIT) -s -w

.PHONY: build run clean test

build:
	GOFLAGS="-trimpath" CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) ./cmd/dockcrew

run: build
	./bin/$(BINARY) --version

clean:
	rm -rf bin/ dist/ coverage.*

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test ./...
