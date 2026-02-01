package sources

import "io"

// Source は KEY=VALUE 形式でデータを提供するインターフェース
//
// フォーマット仕様:
//   - 各行は KEY=VALUE 形式（1行1エントリ）
//   - 空行は無視される
//   - # で始まる行はコメントとして無視される
//   - KEY と VALUE は = で区切られる
//   - VALUE はクォート（"..." または '...'）で囲むことができる
//
// 例:
//
//	APP_NAME=myapp
//	DATABASE_URL="postgres://localhost/db"
//	# これはコメント
//	DEBUG=true
type Source interface {
	// Load returns an io.Reader that streams data in KEY=VALUE format
	// Each line contains one key-value pair separated by '='
	// Empty lines and lines starting with '#' are treated as comments
	Load() (io.Reader, error)
}