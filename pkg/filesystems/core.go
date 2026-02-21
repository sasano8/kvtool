package filesystems

import (
	"io"
)

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
