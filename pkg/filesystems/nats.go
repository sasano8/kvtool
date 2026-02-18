// Package filesystems provides unified filesystem interfaces.
//
// # NATS JetStream KV Filesystem
//
// NATS JetStream KV ファイルシステムは、NATS JetStream Key-Value Store から設定値を読み込みます。
// バケット内のキーをパスとしてアクセスし、値を JSON またはバイト列として取得します。
//
// ## Driver Name
//
// nats
//
// ## Use Cases
//
// - マイクロサービス間の設定共有
// - リアルタイムな設定配信基盤との統合
// - NATS を既に使用している環境での設定管理
//
// ## Security
//
// - トークン認証のサポート
// - ユーザー/パスワード認証のサポート
// - 認証情報ファイル (credentials) のサポート
package filesystems

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// NatsFsConfig は NATS JetStream KV ファイルシステムの設定
type NatsFsConfig struct {
	// URL は NATS サーバーの接続 URL
	URL string `yaml:"url" doc:"NATS サーバーの接続 URL" required:"true" example:"nats://localhost:4222"`

	// Bucket は JetStream KV バケット名
	Bucket string `yaml:"bucket" doc:"JetStream KV バケット名" required:"true" example:"config"`

	// Token は NATS 認証トークン
	Token string `yaml:"token" doc:"NATS 認証トークン" required:"false" default:""`

	// User は NATS ユーザー名（パスワード認証用）
	User string `yaml:"user" doc:"NATS ユーザー名" required:"false" default:""`

	// Password は NATS パスワード（パスワード認証用）
	Password string `yaml:"password" doc:"NATS パスワード" required:"false" default:""`

	// CredsFile は NATS 認証情報ファイルパス
	CredsFile string `yaml:"creds_file" doc:"NATS 認証情報ファイルパス" required:"false" default:""`

	// Timeout は接続およびリクエストタイムアウト
	Timeout time.Duration `yaml:"timeout" doc:"接続およびリクエストタイムアウト" required:"false" default:"10s" example:"30s"`
}

// NatsFs は NATS JetStream KV ファイルシステム
type NatsFs struct {
	ctx     context.Context
	conn    *nats.Conn
	kv      jetstream.KeyValue
	bucket  string
	timeout time.Duration
}

// NatsFile は NATS JetStream KV のエントリを表すファイル
type NatsFile struct {
	fs  *NatsFs
	key string
}

// NewNatsFs は新しい NATS JetStream KV ファイルシステムを作成する
func NewNatsFs(ctx context.Context, config *NatsFsConfig) (*NatsFs, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("url is required")
	}
	if config.Bucket == "" {
		return nil, fmt.Errorf("bucket is required")
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	// NATS 接続オプション
	opts := []nats.Option{
		nats.Timeout(timeout),
	}
	if config.Token != "" {
		opts = append(opts, nats.Token(config.Token))
	}
	if config.User != "" {
		opts = append(opts, nats.UserInfo(config.User, config.Password))
	}
	if config.CredsFile != "" {
		opts = append(opts, nats.UserCredentials(config.CredsFile))
	}

	// NATS 接続
	conn, err := nats.Connect(config.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	// JetStream コンテキスト
	js, err := jetstream.New(conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	// バケット存在確認
	connectCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	kv, err := js.KeyValue(connectCtx, config.Bucket)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to access KV bucket '%s': %w", config.Bucket, err)
	}

	return &NatsFs{
		ctx:     ctx,
		conn:    conn,
		kv:      kv,
		bucket:  config.Bucket,
		timeout: timeout,
	}, nil
}

// GetFile は指定されたキーの値を取得するファイルを返す
func (fs *NatsFs) GetFile(path string) (File, error) {
	return &NatsFile{
		fs:  fs,
		key: path,
	}, nil
}

// Close は NATS 接続を閉じる
func (fs *NatsFs) Close() error {
	if fs.conn != nil {
		fs.conn.Close()
	}
	return nil
}

// fetchValue は NATS KV から値を取得する
func (f *NatsFile) fetchValue() ([]byte, error) {
	ctx, cancel := context.WithTimeout(f.fs.ctx, f.fs.timeout)
	defer cancel()

	entry, err := f.fs.kv.Get(ctx, f.key)
	if err != nil {
		return nil, fmt.Errorf("key not found: %s: %w", f.key, err)
	}

	return entry.Value(), nil
}

// LoadAsJson は値を JSON としてパースして返す
// JSON でない場合は文字列として返す
func (f *NatsFile) LoadAsJson() (any, error) {
	value, err := f.fetchValue()
	if err != nil {
		return nil, err
	}

	// JSON としてパース
	var result any
	if err := json.Unmarshal(value, &result); err != nil {
		// JSON でない場合は文字列として返す
		return string(value), nil
	}
	return result, nil
}

// OpenReader は値をストリームとして返す
func (f *NatsFile) OpenReader() (io.ReadCloser, error) {
	value, err := f.fetchValue()
	if err != nil {
		return nil, err
	}

	// JSON としてパース可能か確認
	var result any
	if err := json.Unmarshal(value, &result); err != nil {
		// JSON でない場合は文字列を JSON エンコード
		encoded, err := json.Marshal(string(value))
		if err != nil {
			return nil, fmt.Errorf("failed to encode value as JSON: %w", err)
		}
		return io.NopCloser(bytes.NewReader(encoded)), nil
	}

	// そのまま返す
	return io.NopCloser(bytes.NewReader(value)), nil
}
