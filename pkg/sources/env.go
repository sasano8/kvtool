package sources

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// SourceEnv は環境変数を Source として提供する実装
// os.Environ() から取得した環境変数を KEY=VALUE 形式で返す
type SourceEnv struct{}

// Load は環境変数を KEY=VALUE 形式のストリームとして返す
// 特殊文字を含む値は dotenv 形式に従ってエスケープされる
// Source インターフェースを実装
func (s *SourceEnv) Load() (io.Reader, error) {
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		for _, kv := range os.Environ() {
			// KEY=VALUE を分割
			idx := strings.Index(kv, "=")
			if idx < 0 {
				continue // スキップ（ありえないが念のため）
			}

			key := kv[:idx]
			val := kv[idx+1:]

			// エスケープが必要かチェック（\n, \r, ", スペース）
			// Node.js dotenv 互換: \t, \\ はエスケープしない
			// スペースを含む値もクォートが必要（クォートなしスペースはパースエラー）
			needsQuote := strings.ContainsAny(val, "\n\r\" ")
			if needsQuote {
				// ダブルクォートで囲んでエスケープ
				escapedVal := escapeEnvValue(val)
				line := fmt.Sprintf("%s=\"%s\"\n", key, escapedVal)
				if _, err := io.WriteString(pw, line); err != nil {
					_ = pw.CloseWithError(err)
					return
				}
			} else {
				// そのまま出力
				if _, err := io.WriteString(pw, kv+"\n"); err != nil {
					_ = pw.CloseWithError(err)
					return
				}
			}
		}
	}()
	return pr, nil
}

// escapeEnvValue は dotenv 形式に従って値をエスケープする
// サポートするエスケープ: \n (改行), \r (CR), \" (ダブルクォート)
// Node.js dotenv 互換: \t, \\ はエスケープしない
func escapeEnvValue(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '\n':
			b.WriteString("\\n")
		case '\r':
			b.WriteString("\\r")
		case '"':
			b.WriteString("\\\"")
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}

// コンパイル時に Source インターフェースを実装していることを確認
var _ Source = (*SourceEnv)(nil)
