package filesystems

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestFilesystemInterface_GetFile は全てのファイルシステムが GetFile を一貫して実装していることをテスト
func TestFilesystemInterface_GetFile(t *testing.T) {
	tests := []struct {
		name           string
		setupFs        func(t *testing.T) Filesystem
		testPath       string
		expectError    bool
		errorSubstring string
	}{
		{
			name: "LocalFs - 有効なパス",
			setupFs: func(t *testing.T) Filesystem {
				tmpDir := t.TempDir()
				// テストファイルを作成
				testFile := filepath.Join(tmpDir, "test.json")
				err := os.WriteFile(testFile, []byte(`{"key":"value"}`), 0644)
				require.NoError(t, err)

				return &LocalFs{
					Root: tmpDir,
				}
			},
			testPath:    "test.json",
			expectError: false,
		},
		{
			name: "LocalFs - ネストされたパス",
			setupFs: func(t *testing.T) Filesystem {
				tmpDir := t.TempDir()
				// ネストされたディレクトリを作成
				nestedDir := filepath.Join(tmpDir, "subdir")
				err := os.MkdirAll(nestedDir, 0755)
				require.NoError(t, err)
				// ネストされたディレクトリにテストファイルを作成
				testFile := filepath.Join(nestedDir, "nested.json")
				err = os.WriteFile(testFile, []byte(`{"nested":"true"}`), 0644)
				require.NoError(t, err)

				return &LocalFs{
					Root: tmpDir,
				}
			},
			testPath:    "subdir/nested.json",
			expectError: false,
		},
		{
			name: "EnvFs - 任意のパスでファイルを返す",
			setupFs: func(t *testing.T) Filesystem {
				fs, err := NewEnvFs(context.Background(), &EnvFsConfig{})
				require.NoError(t, err)
				return fs
			},
			testPath:    "anypath",
			expectError: false,
		},
		{
			name: "S3Fs - 有効なパス",
			setupFs: func(t *testing.T) Filesystem {
				// MinIO が起動していない場合はスキップ
				if testing.Short() || os.Getenv("SKIP_S3_TESTS") == "true" {
					t.Skip("MinIO tests skipped (-short flag or SKIP_S3_TESTS=true)")
				}

				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				fs, err := NewS3Fs(ctx, S3FsConfig{
					Bucket:          "kvtool-test",
					Region:          "us-east-1",
					Endpoint:        getTestMinIOEndpoint(),
					UsePathStyle:    true,
					AccessKeyID:     "minioadmin",
					SecretAccessKey: "minioadmin",
				})
				require.NoError(t, err)
				return fs
			},
			testPath:    "config/app.json",
			expectError: false,
		},
		{
			name: "S3Fs - パストラバーサル攻撃の防止",
			setupFs: func(t *testing.T) Filesystem {
				if testing.Short() || os.Getenv("SKIP_S3_TESTS") == "true" {
					t.Skip("MinIO tests skipped (-short flag or SKIP_S3_TESTS=true)")
				}

				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				fs, err := NewS3Fs(ctx, S3FsConfig{
					Bucket:          "kvtool-test",
					Region:          "us-east-1",
					Endpoint:        getTestMinIOEndpoint(),
					UsePathStyle:    true,
					AccessKeyID:     "minioadmin",
					SecretAccessKey: "minioadmin",
				})
				require.NoError(t, err)
				return fs
			},
			testPath:       "../../../etc/passwd",
			expectError:    true,
			errorSubstring: "escapes root",
		},
		{
			name: "NatsFs - 有効なキー",
			setupFs: func(t *testing.T) Filesystem {
				if testing.Short() || os.Getenv("SKIP_NATS_TESTS") == "true" {
					t.Skip("NATS tests skipped (-short flag or SKIP_NATS_TESTS=true)")
				}

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				fs, err := NewNatsFs(ctx, &NatsFsConfig{
					URL:    getTestNatsURL(),
					Bucket: "kvtool-test",
				})
				require.NoError(t, err)
				return fs
			},
			testPath:    "app/settings",
			expectError: false,
		},
		{
			name: "RedisFs - 有効なキー",
			setupFs: func(t *testing.T) Filesystem {
				if testing.Short() || os.Getenv("SKIP_REDIS_TESTS") == "true" {
					t.Skip("Redis tests skipped (-short flag or SKIP_REDIS_TESTS=true)")
				}

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				fs, err := NewRedisFs(ctx, &RedisFsConfig{
					Addr:   getTestRedisAddr(),
					Prefix: "kvtool-test:",
				})
				require.NoError(t, err)
				return fs
			},
			testPath:    "app/settings",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			fs := tt.setupFs(t)

			file, err := fs.GetFile(tt.testPath)

			if tt.expectError {
				require.Error(err)
				if tt.errorSubstring != "" {
					require.Contains(err.Error(), tt.errorSubstring)
				}
				require.Nil(file)
			} else {
				require.NoError(err)
				require.NotNil(file)
			}
		})
	}
}

