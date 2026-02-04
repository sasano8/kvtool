package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/sasano8/kvtool/pkg/config"
	"github.com/stretchr/testify/require"
)

func TestServer_KVEndpoint(t *testing.T) {
	// テスト用の一時ディレクトリとファイルを作成
	tmpDir, err := os.MkdirTemp("", "kvtool-server-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// テスト用 JSON ファイル
	testData := map[string]any{"key": "value", "number": 42}
	jsonBytes, _ := json.Marshal(testData)
	err = os.WriteFile(filepath.Join(tmpDir, "test.json"), jsonBytes, 0644)
	require.NoError(t, err)

	// テスト用設定
	cfg := &config.KvtoolConfig{
		Version: 1.0,
		Namespaces: map[string]map[string]config.StoreInfo{
			"default": {
				"local": {
					Driver: "local",
					Args: map[string]any{
						"root": tmpDir,
					},
				},
			},
			"production": {
				"local": {
					Driver: "local",
					Args: map[string]any{
						"root": tmpDir,
					},
				},
			},
		},
	}

	ctx := context.Background()
	srv := NewServer(ctx, cfg)
	handler := srv.Handler()

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedBody   map[string]any
	}{
		{
			name:           "正常系: default namespace",
			path:           "/v1/namespaces/default/kv/local/test.json",
			expectedStatus: http.StatusOK,
			expectedBody:   testData,
		},
		{
			name:           "正常系: production namespace",
			path:           "/v1/namespaces/production/kv/local/test.json",
			expectedStatus: http.StatusOK,
			expectedBody:   testData,
		},
		{
			name:           "エラー系: 存在しない namespace",
			path:           "/v1/namespaces/nonexistent/kv/local/test.json",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "エラー系: 存在しないストア",
			path:           "/v1/namespaces/default/kv/nonexistent/test.json",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "エラー系: 存在しないファイル",
			path:           "/v1/namespaces/default/kv/local/nonexistent.json",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			require.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedBody != nil {
				var result map[string]any
				err := json.Unmarshal(rec.Body.Bytes(), &result)
				require.NoError(t, err)
				require.Equal(t, tt.expectedBody["key"], result["key"])
				require.Equal(t, float64(tt.expectedBody["number"].(int)), result["number"])
			}
		})
	}
}

func TestServer_StorageEndpoint(t *testing.T) {
	// テスト用の一時ディレクトリとファイルを作成
	tmpDir, err := os.MkdirTemp("", "kvtool-server-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// テスト用ファイル
	testContent := `{"raw": "content"}`
	err = os.WriteFile(filepath.Join(tmpDir, "test.json"), []byte(testContent), 0644)
	require.NoError(t, err)

	// テスト用設定
	cfg := &config.KvtoolConfig{
		Version: 1.0,
		Namespaces: map[string]map[string]config.StoreInfo{
			"default": {
				"local": {
					Driver: "local",
					Args: map[string]any{
						"root": tmpDir,
					},
				},
			},
			"production": {
				"local": {
					Driver: "local",
					Args: map[string]any{
						"root": tmpDir,
					},
				},
			},
		},
	}

	ctx := context.Background()
	srv := NewServer(ctx, cfg)
	handler := srv.Handler()

	tests := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{
			name:           "正常系: default namespace",
			path:           "/v1/namespaces/default/storage/local/test.json",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "正常系: production namespace",
			path:           "/v1/namespaces/production/storage/local/test.json",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			require.Equal(t, tt.expectedStatus, rec.Code)
			require.Equal(t, "application/octet-stream", rec.Header().Get("Content-Type"))

			body, _ := io.ReadAll(rec.Body)
			require.Equal(t, testContent, string(body))
		})
	}
}

func TestServer_HealthEndpoint(t *testing.T) {
	cfg := &config.KvtoolConfig{
		Version:    1.0,
		Namespaces: map[string]map[string]config.StoreInfo{},
	}

	ctx := context.Background()
	srv := NewServer(ctx, cfg)
	handler := srv.Handler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "ok", rec.Body.String())
}

func TestServer_MethodNotAllowed(t *testing.T) {
	cfg := &config.KvtoolConfig{
		Version:    1.0,
		Namespaces: map[string]map[string]config.StoreInfo{},
	}

	ctx := context.Background()
	srv := NewServer(ctx, cfg)
	handler := srv.Handler()

	tests := []struct {
		name   string
		path   string
		method string
	}{
		{"POST to kv", "/v1/namespaces/default/kv/local/test", http.MethodPost},
		{"PUT to kv", "/v1/namespaces/default/kv/local/test", http.MethodPut},
		{"DELETE to kv", "/v1/namespaces/default/kv/local/test", http.MethodDelete},
		{"POST to storage", "/v1/namespaces/default/storage/local/test", http.MethodPost},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			require.Equal(t, http.StatusMethodNotAllowed, rec.Code)
		})
	}
}
