// Package filesystems provides unified filesystem interfaces.
//
// # Redis Filesystem
//
// Redis ファイルシステムは、Redis から設定値を読み込みます。
// キーをパスとしてアクセスし、値を JSON またはバイト列として取得します。
//
// ## Driver Name
//
// redis
//
// ## Use Cases
//
// - アプリケーション設定のキャッシュ・配信
// - セッション情報や機能フラグの管理
// - Redis を既に使用している環境での設定管理
//
// ## Security
//
// - パスワード認証のサポート
// - TLS 接続のサポート
// - データベース番号による分離
package filesystems

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisFsConfig は Redis ファイルシステムの設定
type RedisFsConfig struct {
	// Addr は Redis サーバーのアドレス
	Addr string `yaml:"addr" doc:"Redis サーバーのアドレス" required:"true" example:"localhost:6379"`

	// Password は Redis 認証パスワード
	Password string `yaml:"password" doc:"Redis 認証パスワード" required:"false" default:""`

	// DB は Redis データベース番号
	DB int `yaml:"db" doc:"Redis データベース番号" required:"false" default:"0" example:"0"`

	// Prefix はキーのプレフィックス
	Prefix string `yaml:"prefix" doc:"キーのプレフィックス" required:"false" default:"" example:"config:"`

	// Timeout は接続およびリクエストタイムアウト
	Timeout time.Duration `yaml:"timeout" doc:"接続およびリクエストタイムアウト" required:"false" default:"10s" example:"30s"`
}

// RedisFs は Redis ファイルシステム
type RedisFs struct {
	ctx     context.Context
	client  *redis.Client
	prefix  string
	timeout time.Duration
}

// RedisFile は Redis のエントリを表すファイル
type RedisFile struct {
	fs  *RedisFs
	key string
}

// NewRedisFs は新しい Redis ファイルシステムを作成する
func NewRedisFs(ctx context.Context, config *RedisFsConfig) (*RedisFs, error) {
	if config.Addr == "" {
		return nil, fmt.Errorf("addr is required")
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	client := redis.NewClient(&redis.Options{
		Addr:        config.Addr,
		Password:    config.Password,
		DB:          config.DB,
		DialTimeout: timeout,
		ReadTimeout: timeout,
	})

	// 接続確認
	pingCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to connect to Redis at '%s': %w", config.Addr, err)
	}

	return &RedisFs{
		ctx:     ctx,
		client:  client,
		prefix:  config.Prefix,
		timeout: timeout,
	}, nil
}

// GetFile は指定されたキーの値を取得するファイルを返す
func (fs *RedisFs) GetFile(path string) (File, error) {
	return &RedisFile{
		fs:  fs,
		key: fs.prefix + path,
	}, nil
}

// Close は Redis 接続を閉じる
func (fs *RedisFs) Close() error {
	if fs.client != nil {
		return fs.client.Close()
	}
	return nil
}

// fetchValue は Redis から値を取得する
func (f *RedisFile) fetchValue() (string, error) {
	ctx, cancel := context.WithTimeout(f.fs.ctx, f.fs.timeout)
	defer cancel()

	val, err := f.fs.client.Get(ctx, f.key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("key not found: %s", f.key)
		}
		return "", fmt.Errorf("failed to get key '%s': %w", f.key, err)
	}

	return val, nil
}

// LoadAsJson は値を JSON としてパースして返す
// JSON でない場合は文字列として返す
func (f *RedisFile) LoadAsJson() (any, error) {
	value, err := f.fetchValue()
	if err != nil {
		return nil, err
	}

	// JSON としてパース
	var result any
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		// JSON でない場合は文字列として返す
		return value, nil
	}
	return result, nil
}

// OpenReader は値をストリームとして返す
func (f *RedisFile) OpenReader() (io.ReadCloser, error) {
	value, err := f.fetchValue()
	if err != nil {
		return nil, err
	}

	// JSON としてパース可能か確認
	var result any
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		// JSON でない場合は文字列を JSON エンコード
		encoded, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("failed to encode value as JSON: %w", err)
		}
		return io.NopCloser(bytes.NewReader(encoded)), nil
	}

	// そのまま返す
	return io.NopCloser(bytes.NewReader([]byte(value))), nil
}
