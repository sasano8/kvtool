package filesystems

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRestFs_GetFile(t *testing.T) {
	// テスト用サーバー
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/data/config":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"key": "value",
				"num": 42,
			})
		case "/api/v1/data/nested/deep/file":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"nested": true})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	tests := []struct {
		name          string
		baseURL       string
		path          string
		expectedKey   string
		expectedValue any
		expectError   bool
		errorContains string
	}{
		{
			name:          "正常系: ルートからファイル取得",
			baseURL:       server.URL + "/api/v1/data",
			path:          "config",
			expectedKey:   "key",
			expectedValue: "value",
		},
		{
			name:          "正常系: ネストしたパス",
			baseURL:       server.URL + "/api/v1/data",
			path:          "nested/deep/file",
			expectedKey:   "nested",
			expectedValue: true,
		},
		{
			name:          "エラー系: 存在しないパス",
			baseURL:       server.URL + "/api/v1/data",
			path:          "notfound",
			expectError:   true,
			errorContains: "not found",
		},
		{
			name:          "エラー系: 絶対パス",
			baseURL:       server.URL + "/api/v1/data",
			path:          "/etc/passwd",
			expectError:   true,
			errorContains: "must be relative",
		},
		{
			name:          "エラー系: パストラバーサル",
			baseURL:       server.URL + "/api/v1/data",
			path:          "../../../etc/passwd",
			expectError:   true,
			errorContains: "must be relative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			config := &RestFsConfig{
				BaseURL: tt.baseURL,
			}

			fs, err := NewRestFs(context.Background(), config)
			require.NoError(err)

			file, err := fs.GetFile(tt.path)
			if tt.expectError && err != nil {
				require.Contains(err.Error(), tt.errorContains)
				return
			}
			require.NoError(err)

			data, err := file.LoadAsJson()
			if tt.expectError {
				require.Error(err)
				require.Contains(err.Error(), tt.errorContains)
				return
			}

			require.NoError(err)
			result := data.(map[string]any)
			require.Equal(tt.expectedValue, result[tt.expectedKey])
		})
	}
}

func TestRestFs_Auth(t *testing.T) {
	// 認証チェックするサーバー
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")

		switch {
		case auth == "Bearer test-token":
			json.NewEncoder(w).Encode(map[string]any{"auth": "bearer"})
		case r.URL.Path == "/basic":
			user, pass, ok := r.BasicAuth()
			if ok && user == "testuser" && pass == "testpass" {
				json.NewEncoder(w).Encode(map[string]any{"auth": "basic"})
				return
			}
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		default:
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		}
	}))
	defer server.Close()

	t.Run("Bearer 認証", func(t *testing.T) {
		require := require.New(t)

		config := &RestFsConfig{
			BaseURL:  server.URL,
			AuthType: "bearer",
			Token:    "test-token",
		}

		fs, err := NewRestFs(context.Background(), config)
		require.NoError(err)

		file, err := fs.GetFile("data")
		require.NoError(err)

		data, err := file.LoadAsJson()
		require.NoError(err)

		result := data.(map[string]any)
		require.Equal("bearer", result["auth"])
	})

	t.Run("Basic 認証", func(t *testing.T) {
		require := require.New(t)

		config := &RestFsConfig{
			BaseURL:  server.URL,
			AuthType: "basic",
			Username: "testuser",
			Password: "testpass",
		}

		fs, err := NewRestFs(context.Background(), config)
		require.NoError(err)

		file, err := fs.GetFile("basic")
		require.NoError(err)

		data, err := file.LoadAsJson()
		require.NoError(err)

		result := data.(map[string]any)
		require.Equal("basic", result["auth"])
	})
}

func TestRestFs_Validation(t *testing.T) {
	t.Run("base_url が必須", func(t *testing.T) {
		require := require.New(t)

		config := &RestFsConfig{}
		_, err := NewRestFs(context.Background(), config)
		require.Error(err)
		require.Contains(err.Error(), "base_url is required")
	})
}
