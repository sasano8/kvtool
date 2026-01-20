package sources

import (
	"io"
	"os"
)

type SourceEnv struct{}

func (srouce *SourceEnv) Load() (io.Reader, error) {
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
