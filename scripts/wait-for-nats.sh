#!/bin/bash
set -e

# NATS ヘルスチェックスクリプト（curl ベース）
# 使い方: ./scripts/wait-for-nats.sh [monitoring_endpoint]

ENDPOINT="${1:-http://127.0.0.1:8222}"
MAX_RETRIES=30
RETRY_INTERVAL=1

echo "=== NATS ヘルスチェック ==="
echo "Monitoring endpoint: $ENDPOINT"
echo "Max retries: $MAX_RETRIES (total: ${MAX_RETRIES}s)"

if ! command -v curl &> /dev/null; then
    echo "Error: curl is not installed"
    exit 1
fi

echo "Waiting for NATS to be ready..."
for i in $(seq 1 $MAX_RETRIES); do
    if curl -sf "${ENDPOINT}/healthz" >/dev/null 2>&1; then
        echo "✓ NATS is ready! (attempt $i/$MAX_RETRIES)"
        exit 0
    fi

    echo "Retrying... ($i/$MAX_RETRIES)"
    sleep $RETRY_INTERVAL
done

echo "✗ NATS is not ready after $MAX_RETRIES attempts"
exit 1
