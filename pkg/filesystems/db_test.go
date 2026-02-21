package filesystems

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestDbFs(t *testing.T) {
	// テスト用の一時ファイルを作成
	tmpFile, err := os.CreateTemp("", "kvtool-db-test-*.db")
	require.NoError(t, err)
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	// テスト用データベースをセットアップ
	db, err := sql.Open("sqlite", tmpFile.Name())
	require.NoError(t, err)
	defer db.Close()

	// テーブル作成
	_, err = db.Exec(`
		CREATE TABLE config (
			key TEXT NOT NULL,
			namespace TEXT NOT NULL,
			value TEXT NOT NULL,
			PRIMARY KEY (key, namespace)
		)
	`)
	require.NoError(t, err)

	// テストデータ挿入
	_, err = db.Exec(`INSERT INTO config (key, namespace, value) VALUES (?, ?, ?)`,
		"app/settings", "default", `{"debug": true, "port": 8080}`)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO config (key, namespace, value) VALUES (?, ?, ?)`,
		"app/settings", "production", `{"debug": false, "port": 80}`)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO config (key, namespace, value) VALUES (?, ?, ?)`,
		"simple", "default", `hello world`)
	require.NoError(t, err)

	db.Close()

	tests := []struct {
		name        string
		query       string
		namespace   string
		key         string
		expected    any
		expectError bool
	}{
		{
			name:      "JSON値の取得",
			query:     "SELECT value FROM config WHERE key = {key} AND namespace = {namespace}",
			namespace: "default",
			key:       "app/settings",
			expected: map[string]any{
				"debug": true,
				"port":  float64(8080),
			},
		},
		{
			name:      "production namespace",
			query:     "SELECT value FROM config WHERE key = {key} AND namespace = {namespace}",
			namespace: "production",
			key:       "app/settings",
			expected: map[string]any{
				"debug": false,
				"port":  float64(80),
			},
		},
		{
			name:      "文字列値の取得",
			query:     "SELECT value FROM config WHERE key = {key} AND namespace = {namespace}",
			namespace: "default",
			key:       "simple",
			expected:  "hello world",
		},
		{
			name:        "存在しないキー",
			query:       "SELECT value FROM config WHERE key = {key} AND namespace = {namespace}",
			namespace:   "default",
			key:         "nonexistent",
			expectError: true,
		},
		{
			name:      "key のみのクエリ",
			query:     "SELECT value FROM config WHERE key = {key} LIMIT 1",
			namespace: "default",
			key:       "app/settings",
			expected: map[string]any{
				"debug": true,
				"port":  float64(8080),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			config := &DbFsConfig{
				ConnectionString: tmpFile.Name(),
				Driver:           "sqlite",
				Query:            tt.query,
				Namespace:        tt.namespace,
			}

			fs, err := NewDbFs(ctx, config)
			require.NoError(err)
			defer fs.Close()

			file, err := fs.GetFile(tt.key)
			require.NoError(err)

			result, err := file.LoadAsJson()
			if tt.expectError {
				require.Error(err)
				return
			}
			require.NoError(err)
			require.Equal(tt.expected, result)
		})
	}
}

func TestDbFs_DetectDriver(t *testing.T) {
	tests := []struct {
		connStr  string
		expected string
	}{
		{"postgres://user:pass@localhost/db", "postgres"},
		{"postgresql://user:pass@localhost/db", "postgres"},
		{"user:pass@tcp(localhost:3306)/db", "mysql"},
		{"file:test.db", "sqlite"},
		{"test.db", "sqlite"},
		{"unknown://something", ""},
	}

	for _, tt := range tests {
		t.Run(tt.connStr, func(t *testing.T) {
			result := detectDriver(tt.connStr)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestDbFs_ConvertPlaceholders(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		driver   string
		hasKey   bool
		hasNs    bool
		expected string
	}{
		{
			name:     "PostgreSQL with key and namespace",
			query:    "SELECT value FROM config WHERE key = {key} AND namespace = {namespace}",
			driver:   "postgres",
			hasKey:   true,
			hasNs:    true,
			expected: "SELECT value FROM config WHERE key = $1 AND namespace = $2",
		},
		{
			name:     "MySQL with key and namespace",
			query:    "SELECT value FROM config WHERE key = {key} AND namespace = {namespace}",
			driver:   "mysql",
			hasKey:   true,
			hasNs:    true,
			expected: "SELECT value FROM config WHERE key = ? AND namespace = ?",
		},
		{
			name:     "SQLite with key only",
			query:    "SELECT value FROM config WHERE key = {key}",
			driver:   "sqlite",
			hasKey:   true,
			hasNs:    false,
			expected: "SELECT value FROM config WHERE key = ?",
		},
		{
			name:     "PostgreSQL with namespace only",
			query:    "SELECT value FROM config WHERE namespace = {namespace}",
			driver:   "postgres",
			hasKey:   false,
			hasNs:    true,
			expected: "SELECT value FROM config WHERE namespace = $1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertPlaceholders(tt.query, tt.driver, tt.hasKey, tt.hasNs)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestDbFs_ConfigValidation(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		config      *DbFsConfig
		expectError string
	}{
		{
			name: "接続文字列なし",
			config: &DbFsConfig{
				Query: "SELECT value FROM config WHERE key = {key}",
			},
			expectError: "connection_string is required",
		},
		{
			name: "クエリなし",
			config: &DbFsConfig{
				ConnectionString: "file:test.db",
			},
			expectError: "query is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			_, err := NewDbFs(ctx, tt.config)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectError)
		})
	}
}
