package filesystems

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/require"
)

func TestNatsFs(t *testing.T) {
	if testing.Short() || os.Getenv("SKIP_NATS_TESTS") == "true" {
		t.Skip("NATS tests skipped (-short flag or SKIP_NATS_TESTS=true)")
	}

	ctx := context.Background()
	natsURL := getTestNatsURL()

	// テストデータをセットアップ
	conn, err := nats.Connect(natsURL)
	require.NoError(t, err)
	defer conn.Close()

	js, err := jetstream.New(conn)
	require.NoError(t, err)

	// テスト用バケット作成（既存なら削除して再作成）
	_ = js.DeleteKeyValue(ctx, "kvtool-nats-test")
	kv, err := js.CreateKeyValue(ctx, jetstream.KeyValueConfig{Bucket: "kvtool-nats-test"})
	require.NoError(t, err)
	defer js.DeleteKeyValue(ctx, "kvtool-nats-test")

	_, err = kv.Put(ctx, "app.settings", []byte(`{"debug": true, "port": 8080}`))
	require.NoError(t, err)
	_, err = kv.Put(ctx, "simple", []byte(`hello world`))
	require.NoError(t, err)

	tests := []struct {
		name        string
		key         string
		expected    any
		expectError bool
	}{
		{
			name: "JSON値の取得",
			key:  "app.settings",
			expected: map[string]any{
				"debug": true,
				"port":  float64(8080),
			},
		},
		{
			name:     "文字列値の取得",
			key:      "simple",
			expected: "hello world",
		},
		{
			name:        "存在しないキー",
			key:         "nonexistent",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			config := &NatsFsConfig{
				URL:    natsURL,
				Bucket: "kvtool-nats-test",
			}

			natsCtx, natsCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer natsCancel()
			fs, err := NewNatsFs(natsCtx, config)
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

func TestNatsFs_ConfigValidation(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		config      *NatsFsConfig
		expectError string
	}{
		{
			name:        "URL なし",
			config:      &NatsFsConfig{Bucket: "test"},
			expectError: "url is required",
		},
		{
			name:        "バケット名なし",
			config:      &NatsFsConfig{URL: "nats://localhost:4222"},
			expectError: "bucket is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			natsCtx, natsCancel := context.WithTimeout(ctx, 5*time.Second)
			defer natsCancel()
			_, err := NewNatsFs(natsCtx, tt.config)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectError)
		})
	}
}
