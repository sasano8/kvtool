#!/bin/bash
set -e

# Redis ヘルスチェックスクリプト（redis-cli 不要版）
# TCP接続 + PING コマンドで確認
# 使い方: ./scripts/wait-for-redis.sh [host] [port]

HOST="${1:-127.0.0.1}"
PORT="${2:-6379}"
MAX_RETRIES=30
RETRY_INTERVAL=1

echo "=== Redis ヘルスチェック ==="
echo "Host: $HOST:$PORT"
echo "Max retries: $MAX_RETRIES (total: ${MAX_RETRIES}s)"

echo "Waiting for Redis to be ready..."
for i in $(seq 1 $MAX_RETRIES); do
    # bash の /dev/tcp を使って PING コマンドを送信し PONG を確認
    if (echo -e "PING\r\n"; sleep 0.1) | nc -w 1 "$HOST" "$PORT" 2>/dev/null | grep -q PONG; then
        echo "✓ Redis is ready! (attempt $i/$MAX_RETRIES)"
        exit 0
    fi

    echo "Retrying... ($i/$MAX_RETRIES)"
    sleep $RETRY_INTERVAL
done

echo "✗ Redis is not ready after $MAX_RETRIES attempts"
exit 1
