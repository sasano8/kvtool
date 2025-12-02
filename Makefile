# Makefile
BIN_DIR := bin
BINARY_NAME := kvtool
MAIN_GO_PKG := ./cmd/kvtool

.PHONY: build
build:
	@go build -o $(BIN_DIR)/$(BINARY_NAME) ${MAIN_GO_PKG}
	@chmod +x $(BIN_DIR)/$(BINARY_NAME)

build-release:
	@mkdir -p dist
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/${BINARY_NAME}-linux-amd64 ${MAIN_GO_PKG}
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o dist/${BINARY_NAME}-darwin-arm64 ${MAIN_GO_PKG}
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o dist/${BINARY_NAME}-windows-amd64.exe ${MAIN_GO_PKG}

.PHONY: install
install: build
	@mv $(BIN_DIR)/$(BINARY_NAME) /usr/local/bin/

.PHONY: test
test:
	@go test ./...

.PHONY: format
format:
	@go fmt ./...