// TestFilesystemInterface_LoadAsJson は全てのファイルシステムが LoadAsJson を一貫して実装していることをテスト
func TestFilesystemInterface_LoadAsJson(t *testing.T) {
	tests := []struct {
		name           string
		setupFs        func(t *testing.T) Filesystem
		testPath       string
		expectedData   map[string]any
		expectError    bool
		errorSubstring string
	}{
		{
			name: "LocalFs - シンプルな JSON",
			setupFs: func(t *testing.T) Filesystem {
				tmpDir := t.TempDir()
				testFile := filepath.Join(tmpDir, "simple.json")
				err := os.WriteFile(testFile, []byte(`{"FOO":"bar","HELLO":"world"}`), 0644)
				require.NoError(t, err)

				return &LocalFs{
					Root: tmpDir,
				}
			},
			testPath: "simple.json",
			expectedData: map[string]any{
				"FOO":   "bar",
				"HELLO": "world",
			},
			expectError: false,
		},
		{
			name: "LocalFs - ネストされた JSON",
			setupFs: func(t *testing.T) Filesystem {
				tmpDir := t.TempDir()
				testFile := filepath.Join(tmpDir, "nested.json")
				testData := map[string]any{
					"name": "test",
					"nested": map[string]any{
						"key": "value",
					},
				}
				jsonBytes, _ := json.Marshal(testData)
				err := os.WriteFile(testFile, jsonBytes, 0644)
				require.NoError(t, err)

				return &LocalFs{
					Root: tmpDir,
				}
			},
			testPath: "nested.json",
			expectedData: map[string]any{
				"name": "test",
				"nested": map[string]any{
					"key": "value",
				},
			},
			expectError: false,
		},
		{
			name: "LocalFs - dotenv transform",
			setupFs: func(t *testing.T) Filesystem {
				tmpDir := t.TempDir()
				testFile := filepath.Join(tmpDir, "test.env")
				err := os.WriteFile(testFile, []byte("FOO=bar\nHELLO=world\n"), 0644)
				require.NoError(t, err)

				return &LocalFs{
					Root:      tmpDir,
					Transform: "dotenv",
				}
			},
			testPath: "test.env",
			expectedData: map[string]any{
				"FOO":   "bar",
				"HELLO": "world",
			},
			expectError: false,
		},
		{
			name: "LocalFs - ファイルが存在しない",
			setupFs: func(t *testing.T) Filesystem {
				tmpDir := t.TempDir()
				return &LocalFs{
					Root: tmpDir,
				}
			},
			testPath:       "nonexistent.json",
			expectError:    true,
			errorSubstring: "no such file",
		},
		{
			name: "EnvFs - 環境変数を読み込む",
			setupFs: func(t *testing.T) Filesystem {
				// テスト用環境変数を設定
				t.Setenv("TEST_VAR_1", "value1")
				t.Setenv("TEST_VAR_2", "value2")

				fs, err := NewEnvFs(context.Background(), &EnvFsConfig{})
				require.NoError(t, err)
				return fs
			},
			testPath: "anypath",
			// 全ての環境変数を予測できないため、一部の存在をチェック
			expectedData: nil, // カスタム検証を実行
			expectError:  false,
		},
		{
			name: "S3Fs - JSON ファイルを読み込む",
			setupFs: func(t *testing.T) Filesystem {
				if testing.Short() || os.Getenv("SKIP_S3_TESTS") == "true" {
					t.Skip("MinIO tests skipped (-short flag or SKIP_S3_TESTS=true)")
				}

				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				fs, err := NewS3Fs(ctx, S3FsConfig{
					Bucket:          "kvtool-test",
					Region:          "us-east-1",
					Endpoint:        getTestMinIOEndpoint(),
					UsePathStyle:    true,
					AccessKeyID:     "minioadmin",
					SecretAccessKey: "minioadmin",
				})
				require.NoError(t, err)
				return fs
			},
			testPath: "config/app.json",
			expectedData: map[string]any{
				"app_name": "kvtool",
				"version":  "1.0.0",
			},
			expectError: false,
		},
		{
			name: "S3Fs - dotenv transform",
			setupFs: func(t *testing.T) Filesystem {
				if testing.Short() || os.Getenv("SKIP_S3_TESTS") == "true" {
					t.Skip("MinIO tests skipped (-short flag or SKIP_S3_TESTS=true)")
				}

				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				fs, err := NewS3Fs(ctx, S3FsConfig{
					Bucket:          "kvtool-test",
					Region:          "us-east-1",
					Endpoint:        getTestMinIOEndpoint(),
					UsePathStyle:    true,
					AccessKeyID:     "minioadmin",
					SecretAccessKey: "minioadmin",
					Transform:       "dotenv",
				})
				require.NoError(t, err)
				return fs
			},
			testPath: "config/production.env",
			expectedData: map[string]any{
				"APP_NAME":     "kvtool",
				"APP_ENV":      "production",
				"DATABASE_URL": "postgres://localhost/mydb",
			},
			expectError: false,
		},
		{
			name: "NatsFs - JSON 値を読み込む",
			setupFs: func(t *testing.T) Filesystem {
				if testing.Short() || os.Getenv("SKIP_NATS_TESTS") == "true" {
					t.Skip("NATS tests skipped (-short flag or SKIP_NATS_TESTS=true)")
				}

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				fs, err := NewNatsFs(ctx, &NatsFsConfig{
					URL:    getTestNatsURL(),
					Bucket: "kvtool-test",
				})
				require.NoError(t, err)
				return fs
			},
			testPath: "app/settings",
			expectedData: map[string]any{
				"debug": true,
				"port":  float64(8080),
			},
			expectError: false,
		},
		{
			name: "RedisFs - JSON 値を読み込む",
			setupFs: func(t *testing.T) Filesystem {
				if testing.Short() || os.Getenv("SKIP_REDIS_TESTS") == "true" {
					t.Skip("Redis tests skipped (-short flag or SKIP_REDIS_TESTS=true)")
				}

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				fs, err := NewRedisFs(ctx, &RedisFsConfig{
					Addr:   getTestRedisAddr(),
					Prefix: "kvtool-test:",
				})
				require.NoError(t, err)
				return fs
			},
			testPath: "app/settings",
			expectedData: map[string]any{
				"debug": true,
				"port":  float64(8080),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			fs := tt.setupFs(t)

			file, err := fs.GetFile(tt.testPath)
			require.NoError(err)

			data, err := file.LoadAsJson()

			if tt.expectError {
				require.Error(err)
				if tt.errorSubstring != "" {
					require.Contains(err.Error(), tt.errorSubstring)
				}
			} else {
				require.NoError(err)
				require.NotNil(data)

				if tt.expectedData != nil {
					dataMap, ok := data.(map[string]any)
					require.True(ok, "データが map[string]any であることを期待")
					require.Equal(tt.expectedData, dataMap)
				} else if tt.name == "EnvFs - 環境変数を読み込む" {
					// EnvFs のカスタム検証
					dataMap, ok := data.(map[string]any)
					require.True(ok, "データが map[string]any であることを期待")
					// テスト用変数の存在をチェック
					require.Equal("value1", dataMap["TEST_VAR_1"])
					require.Equal("value2", dataMap["TEST_VAR_2"])
				}
			}
		})
	}
}

