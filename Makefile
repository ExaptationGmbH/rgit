BINARY  := rgit
PREFIX  ?= /usr/local
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: all build install uninstall test vet fmt clean

all: build

build: ## Build the rgit binary into ./bin
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) .

install: build ## Install rgit into $(PREFIX)/bin (default /usr/local/bin)
	@mkdir -p "$(DESTDIR)$(PREFIX)/bin" 2>/dev/null || true
	@if install -m 0755 bin/$(BINARY) "$(DESTDIR)$(PREFIX)/bin/$(BINARY)" 2>/dev/null; then \
		echo "installed $(BINARY) -> $(DESTDIR)$(PREFIX)/bin/$(BINARY)"; \
	else \
		echo "rgit: cannot write to $(PREFIX)/bin (permission denied)."; \
		echo "  Try one of:"; \
		echo "    sudo make install"; \
		echo "    make install PREFIX=\$$HOME/.local   # then add ~/.local/bin to PATH"; \
		exit 1; \
	fi

uninstall:
	rm -f "$(DESTDIR)$(PREFIX)/bin/$(BINARY)"

test:
	go test ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

clean:
	rm -rf bin
