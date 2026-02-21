package filesystems

import (
	"context"
	"io"
	"time"
)

// extractTimeout は ctx の Deadline から timeout duration を抽出する
// deadline がない場合は defaultTimeout を返す
func extractTimeout(ctx context.Context, defaultTimeout time.Duration) time.Duration {
	if deadline, ok := ctx.Deadline(); ok {
		if d := time.Until(deadline); d > 0 {
			return d
		}
	}
	return defaultTimeout
}

// Filesystem は全てのファイルシステムが実装するインターフェース
type Filesystem interface {
	GetFile(path string) (File, error)
}

// Syncable はリモートソースからの同期をサポートするファイルシステム
type Syncable interface {
	Sync() error
}

// File は全てのファイルオブジェクトが実装するインターフェース
type File interface {
	// LoadAsJson はファイルの内容を JSON としてデコードして返す
	LoadAsJson() (any, error)

	// OpenReader はファイルの内容を JSON エンコードしたストリームとして返す
	OpenReader() (io.ReadCloser, error)
}
