// Package filesystems provides unified filesystem interfaces.
//
// # Database Filesystem
//
// データベースから設定値を読み込むファイルシステムです。
// ユーザー定義の SQL クエリを使用して、任意のテーブルから値を取得できます。
//
// ## Driver Name
//
// db
//
// ## Use Cases
//
// - マルチテナント設定管理
// - データベースに保存された設定の取得
// - 既存の設定テーブルとの統合
//
// ## Security
//
// - プリペアドステートメントによる SQL インジェクション防止
// - {key} と {namespace} はプレースホルダーに変換
package filesystems

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

// DbFsConfig はデータベースファイルシステムの設定
type DbFsConfig struct {
	// ConnectionString はデータベース接続文字列
	// PostgreSQL: postgres://user:pass@localhost/dbname
	// MySQL: user:pass@tcp(localhost:3306)/dbname
	// SQLite: file:test.db
	ConnectionString string `yaml:"connection_string" doc:"データベース接続文字列" required:"true" example:"postgres://user:pass@localhost/db"`

	// Driver はデータベースドライバー名（postgres, mysql, sqlite）
	// 省略時は接続文字列から自動判定
	Driver string `yaml:"driver" doc:"データベースドライバー（postgres, mysql, sqlite）。省略時は接続文字列から自動判定" required:"false"`

	// Query は値を取得する SQL クエリ
	// {key} と {namespace} はプレースホルダーとして展開される
	Query string `yaml:"query" doc:"SQL クエリ。{key} と {namespace} がプレースホルダーとして使用可能" required:"true" example:"SELECT value FROM config WHERE key = {key}"`

	// Namespace はデフォルトの namespace 値
	Namespace string `yaml:"namespace" doc:"デフォルトの namespace 値" required:"false" default:"default" example:"production"`
}

// DbFs はデータベースファイルシステム
type DbFs struct {
	ctx     context.Context
	db      *sql.DB
	config  *DbFsConfig
	query   string // プレースホルダー変換後のクエリ
	hasKey  bool   // クエリに {key} が含まれるか
	hasNs   bool   // クエリに {namespace} が含まれるか
	timeout time.Duration
}

// DbFile はデータベースから取得したファイル
type DbFile struct {
	ctx     context.Context
	db      *sql.DB
	query   string
	key     string
	ns      string
	hasKey  bool
	hasNs   bool
	timeout time.Duration
}

// NewDbFs は新しいデータベースファイルシステムを作成する
func NewDbFs(ctx context.Context, config *DbFsConfig) (*DbFs, error) {
	timeout := extractTimeout(ctx, 10*time.Second)

	if config.ConnectionString == "" {
		return nil, fmt.Errorf("connection_string is required")
	}
	if config.Query == "" {
		return nil, fmt.Errorf("query is required")
	}

	// ドライバーを判定
	driver := config.Driver
	if driver == "" {
		driver = detectDriver(config.ConnectionString)
		if driver == "" {
			return nil, fmt.Errorf("could not detect database driver from connection string, please specify 'driver' explicitly")
		}
	}

	// データベース接続
	db, err := sql.Open(driver, config.ConnectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 接続確認
	pingCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// クエリ内のプレースホルダーを検出
	hasKey := strings.Contains(config.Query, "{key}")
	hasNs := strings.Contains(config.Query, "{namespace}")

	// プレースホルダーを SQL パラメータに変換
	query := convertPlaceholders(config.Query, driver, hasKey, hasNs)

	return &DbFs{
		ctx:     context.Background(),
		db:      db,
		config:  config,
		query:   query,
		hasKey:  hasKey,
		hasNs:   hasNs,
		timeout: timeout,
	}, nil
}

// GetFile はデータベースから値を取得する
func (fs *DbFs) GetFile(path string) (File, error) {
	ns := fs.config.Namespace
	if ns == "" {
		ns = "default"
	}

	return &DbFile{
		ctx:     fs.ctx,
		db:      fs.db,
		query:   fs.query,
		key:     path,
		ns:      ns,
		hasKey:  fs.hasKey,
		hasNs:   fs.hasNs,
		timeout: fs.timeout,
	}, nil
}

// Close はデータベース接続を閉じる
func (fs *DbFs) Close() error {
	if fs.db != nil {
		return fs.db.Close()
	}
	return nil
}

// LoadAsJson はクエリ結果を JSON として返す
func (f *DbFile) LoadAsJson() (any, error) {
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

// OpenReader はクエリ結果をストリームとして返す
func (f *DbFile) OpenReader() (io.ReadCloser, error) {
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
	return io.NopCloser(strings.NewReader(value)), nil
}

// fetchValue はデータベースから値を取得する
func (f *DbFile) fetchValue() (string, error) {
	ctx, cancel := context.WithTimeout(f.ctx, f.timeout)
	defer cancel()

	// パラメータを構築
	args := buildQueryArgs(f.key, f.ns, f.hasKey, f.hasNs)

	// クエリ実行
	var value string
	err := f.db.QueryRowContext(ctx, f.query, args...).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("key not found: %s", f.key)
		}
		return "", fmt.Errorf("query failed: %w", err)
	}

	return value, nil
}

// detectDriver は接続文字列からドライバーを判定する
func detectDriver(connStr string) string {
	connStr = strings.ToLower(connStr)
	switch {
	case strings.HasPrefix(connStr, "postgres://"), strings.HasPrefix(connStr, "postgresql://"):
		return "postgres"
	case strings.Contains(connStr, "@tcp("), strings.Contains(connStr, "mysql"):
		return "mysql"
	case strings.HasPrefix(connStr, "file:"), strings.HasSuffix(connStr, ".db"):
		return "sqlite"
	default:
		return ""
	}
}

// convertPlaceholders は {key} と {namespace} を SQL パラメータに変換する
func convertPlaceholders(query, driver string, hasKey, hasNs bool) string {
	result := query
	paramIdx := 1

	// PostgreSQL は $1, $2 形式
	// MySQL/SQLite は ? 形式
	useDollar := driver == "postgres"

	if hasKey {
		placeholder := "?"
		if useDollar {
			placeholder = fmt.Sprintf("$%d", paramIdx)
			paramIdx++
		}
		result = strings.ReplaceAll(result, "{key}", placeholder)
	}

	if hasNs {
		placeholder := "?"
		if useDollar {
			placeholder = fmt.Sprintf("$%d", paramIdx)
		}
		result = strings.ReplaceAll(result, "{namespace}", placeholder)
	}

	return result
}

// buildQueryArgs はクエリパラメータを構築する
func buildQueryArgs(key, ns string, hasKey, hasNs bool) []any {
	var args []any
	if hasKey {
		args = append(args, key)
	}
	if hasNs {
		args = append(args, ns)
	}
	return args
}
