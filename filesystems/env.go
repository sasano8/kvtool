package filesystems

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/sasano8/kvtool/pkg/decoders"
	"github.com/sasano8/kvtool/pkg/sources"
)

// FsEnvFilesystem は環境変数をファイルシステムとして扱う
// パスは無視され、常に全環境変数を返す
type FsEnvFilesystem struct {
	// Root はインターフェース統一のため存在するが使用されない
	Root string
}

// GetFile は Filesystem インターフェースを実装
// path は無視され、常に環境変数ファイルを返す
func (fs *FsEnvFilesystem) GetFile(path string) (File, error) {
	// パスは無視（環境変数は単一のグローバルソース）
	return &FsEnvFile{}, nil
}

// FsEnvFile は環境変数を表すファイル
type FsEnvFile struct {
}

// LoadAsJson は全環境変数を JSON として返す
func (f *FsEnvFile) LoadAsJson() (any, error) {
	source := sources.SourceEnv{}
	stream, err := source.Load()
	if err != nil {
		return nil, err
	}

	result, err := decoders.EnvToJson(stream)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// OpenReader は JSON エンコードされたストリームを返す
func (f *FsEnvFile) OpenReader() (io.ReadCloser, error) {
	data, err := f.LoadAsJson()
	if err != nil {
		return nil, err
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to encode JSON: %w", err)
	}

	return io.NopCloser(bytes.NewReader(jsonBytes)), nil
}