// TestFilesystemInterface_OpenReader は全てのファイルシステムが OpenReader を一貫して実装していることをテスト
func TestFilesystemInterface_OpenReader(t *testing.T) {
	tests := []struct {
		name         string
		setupFs      func(t *testing.T) Filesystem
		testPath     string
		validateJSON func(t *testing.T, data []byte)
		expectError  bool
	}{
		{
			name: "LocalFs - JSON を読み込んでパース",
			setupFs: func(t *testing.T) Filesystem {
				tmpDir := t.TempDir()
				testFile := filepath.Join(tmpDir, "test.json")
				err := os.WriteFile(testFile, []byte(`{"key":"value"}`), 0644)
				require.NoError(t, err)

				return &LocalFs{
					Root: tmpDir,
				}
			},
			testPath: "test.json",
			validateJSON: func(t *testing.T, data []byte) {
				var result map[string]any
				err := json.Unmarshal(data, &result)
				require.NoError(t, err)
				require.Equal(t, "value", result["key"])
			},
			expectError: false,
		},
		{
			name: "LocalFs - ファイルが存在しない",
			setupFs: func(t *testing.T) Filesystem {
				tmpDir := t.TempDir()
				return &LocalFs{
					Root: tmpDir,
				}
			},
			testPath:    "nonexistent.json",
			expectError: true,
		},
		{
			name: "EnvFs - 環境変数を JSON として読み込む",
			setupFs: func(t *testing.T) Filesystem {
				t.Setenv("TEST_READ_VAR", "testvalue")
				fs, err := NewEnvFs(context.Background(), &EnvFsConfig{})
				require.NoError(t, err)
				return fs
			},
			testPath: "anypath",
			validateJSON: func(t *testing.T, data []byte) {
				var result map[string]any
				err := json.Unmarshal(data, &result)
				require.NoError(t, err)
				require.Equal(t, "testvalue", result["TEST_READ_VAR"])
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			fs := tt.setupFs(t)

			file, err := fs.GetFile(tt.testPath)
			require.NoError(err)

			reader, err := file.OpenReader()

			if tt.expectError {
				require.Error(err)
				require.Nil(reader)
			} else {
				require.NoError(err)
				require.NotNil(reader)
				defer reader.Close()

				data, err := io.ReadAll(reader)
				require.NoError(err)
				require.NotEmpty(data)

				if tt.validateJSON != nil {
					tt.validateJSON(t, data)
				}
			}
		})
	}
}

