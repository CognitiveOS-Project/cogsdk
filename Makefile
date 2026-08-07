SHELL := /bin/sh
.SHELLFLAGS := -eu -c
.ONESHELL:
.DELETE_ON_ERROR:

BUILD_DIR := build
GO := go

.PHONY: build test lint clean

build:
	$(GO) build ./...

test:
	$(GO) test ./... -count=1

lint:
	$(GO) vet ./...

clean:
	rm -rf $(BUILD_DIR)
