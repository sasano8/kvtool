package filesystems

import "os"

// getTestMinIOEndpoint returns MinIO endpoint for testing
// Makefile の test-ci ターゲットで MINIO_ENDPOINT を設定する
func getTestMinIOEndpoint() string {
	if endpoint := os.Getenv("MINIO_ENDPOINT"); endpoint != "" {
		return endpoint
	}
	return "http://localhost:9000" // ローカル開発用デフォルト
}
