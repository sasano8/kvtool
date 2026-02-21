package filesystems

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

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

// getTestSkipGit returns whether git tests should be skipped
func getTestSkipGit() bool {
	return os.Getenv("SKIP_GIT_TESTS") == "true"
}

// setupTestGitRepo はテスト用のローカル bare リポジトリを作成する
// config.json を含む main ブランチを持つ
func setupTestGitRepo(t *testing.T) (string, func()) {
	t.Helper()

	tmpDir := t.TempDir()
	bareRepo := filepath.Join(tmpDir, "repo.git")
	workDir := filepath.Join(tmpDir, "work")

	// bare リポジトリを作成
	run(t, "git", "init", "--bare", bareRepo)

	// 作業ディレクトリを作成して clone
	run(t, "git", "clone", bareRepo, workDir)

	// テストデータを作成
	configPath := filepath.Join(workDir, "config.json")
	if err := os.WriteFile(configPath, []byte(`{"app": "test", "version": "1.0"}`), 0644); err != nil {
		t.Fatal(err)
	}

	// commit & push
	run(t, "git", "-C", workDir, "add", ".")
	run(t, "git", "-C", workDir, "-c", "user.email=test@test.com", "-c", "user.name=test", "commit", "-m", "initial")
	run(t, "git", "-C", workDir, "branch", "-M", "main")
	run(t, "git", "-C", workDir, "push", "-u", "origin", "main")

	cleanup := func() {
		// t.TempDir() が自動でクリーンアップする
	}

	return bareRepo, cleanup
}

func run(t *testing.T, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v failed: %s: %v", name, args, string(output), err)
	}
}
