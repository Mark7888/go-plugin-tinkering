.PHONY: build build-plugin build-counter-plugin build-all run swagger clean help

BINARY          := bin/app
PLUGIN          := plugins/sample.myext
COUNTER_PLUGIN  := plugins/counter.myext
SWAG            := $(shell go env GOPATH)/bin/swag

## build: Build the main application binary
build:
	go build -o $(BINARY) ./cmd/app/

## build-plugin: Build the sample plugin binary
build-plugin:
	go build -o $(PLUGIN) ./cmd/sample-plugin/

## build-counter-plugin: Build the counter plugin binary
build-counter-plugin:
	go build -o $(COUNTER_PLUGIN) ./cmd/counter-plugin/

## build-all: Build the application and all plugin binaries
build-all: build build-plugin build-counter-plugin

## run: Build all plugins then start the application
run: build-plugin build-counter-plugin
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
