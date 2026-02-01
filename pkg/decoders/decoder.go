package decoders

import "io"

// Decoder は io.Reader からデータを読み取り、指定された形式にデコードするインターフェース
type Decoder interface {
	// Decode reads from io.Reader and returns decoded data
	// The expected input format depends on the implementation
	Decode(r io.Reader) (any, error)
}

// EnvDecoder は KEY=VALUE 形式のデータを JSON オブジェクトにデコードする
//
// 入力フォーマット:
//   - 各行は KEY=VALUE 形式（1行1エントリ）
//   - 空行は無視される
//   - # で始まる行はコメントとして無視される
//   - VALUE はクォート（"..." または '...'）で囲むことができる
//   - ダブルクォート内ではエスケープシーケンス（\n, \t など）が使用可能
//
// 出力:
//   - map[string]any 形式の JSON オブジェクト
//   - すべての値は文字列として扱われる
type EnvDecoder struct{}

// Decode は KEY=VALUE 形式のデータを JSON オブジェクトにデコードする
// Decoder インターフェースを実装
func (d *EnvDecoder) Decode(r io.Reader) (any, error) {
	return EnvToJson(r)
}

// コンパイル時に Decoder インターフェースを実装していることを確認
var _ Decoder = (*EnvDecoder)(nil)