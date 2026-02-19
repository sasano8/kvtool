package filesystems

import "os"

// getTestMinIOEndpoint returns MinIO endpoint for testing
// Makefile の test-ci ターゲットで MINIO_ENDPOINT を設定する
func getTestMinIOEndpoint() string {
	if endpoint := os.Getenv("MINIO_ENDPOINT"); endpoint != "" {
		return endpoint
	}
	return "http://127.0.0.1:9000" // ローカル開発用デフォルト
}

// getTestVaultAddr returns Vault address for testing
// Makefile の test-ci ターゲットで VAULT_ADDR を設定する
func getTestVaultAddr() string {
	if addr := os.Getenv("VAULT_ADDR"); addr != "" {
		return addr
	}
	return "http://127.0.0.1:8200" // ローカル開発用デフォルト
}

// getTestVaultToken returns Vault token for testing
// Makefile の test-ci ターゲットで VAULT_TOKEN を設定する
func getTestVaultToken() string {
	if token := os.Getenv("VAULT_TOKEN"); token != "" {
		return token
	}
	return "root" // ローカル開発用デフォルト（dev mode）
}

// getTestNatsURL returns NATS URL for testing
// Makefile の test-ci ターゲットで NATS_URL を設定する
func getTestNatsURL() string {
	if url := os.Getenv("NATS_URL"); url != "" {
		return url
	}
	return "nats://127.0.0.1:4222" // ローカル開発用デフォルト
}

// getTestRedisAddr returns Redis address for testing
// Makefile の test-ci ターゲットで REDIS_ADDR を設定する
func getTestRedisAddr() string {
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		return addr
	}
	return "127.0.0.1:6379" // ローカル開発用デフォルト
}
