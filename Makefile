# Makefile
BIN_DIR := bin
BINARY_NAME := kvtool

.PHONY: build
build:
	@go build -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/kvtool
	@chmod +x $(BIN_DIR)/$(BINARY_NAME)

.PHONY: install
install: build
	@mv $(BIN_DIR)/$(BINARY_NAME) /usr/local/bin/

.PHONY: test
test:
	@go test ./...

.PHONY: format
format:
	@go fmt ./...
