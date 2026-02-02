#!/bin/bash
set -e

# MinIO ヘルスチェックスクリプト（curl ベース）
# 使い方: ./scripts/wait-for-minio.sh [endpoint]

ENDPOINT="${1:-http://127.0.0.1:9000}"
MAX_RETRIES=60
RETRY_INTERVAL=2

echo "=== MinIO ヘルスチェック ==="
echo "Endpoint: $ENDPOINT"
echo "Max retries: $MAX_RETRIES (total: $((MAX_RETRIES * RETRY_INTERVAL))s)"

# curl がインストールされているか確認
if ! command -v curl &> /dev/null; then
    echo "Error: curl is not installed"
    exit 1
fi

# MinIO が準備できるまで待つ
echo "Waiting for MinIO to be ready..."
for i in $(seq 1 $MAX_RETRIES); do
    # MinIO の liveness probe を使用
    if curl -sf "${ENDPOINT}/minio/health/live" >/dev/null 2>&1; then
        echo "✓ MinIO is ready! (attempt $i/$MAX_RETRIES)"
        exit 0
    fi

    # デバッグ: 最初の数回と最後の数回は詳細を表示
    if [ $i -le 3 ] || [ $i -ge $((MAX_RETRIES - 2)) ]; then
        echo "Retrying... ($i/$MAX_RETRIES) - curl exit code: $?"
        curl -v "${ENDPOINT}/minio/health/live" 2>&1 | head -5 || true
    else
        echo "Retrying... ($i/$MAX_RETRIES)"
    fi

    sleep $RETRY_INTERVAL
done

echo "✗ MinIO is not ready after $MAX_RETRIES attempts"
exit 1
