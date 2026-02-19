#!/bin/bash
set -e

# 統合テスト実行スクリプト
# Vault と MinIO を起動し、初期化してからテストを実行します
#
# 使い方:
#   ./scripts/test-with-services.sh          # 通常実行
#   ./scripts/test-with-services.sh --ci     # CI モード（詳細ログ）

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# CI モードかどうか
CI_MODE=false
if [ "$1" = "--ci" ] || [ "$CI" = "true" ]; then
    CI_MODE=true
fi

# テスト用環境変数
export VAULT_ADDR="${VAULT_ADDR:-http://127.0.0.1:8200}"
export VAULT_TOKEN="${VAULT_TOKEN:-root}"
export MINIO_ENDPOINT="${MINIO_ENDPOINT:-http://127.0.0.1:9000}"
export NATS_URL="${NATS_URL:-nats://127.0.0.1:4222}"
export REDIS_ADDR="${REDIS_ADDR:-127.0.0.1:6379}"

log() {
    echo "=== $1 ==="
}

cleanup() {
    log "Stopping services"
    docker compose -f "$PROJECT_DIR/docker-compose.yml" down --remove-orphans 2>/dev/null || true
}

# 終了時にクリーンアップ
trap cleanup EXIT

cd "$PROJECT_DIR"

log "Starting services (Vault + MinIO + NATS + Redis)"
docker compose up -d vault minio nats redis

log "Waiting for Vault"
"$SCRIPT_DIR/wait-for-vault.sh" "$VAULT_ADDR"

log "Initializing Vault"
docker compose up vault-init

log "Waiting for MinIO"
"$SCRIPT_DIR/wait-for-minio.sh" "$MINIO_ENDPOINT"

log "Initializing MinIO"
docker compose up minio-init

log "Waiting for NATS"
"$SCRIPT_DIR/wait-for-nats.sh" "http://127.0.0.1:8222"

log "Initializing NATS"
docker compose up nats-init

log "Waiting for Redis"
"$SCRIPT_DIR/wait-for-redis.sh"

log "Initializing Redis"
docker compose up redis-init

log "Running tests"
if [ "$CI_MODE" = true ]; then
    go test -v ./...
else
    go test ./...
fi

log "All tests passed!"
