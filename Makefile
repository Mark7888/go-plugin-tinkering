.PHONY: build build-plugin build-all run swagger clean help

BINARY   := bin/app
PLUGIN   := plugins/sample.myext
SWAG     := $(shell go env GOPATH)/bin/swag

## build: Build the main application binary
build:
	go build -o $(BINARY) ./cmd/app/

## build-plugin: Build the sample plugin binary
build-plugin:
	go build -o $(PLUGIN) ./cmd/sample-plugin/

## build-all: Build both the application and the sample plugin
build-all: build build-plugin

## run: Build the sample plugin then start the application
run: build-plugin
	go run ./cmd/app/

## swagger: Generate swagger documentation
swagger:
	$(SWAG) init -g cmd/app/main.go -o docs

## clean: Remove built binaries and plugin files
clean:
	rm -rf bin/
	rm -f plugins/*.myext

## help: Show this help message
help:
	@echo "Usage: make <target>"
	@echo ""
	@grep -E '^## ' Makefile | sed 's/## /  /'
