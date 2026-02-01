package sources

import (
	"io"
	"os"
)

// SourceEnv は環境変数を Source として提供する実装
// os.Environ() から取得した環境変数を KEY=VALUE 形式で返す
type SourceEnv struct{}

// Load は環境変数を KEY=VALUE 形式のストリームとして返す
// Source インターフェースを実装
func (s *SourceEnv) Load() (io.Reader, error) {
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		for _, kv := range os.Environ() {
			if _, err := io.WriteString(pw, kv+"\n"); err != nil {
				_ = pw.CloseWithError(err)
				return
			}
		}
	}()
	return pr, nil
}

// コンパイル時に Source インターフェースを実装していることを確認
var _ Source = (*SourceEnv)(nil)
