// Package filesystems provides unified filesystem interfaces.
//
// # S3 Filesystem
//
// S3 ファイルシステムは、Amazon S3 および S3 互換ストレージからファイルを読み込みます。
//
// ## Driver Name
//
// s3
//
// ## Use Cases
//
// - CI/CD パイプラインでの設定取得
// - 複数環境での設定の中央管理
// - S3 互換ストレージ（MinIO, Cloudflare R2）との統合
//
// ## Example Configuration
//
//	version: 0.1
//	namespaces:
//	  default:
//	    s3config:
//	      driver: s3
//	      args:
//	        bucket: my-config-bucket
//	        region: ap-northeast-1
//	      mount:
//	        dir: config
//
// ## Security
//
// - パストラバーサル攻撃の防止
// - 相対パスのみサポート
// - IAM ロールによる認証推奨
package filesystems

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sasano8/kvtool/pkg/decoders"
)

// S3FsConfig は S3 ファイルシステムの設定
type S3FsConfig struct {
	// Bucket は S3 バケット名（必須）
	Bucket string `yaml:"bucket" doc:"S3 バケット名" required:"true" example:"my-config-bucket"`

	// Region は AWS リージョン（必須）
	Region string `yaml:"region" doc:"AWS リージョン" required:"true" example:"ap-northeast-1"`

	// AccessKeyID は AWS アクセスキー ID（必須）
	AccessKeyID string `yaml:"access_key_id" doc:"AWS アクセスキー ID" required:"true" example:"AKIAIOSFODNN7EXAMPLE"`

	// SecretAccessKey は AWS シークレットアクセスキー（必須）
	SecretAccessKey string `yaml:"secret_access_key" doc:"AWS シークレットアクセスキー" required:"true" example:"wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"`

	// SessionToken は AWS セッショントークン（オプション、一時認証情報使用時）
	SessionToken string `yaml:"session_token" doc:"AWS セッショントークン（一時的な認証情報使用時）" required:"false" default:"" example:"FwoGZXIvYXdzE..."`

	// Endpoint はカスタムエンドポイント（オプション、S3 互換ストレージ用）
	Endpoint string `yaml:"endpoint" doc:"カスタムエンドポイント（MinIO など S3 互換ストレージ用）" required:"false" default:"" example:"http://localhost:9000"`

	// UsePathStyle はパススタイルアクセスの使用（オプション）
	UsePathStyle bool `yaml:"use_path_style" doc:"パススタイルアクセスを使用（S3 互換ストレージで必要な場合あり）" required:"false" default:"false" example:"true"`

	// Transform は読み込み時の変換方法（オプション）
	Transform string `yaml:"transform" doc:"読み込み時の変換方法（dotenv, json など）" required:"false" default:"" example:"dotenv"`
}

// S3Fs は S3 ファイルシステムの実装
type S3Fs struct {
	client    *s3.Client
	bucket    string
	transform string
}

// NewS3Fs は S3 ファイルシステムを作成します。
//
// 認証情報は設定ファイルで明示的に指定する必要があります。
// 環境変数からの自動読み込みは行いません。
//
// エラーが発生する場合：
//   - 設定が不正な場合（バケット名、region、認証情報が空）
//   - S3 への接続に失敗した場合
func NewS3Fs(ctx context.Context, cfg S3FsConfig) (*S3Fs, error) {
	timeout := extractTimeout(ctx, 30*time.Second)
	// バケット名と region は必須
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("bucket is required")
	}
	if cfg.Region == "" {
		return nil, fmt.Errorf("region is required")
	}

	// バケット名にパスが含まれていないかチェック
	// endpoint + bucket がバケットを示すため、bucket 自体にスラッシュは不可
	if strings.Contains(cfg.Bucket, "/") {
		return nil, fmt.Errorf("bucket name must not contain '/': %s (did you mean to use 'root'?)", cfg.Bucket)
	}

	// エンドポイントの URL 形式チェックのみ（パスの有無はチェックしない）
	// マルチテナント環境で endpoint にパスを含めるケースに対応
	if cfg.Endpoint != "" {
		_, err := url.Parse(cfg.Endpoint)
		if err != nil {
			return nil, fmt.Errorf("invalid endpoint URL: %s: %w", cfg.Endpoint, err)
		}
	}

	// AWS 設定を読み込み
	var awsCfg aws.Config
	var err error

	// 認証情報は明示的に指定が必須（環境変数からは読み込まない）
	if cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" {
		return nil, fmt.Errorf("access_key_id and secret_access_key are required")
	}

	awsCfg, err = config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			cfg.SessionToken,
		)),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// S3 クライアントのオプション
	s3Options := []func(*s3.Options){
		func(o *s3.Options) {
			o.Region = cfg.Region
		},
	}

	// カスタムエンドポイント（MinIO など）
	if cfg.Endpoint != "" {
		s3Options = append(s3Options, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = cfg.UsePathStyle
		})
	}

	client := s3.NewFromConfig(awsCfg, s3Options...)

	fs := &S3Fs{
		client:    client,
		bucket:    cfg.Bucket,
		transform: cfg.Transform,
	}

	// バケットの存在を検証（endpoint + bucket が正しいバケットを示しているか確認）
	headCtx, headCancel := context.WithTimeout(ctx, timeout)
	defer headCancel()

	_, err = client.HeadBucket(headCtx, &s3.HeadBucketInput{
		Bucket: aws.String(cfg.Bucket),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to verify bucket '%s': %w (check that endpoint + bucket points to a valid S3 bucket)", cfg.Bucket, err)
	}

	return fs, nil
}

// GetFile は指定されたパスのファイルを取得します
func (fs *S3Fs) GetFile(path string) (File, error) {
	// 絶対パスとチルダパスを拒否
	if filepath.IsAbs(path) || strings.HasPrefix(path, "~") {
		return nil, fmt.Errorf("path must be relative: %s", path)
	}

	// パストラバーサルチェック
	cleanPath := filepath.Clean(path)
	if strings.HasPrefix(cleanPath, "..") {
		return nil, fmt.Errorf("path escapes root: %s", path)
	}

	// S3 のパスはフォワードスラッシュを使用
	s3Path := strings.ReplaceAll(cleanPath, "\\", "/")

	return &S3File{
		client:    fs.client,
		bucket:    fs.bucket,
		path:      s3Path,
		transform: fs.transform,
	}, nil
}

// S3File は S3 上のファイルを表します
type S3File struct {
	client    *s3.Client
	bucket    string
	path      string
	transform string
}

// LoadAsJson はファイルを JSON として読み込みます
func (f *S3File) LoadAsJson() (any, error) {
	reader, err := f.OpenReader()
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// Transform 設定に基づいて適切なデコーダーを選択
	switch f.transform {
	case "dotenv":
		// dotenv 形式として読み込み
		return decoders.DotenvToJson(reader)
	case "", "json":
		// JSON 形式として読み込み（デフォルト）
		var v any
		if err := json.NewDecoder(reader).Decode(&v); err != nil {
			return nil, fmt.Errorf("failed to decode JSON: %w", err)
		}
		return v, nil
	default:
		return nil, fmt.Errorf("unsupported transform format: %s", f.transform)
	}
}

// OpenReader はファイルの ReadCloser を返します
func (f *S3File) OpenReader() (io.ReadCloser, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	output, err := f.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(f.bucket),
		Key:    aws.String(f.path),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get S3 object %s/%s: %w", f.bucket, f.path, err)
	}

	return output.Body, nil
}
