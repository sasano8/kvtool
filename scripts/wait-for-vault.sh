#!/bin/bash
set -e

# Vault ヘルスチェックスクリプト（curl ベース）
# 使い方: ./scripts/wait-for-vault.sh [addr]

VAULT_ADDR="${1:-http://127.0.0.1:8200}"
MAX_RETRIES=30
RETRY_INTERVAL=1

echo "=== Vault ヘルスチェック ==="
echo "Address: $VAULT_ADDR"
echo "Max retries: $MAX_RETRIES (total: ${MAX_RETRIES}s)"

# curl がインストールされているか確認
if ! command -v curl &> /dev/null; then
    echo "Error: curl is not installed"
    exit 1
fi

# Vault が準備できるまで待つ
echo "Waiting for Vault to be ready..."
for i in $(seq 1 $MAX_RETRIES); do
    # Vault の health API を使用（/v1/sys/health）
    # 200 または 429 (standby) が返ってくれば起動している
    if curl -sf "${VAULT_ADDR}/v1/sys/health" >/dev/null 2>&1; then
        echo "✓ Vault is ready! (attempt $i/$MAX_RETRIES)"
        exit 0
    fi
    echo "Retrying... ($i/$MAX_RETRIES)"
    sleep $RETRY_INTERVAL
done

echo "✗ Vault is not ready after $MAX_RETRIES attempts"
exit 1
