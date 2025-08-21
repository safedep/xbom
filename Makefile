SHELL := /bin/bash
GITCOMMIT := $(shell git rev-parse HEAD)
VERSION := "$(shell git describe --tags --abbrev=0)-$(shell git rev-parse --short HEAD)"

BIN_DIR := bin
BIN := $(BIN_DIR)/xbom

GO_CFLAGS=-X 'github.com/safedep/xbom/internal/version.Commit=$(GITCOMMIT)' -X 'github.com/safedep/internal/version.Version=$(VERSION)'
GO_LDFLAGS=-ldflags "-w $(GO_CFLAGS)"

all: create_bin xbom

.PHONY: create_bin
create_bin:
	mkdir -p $(BIN_DIR)

.PHONY: xbom
xbom:
	go build ${GO_LDFLAGS} -o $(BIN)

.PHONY: test
test:
	go test ./...
