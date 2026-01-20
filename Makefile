# Makefile
BIN_DIR := bin
BINARY_NAME := kvtool
MAIN_GO_PKG := .

# @go build -ldflags "-X go.etcd.io/bbolt/version.Version=$(git describe --tags --dirty)"

.PHONY: build
build: lint
	@go build -o $(BIN_DIR)/$(BINARY_NAME) ${MAIN_GO_PKG}
	@chmod +x $(BIN_DIR)/$(BINARY_NAME)

build-full: lint
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

.PHONY: lint
lint:
	@go mod tidy  # 未使用のモジュールなどの整理
	@go vet ./...  # 静的解析
# 	@test -z "$$(gofmt -l .)" || (echo "gofmt needed:" && gofmt -l . && exit 1)  # 差分チェック go fmt は修正してしまうので差分だけチェックしたい

.PHONY: doc-serve
doc-serve:
	# go install golang.org/x/pkgsite/cmd/pkgsite@latest
	@~/go/bin/pkgsite

.PHONY: cobra-init
cobra-init:
	@~/go/bin/cobra-cli init
	@~/go/bin/cobra-cli add load
	@~/go/bin/cobra-cli add env -p loadCmd
