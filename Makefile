# Makefile
BIN_DIR := bin
BINARY_NAME := kvtool

.PHONY: build
build:
	@go build -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/kvtool
	@chmod +x $(BIN_DIR)/$(BINARY_NAME)

.PHONY: install
install:
	@mv $(BIN_DIR)/$(BINARY_NAME) /usr/local/bin/	

.PHONY: test
test:
	@go test ./...
