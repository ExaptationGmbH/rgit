BINARY  := rgit
PREFIX  ?= /usr/local
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: all build install uninstall test vet fmt clean

all: build

build: ## Build the rgit binary into ./bin
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) .

install: ## Install rgit into $(PREFIX)/bin (default /usr/local/bin)
	go build -ldflags "$(LDFLAGS)" -o $(PREFIX)/bin/$(BINARY) .

uninstall:
	rm -f $(PREFIX)/bin/$(BINARY)

test:
	go test ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

clean:
	rm -rf bin
