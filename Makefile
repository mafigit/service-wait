BINARY := service-wait
PKG    := github.com/mafigit/service-wait/cmd/service-wait
BIN    := bin/$(BINARY)
PREFIX ?= $(HOME)/.local

GO         ?= go
GOFLAGS    ?=
VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || printf dev)
COMMIT     ?= $(shell git rev-parse --short HEAD 2>/dev/null || printf none)
DATE       ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || printf unknown)
LDFLAGS    ?= -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

.PHONY: all build install uninstall run update-presets fmt vet tidy test clean help

all: build

build:
	@mkdir -p bin
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN) $(PKG)

run: build
	./$(BIN) $(ARGS)

install: build
	@mkdir -p $(PREFIX)/bin
	install -m 0755 $(BIN) $(PREFIX)/bin/$(BINARY)

uninstall:
	rm -f $(PREFIX)/bin/$(BINARY)

test:
	$(GO) test -v -timeout 120s ./...

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

staticcheck:
	$(GO) run honnef.co/go/tools/cmd/staticcheck@latest ./...

tidy:
	$(GO) mod tidy
