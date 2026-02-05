package filesystems

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestToolFs(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expectError bool
		checkValue  func(t *testing.T, data any)
	}{
		{
			name: "uuid7",
			path: "uuid7",
			checkValue: func(t *testing.T, data any) {
				value, ok := data.(string)
				require.True(t, ok)
				require.Len(t, value, 36) // UUID format
			},
		},
		{
			name: "uuid4",
			path: "uuid4",
			checkValue: func(t *testing.T, data any) {
				value, ok := data.(string)
				require.True(t, ok)
				require.Len(t, value, 36)
			},
		},
		{
			name: "now",
			path: "now",
			checkValue: func(t *testing.T, data any) {
				value, ok := data.(string)
				require.True(t, ok)
				require.Contains(t, value, "T") // RFC3339 format
			},
		},
		{
			name: "timestamp",
			path: "timestamp",
			checkValue: func(t *testing.T, data any) {
				// int64 として返される
				value, ok := data.(int64)
				require.True(t, ok)
				require.Greater(t, value, int64(0))
			},
		},
		{
			name: "random/hex/16",
			path: "random/hex/16",
			checkValue: func(t *testing.T, data any) {
				value, ok := data.(string)
				require.True(t, ok)
				require.Len(t, value, 32) // 16 bytes = 32 hex chars
			},
		},
		{
			name: "random/base64/16",
			path: "random/base64/16",
			checkValue: func(t *testing.T, data any) {
				value, ok := data.(string)
				require.True(t, ok)
				require.NotEmpty(t, value)
			},
		},
		{
			name: "password",
			path: "password",
			checkValue: func(t *testing.T, data any) {
				value, ok := data.(string)
				require.True(t, ok)
				require.Len(t, value, 16)
			},
		},
		{
			name:        "unknown tool",
			path:        "unknown",
			expectError: true,
		},
		{
			name:        "absolute path",
			path:        "/uuid7",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			fs, err := NewToolFs(context.Background(), &ToolFsConfig{})
			require.NoError(err)

			file, err := fs.GetFile(tt.path)
			if tt.expectError && err != nil {
				return
			}
			require.NoError(err)

			data, err := file.LoadAsJson()
			if tt.expectError {
				require.Error(err)
				return
			}

			require.NoError(err)
			if tt.checkValue != nil {
				tt.checkValue(t, data)
			}
		})
	}
}
