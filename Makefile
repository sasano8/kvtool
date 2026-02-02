# Makefile
# kvtool のビルド・テスト・ドキュメント生成を管理します
BIN_DIR := bin
BINARY_NAME := kvtool
MAIN_GO_PKG := .

# @go build -ldflags "-X go.etcd.io/bbolt/version.Version=$(git describe --tags --dirty)"

# ビルド関連
# バイナリをビルドします（lint チェック付き）
.PHONY: build
build: lint
	@go build -o $(BIN_DIR)/$(BINARY_NAME) ${MAIN_GO_PKG}
	@chmod +x $(BIN_DIR)/$(BINARY_NAME)

# 複数プラットフォーム向けにビルドします（Linux, macOS, Windows）
build-full: lint
	@mkdir -p dist
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/${BINARY_NAME}-linux-amd64 ${MAIN_GO_PKG}
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o dist/${BINARY_NAME}-darwin-arm64 ${MAIN_GO_PKG}
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o dist/${BINARY_NAME}-windows-amd64.exe ${MAIN_GO_PKG}

# バイナリを /usr/local/bin にインストールします
.PHONY: install
install: build
	@mv $(BIN_DIR)/$(BINARY_NAME) /usr/local/bin/

# テスト関連
# 全パッケージのテストを実行します
.PHONY: test
test:
	@go test ./...

# CI 環境用のテスト（GitHub Actions services 対応）
.PHONY: test-ci
test-ci:
	MINIO_ENDPOINT=http://127.0.0.1:9000 \
	VAULT_ADDR=http://127.0.0.1:8200 \
	VAULT_TOKEN=root \
	go test -v ./...

# テストを詳細表示で実行します（主要な出力のみ）
.PHONY: test-v
test-v:
	@echo "=== Running tests (verbose) ==="
	@go test -v ./... 2>&1 | grep -E "^(===|---|\?|ok |FAIL)"

# カバレッジレポートを生成します（.cache/coverage.html）
.PHONY: test-coverage
test-coverage:
	@echo "=== Running tests with coverage ==="
	@mkdir -p .cache
	@go test -coverprofile=.cache/coverage.out ./...
	@go tool cover -html=.cache/coverage.out -o .cache/coverage.html
	@echo "Coverage report generated: .cache/coverage.html"

# Vault と MinIO を起動してから全テストを実行します（統合テスト含む）
# scripts/test-with-services.sh に処理を委譲（CI と共通化）
.PHONY: test-full
test-full: lint
	@./scripts/test-with-services.sh

# 開発用サービス（Vault + MinIO）を起動します
# Vault: http://localhost:8200 (token: root)
# MinIO: http://localhost:9000 (user: minioadmin)
.PHONY: services-up
services-up:
	@docker compose up -d vault minio
	@./scripts/wait-for-vault.sh
	@./scripts/wait-for-minio.sh
	@docker compose up vault-init minio-init
	@echo ""
	@echo "Services ready:"
	@echo "  Vault: http://localhost:8200 (token: root)"
	@echo "  MinIO: http://localhost:9000 (user: minioadmin)"

# 全サービスを停止します
.PHONY: services-down
services-down:
	@docker compose down --remove-orphans

# 全サービスのログを表示します
.PHONY: services-logs
services-logs:
	@docker compose logs -f

# コード品質関連
# コードをフォーマットします（go fmt）
.PHONY: format
format:
	@go fmt ./...

# 静的解析を実行します（go mod tidy + go vet）
.PHONY: lint
lint:
	@go mod tidy  # 未使用のモジュールなどの整理
	@go vet ./...  # 静的解析
# 	@test -z "$$(gofmt -l .)" || (echo "gofmt needed:" && gofmt -l . && exit 1)  # 差分チェック go fmt は修正してしまうので差分だけチェックしたい

# ドキュメント関連
# コード内のコメントと構造体タグから API リファレンスを自動生成します
# 出力: docs/api-reference.md
# 詳細: CLAUDE.md の「コード規約」セクション参照
.PHONY: gen-docs
gen-docs:
	@echo "=== Generating documentation from code ==="
	@go run scripts/gen-docs.go
	@echo "✅ Documentation generated successfully"

# ローカルで godoc サーバーを起動します（http://localhost:6060）
.PHONY: doc-serve
doc-serve:
	# go install golang.org/x/pkgsite/cmd/pkgsite@latest
	@~/go/bin/pkgsite

# mdBook でドキュメントをビルドします（出力: docs/book/）
.PHONY: book-build
book-build:
	@echo "=== Building mdBook ==="
	@mdbook build docs
	@echo "✅ Book built successfully"

# mdBook サーバーを起動します（http://localhost:3000）
.PHONY: book-serve
book-serve:
	@echo "=== Starting mdBook server ==="
	@mdbook serve docs --open

.PHONY: cobra-init
cobra-init:
	@~/go/bin/cobra-cli init
	@~/go/bin/cobra-cli add load
	@~/go/bin/cobra-cli add env -p loadCmd