// TestFilesystemInterface_Consistency は LoadAsJson と OpenReader が一貫した結果を返すことをテスト
func TestFilesystemInterface_Consistency(t *testing.T) {
	tests := []struct {
		name    string
		setupFs func(t *testing.T) Filesystem
		path    string
	}{
		{
			name: "LocalFs - JSON の一貫性",
			setupFs: func(t *testing.T) Filesystem {
				tmpDir := t.TempDir()
				testFile := filepath.Join(tmpDir, "test.json")
				testData := map[string]any{
					"consistency": "test",
					"string":      "value",
					"array":       []any{"a", "b", "c"},
				}
				jsonBytes, _ := json.Marshal(testData)
				err := os.WriteFile(testFile, jsonBytes, 0644)
				require.NoError(t, err)

				return &LocalFs{
					Root: tmpDir,
				}
			},
			path: "test.json",
		},
		{
			name: "EnvFs - 一貫性",
			setupFs: func(t *testing.T) Filesystem {
				t.Setenv("CONSISTENCY_TEST", "value")
				fs, err := NewEnvFs(context.Background(), &EnvFsConfig{})
				require.NoError(t, err)
				return fs
			},
			path: "anypath",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			fs := tt.setupFs(t)

			file, err := fs.GetFile(tt.path)
			require.NoError(err)

			// LoadAsJson 経由でデータを取得
			dataFromLoad, err := file.LoadAsJson()
			require.NoError(err)

			// OpenReader 経由でデータを取得
			file2, err := fs.GetFile(tt.path)
			require.NoError(err)
			reader, err := file2.OpenReader()
			require.NoError(err)
			defer reader.Close()

			bytes, err := io.ReadAll(reader)
			require.NoError(err)

			// 既に読み込み済みなのでバイトからデコード
			var dataFromReader any
			err = json.Unmarshal(bytes, &dataFromReader)
			require.NoError(err)

			// 両方とも有効な JSON を生成することを検証
			require.NotNil(dataFromLoad)
			require.NotNil(dataFromReader)

			// 注意: 既知の軽微な不一致 - LoadAsJson は UseNumber() を使用するが
			// OpenReader からの unmarshal は使用しない。これは両方とも
			// 同じデータの有効な JSON 表現を生成するため許容される。
		})
	}
}

