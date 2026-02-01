package filesystems

import (
	"io"
)

// Filesystem は全てのファイルシステムが実装するインターフェース
type Filesystem interface {
	GetFile(path string) (File, error)
}

// File は全てのファイルオブジェクトが実装するインターフェース
type File interface {
	// LoadAsJson はファイルの内容を JSON としてデコードして返す
	LoadAsJson() (any, error)

	// OpenReader はファイルの内容を JSON エンコードしたストリームとして返す
	OpenReader() (io.ReadCloser, error)
}

type MountConfig struct {
	Dir  *string `json:"dir"`
	File *string `json:"file"`
}

type TransformConfig struct {
	Read  string `json:"read"`
	Write string `json:"write"`
}

type StoreConfig struct {
	Driver string         `json:"version"`
	Args   map[string]any `json:"args"`
	Mount  *MountConfig   `json:"mount"`
}

// ".env":
//       driver: "local"
//       args:
//         ext: ".env"
//         transform:
//           read: dotenv
//           write: dotenv
//         root: "."
//       mount:
//         # dir: ""
//         file: ""
