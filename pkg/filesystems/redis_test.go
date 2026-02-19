package filesystems

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestRedisFs(t *testing.T) {
	if testing.Short() || os.Getenv("SKIP_REDIS_TESTS") == "true" {
		t.Skip("Redis tests skipped (-short flag or SKIP_REDIS_TESTS=true)")
	}

	ctx := context.Background()
	redisAddr := getTestRedisAddr()

	// テストデータをセットアップ
	client := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer client.Close()

	err := client.Ping(ctx).Err()
	require.NoError(t, err)

	client.Set(ctx, "kvtool-test:app/settings", `{"debug": true, "port": 8080}`, 0)
	client.Set(ctx, "kvtool-test:simple", `hello world`, 0)
	defer client.Del(ctx, "kvtool-test:app/settings", "kvtool-test:simple")

	tests := []struct {
		name        string
		key         string
		expected    any
		expectError bool
	}{
		{
			name: "JSON値の取得",
			key:  "app/settings",
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

			config := &RedisFsConfig{
				Addr:    redisAddr,
				Prefix:  "kvtool-test:",
				Timeout: 5 * time.Second,
			}

			fs, err := NewRedisFs(ctx, config)
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

func TestRedisFs_ConfigValidation(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		config      *RedisFsConfig
		expectError string
	}{
		{
			name:        "アドレスなし",
			config:      &RedisFsConfig{},
			expectError: "addr is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewRedisFs(ctx, tt.config)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectError)
		})
	}
}