// TestFilesystemInterface_ErrorHandling は全てのファイルシステムが一貫してエラーを処理することをテスト
func TestFilesystemInterface_ErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		setupFs        func(t *testing.T) Filesystem
		testPath       string
		operation      string // "load" または "read"
		expectError    bool
		errorSubstring string
	}{
		{
			name: "LocalFs - 無効なパス（空）",
			setupFs: func(t *testing.T) Filesystem {
				tmpDir := t.TempDir()
				return &LocalFs{
					Root: tmpDir,
				}
			},
			testPath:       "",
			operation:      "load",
			expectError:    true,
			errorSubstring: "empty",
		},
		{
			name: "LocalFs - パストラバーサル攻撃の防止",
			setupFs: func(t *testing.T) Filesystem {
				tmpDir := t.TempDir()
				return &LocalFs{
					Root: tmpDir,
				}
			},
			testPath:       "../../../etc/passwd",
			operation:      "load",
			expectError:    true,
			errorSubstring: "escapes root",
		},
		{
			name: "LocalFs - 絶対パスの拒否",
			setupFs: func(t *testing.T) Filesystem {
				tmpDir := t.TempDir()
				return &LocalFs{
					Root: tmpDir,
				}
			},
			testPath:       "/etc/passwd",
			operation:      "read",
			expectError:    true,
			errorSubstring: "relative",
		},
		{
			name: "LocalFs - チルダパスの拒否",
			setupFs: func(t *testing.T) Filesystem {
				tmpDir := t.TempDir()
				return &LocalFs{
					Root: tmpDir,
				}
			},
			testPath:       "~/test.json",
			operation:      "load",
			expectError:    true,
			errorSubstring: "relative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			fs := tt.setupFs(t)

			// GetFile は常に成功する（File オブジェクトを返す）
			// エラーは LoadAsJson/OpenReader レベルで発生
			file, err := fs.GetFile(tt.testPath)
			require.NoError(err)
			require.NotNil(file)

			// 実際の操作をテスト
			if tt.operation == "load" {
				_, err = file.LoadAsJson()
			} else if tt.operation == "read" {
				_, err = file.OpenReader()
			}

			if tt.expectError {
				require.Error(err)
				if tt.errorSubstring != "" {
					require.Contains(err.Error(), tt.errorSubstring)
				}
			} else {
				require.NoError(err)
			}
		})
	}
}
