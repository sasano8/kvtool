package filesystems

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// getTestMinIOEndpoint returns MinIO endpoint for testing
// Makefile の test-ci ターゲットで MINIO_ENDPOINT を設定する
func getTestMinIOEndpoint() string {
	if endpoint := os.Getenv("MINIO_ENDPOINT"); endpoint != "" {
		return endpoint
	}
	return "http://localhost:9000" // ローカル開発用デフォルト
}

// TestNewS3Fs_Validation は S3FsConfig のバリデーションをテスト
func TestNewS3Fs_Validation(t *testing.T) {
	// MinIO が起動していない場合はスキップ
	if testing.Short() || os.Getenv("SKIP_S3_TESTS") == "true" {
		t.Skip("MinIO tests skipped (-short flag or SKIP_S3_TESTS=true)")
	}

	tests := []struct {
		name           string
		config         S3FsConfig
		expectError    bool
		errorSubstring string
	}{
		{
			name: "正常な設定",
			config: S3FsConfig{
				Bucket:          "kvtool-test",
				Region:          "us-east-1",
				Endpoint:        getTestMinIOEndpoint(),
				UsePathStyle:    true,
				AccessKeyID:     "minioadmin",
				SecretAccessKey: "minioadmin",
			},
			expectError: false,
		},
		{
			name: "バケット名が空",
			config: S3FsConfig{
				Region: "us-east-1",
			},
			expectError:    true,
			errorSubstring: "bucket is required",
		},
		{
			name: "リージョンが空",
			config: S3FsConfig{
				Bucket: "my-bucket",
			},
			expectError:    true,
			errorSubstring: "region is required",
		},
		{
			name: "バケット名にスラッシュが含まれる",
			config: S3FsConfig{
				Bucket: "my-bucket/config",
				Region: "us-east-1",
			},
			expectError:    true,
			errorSubstring: "bucket name must not contain '/'",
		},
		{
			name: "バケット名が完全なパス",
			config: S3FsConfig{
				Bucket: "my-bucket/config/production",
				Region: "us-east-1",
			},
			expectError:    true,
			errorSubstring: "did you mean to use 'root'?",
		},
		{
			name: "エンドポイントにパスが含まれる（許可）",
			config: S3FsConfig{
				Bucket:          "kvtool-test",
				Region:          "us-east-1",
				Endpoint:        getTestMinIOEndpoint(),
				UsePathStyle:    true,
				AccessKeyID:     "minioadmin",
				SecretAccessKey: "minioadmin",
			},
			expectError: false,
		},
		{
			name: "エンドポイントが不正な URL",
			config: S3FsConfig{
				Bucket:   "my-bucket",
				Region:   "us-east-1",
				Endpoint: "://invalid-url",
			},
			expectError:    true,
			errorSubstring: "invalid endpoint URL",
		},
		{
			name: "エンドポイントにスラッシュのみ（許可）",
			config: S3FsConfig{
				Bucket:          "kvtool-test",
				Region:          "us-east-1",
				Endpoint:        "http://localhost:9000/",
				UsePathStyle:    true,
				AccessKeyID:     "minioadmin",
				SecretAccessKey: "minioadmin",
			},
			expectError: false,
		},
		{
			name: "存在しないバケット（HeadBucket で検証）",
			config: S3FsConfig{
				Bucket:          "non-existent-bucket-12345",
				Region:          "us-east-1",
				Endpoint:        getTestMinIOEndpoint(),
				UsePathStyle:    true,
				AccessKeyID:     "minioadmin",
				SecretAccessKey: "minioadmin",
			},
			expectError:    true,
			errorSubstring: "failed to verify bucket",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			fs, err := NewS3Fs(context.Background(), tt.config)

			if tt.expectError {
				require.Error(err)
				if tt.errorSubstring != "" {
					require.Contains(err.Error(), tt.errorSubstring)
				}
				require.Nil(fs)
			} else {
				require.NoError(err)
				require.NotNil(fs)
			}
		})
	}
}
