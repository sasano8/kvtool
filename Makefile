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
.PHONY: test-full
test-full:
	@echo "=== Running lint ==="
	@go mod tidy
	@go vet ./...
	@echo ""
	@echo "=== Starting Vault and MinIO ==="
	@docker compose up -d vault minio
	@echo ""
	@echo "=== Waiting for Vault to be ready ==="
	@./scripts/wait-for-vault.sh
	@echo ""
	@echo "=== Initializing Vault ==="
	@docker compose up vault-init
	@echo ""
	@echo "=== Waiting for MinIO to be ready ==="
	@./scripts/wait-for-minio.sh
	@echo ""
	@echo "=== Initializing MinIO ==="
	@docker compose up minio-init
	@echo ""
	@echo "=== Running all tests (with Vault and MinIO) ==="
	@VAULT_ADDR=http://127.0.0.1:8200 VAULT_TOKEN=root MINIO_ENDPOINT=http://127.0.0.1:9000 go test -v ./...
	@echo ""
	@echo "=== Stopping services ==="
	@docker compose down
	@echo ""
	@echo "=== All tests passed! ==="

# Vault 統合テスト用
# 開発用 Vault を起動します（http://localhost:8200, token: root）
.PHONY: vault-up
vault-up:
	@echo "Starting Vault..."
	@docker compose up -d vault
	@echo "Waiting for Vault to be ready..."
	@sleep 2
	@docker compose up vault-init
	@echo "Vault is ready at http://localhost:8200"
	@echo "  Token: root"
	@echo "  UI: http://localhost:8200/ui"

# Vault を停止します
.PHONY: vault-down
vault-down:
	@echo "Stopping Vault..."
	@docker compose down

# Vault のログを表示します
.PHONY: vault-logs
vault-logs:
	@docker compose logs -f vault

# MinIO (S3 互換) 統合テスト用
# 開発用 MinIO を起動します（http://localhost:9000, console: http://localhost:9001）
.PHONY: minio-up
minio-up:
	@echo "Starting MinIO..."
	@docker compose up -d minio
	@echo "Waiting for MinIO to be ready..."
	@sleep 2
	@docker compose up minio-init
	@echo "MinIO is ready at http://localhost:9000"
	@echo "  Access Key: minioadmin"
	@echo "  Secret Key: minioadmin"
	@echo "  Web Console: http://localhost:9001"

# MinIO を停止します
.PHONY: minio-down
minio-down:
	@echo "Stopping MinIO..."
	@docker compose down minio

# MinIO のログを表示します
.PHONY: minio-logs
minio-logs:
	@docker compose logs -f minio

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
