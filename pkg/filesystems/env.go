// Package filesystems provides unified filesystem interfaces.
//
// # Environment Variable Filesystem
//
// 環境変数ファイルシステムは、OS の環境変数をキーバリューストアとして読み込みます。
//
// ## Driver Name
//
// env
//
// ## Use Cases
//
// - 環境変数からアプリケーション設定を取得
// - CI/CD パイプラインでの設定注入
//
// ## Security
//
// - include/exclude フィルタで露出する環境変数を制御可能
package filesystems

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/sasano8/kvtool/pkg/decoders"
	"github.com/sasano8/kvtool/pkg/sources"
)

// EnvFsConfig は環境変数ファイルシステムの設定
type EnvFsConfig struct {
	// Include は許可する環境変数キーのリスト
	Include []string `yaml:"include" doc:"許可する環境変数キーのリスト。未指定で全許可" required:"false" default:"" example:"APP_NAME,DATABASE_URL"`

	// Exclude は除外する環境変数キーのリスト
	Exclude []string `yaml:"exclude" doc:"除外する環境変数キーのリスト。include より優先" required:"false" default:"" example:"SECRET_KEY,PASSWORD"`
}

// FsEnvFilesystem は環境変数をファイルシステムとして扱う
// パスは無視され、常に環境変数を返す
type FsEnvFilesystem struct {
	include map[string]struct{}
	exclude map[string]struct{}
}

// NewEnvFs は環境変数ファイルシステムを生成する
func NewEnvFs(_ context.Context, config *EnvFsConfig) (*FsEnvFilesystem, error) {
	fs := &FsEnvFilesystem{}
	if len(config.Include) > 0 {
		fs.include = toSet(config.Include)
	}
	if len(config.Exclude) > 0 {
		fs.exclude = toSet(config.Exclude)
	}
	return fs, nil
}

func toSet(keys []string) map[string]struct{} {
	m := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		m[k] = struct{}{}
	}
	return m
}

// GetFile は Filesystem インターフェースを実装
// path は無視され、常に環境変数ファイルを返す
func (fs *FsEnvFilesystem) GetFile(path string) (File, error) {
	// パスは無視（環境変数は単一のグローバルソース）
	return &FsEnvFile{
		fs: fs,
	}, nil
}

// FsEnvFile は環境変数を表すファイル
type FsEnvFile struct {
	fs *FsEnvFilesystem
}

// LoadAsJson は環境変数を JSON として返す
// Source と Decoder のインターフェースを通じて、疎結合な設計を実現
func (f *FsEnvFile) LoadAsJson() (any, error) {
	// Source インターフェースを通じてデータを取得
	var source sources.Source = &sources.SourceEnv{}
	stream, err := source.Load()
	if err != nil {
		return nil, err
	}

	// Decoder インターフェースを通じてデータを変換
	// os.Environ() からのデータは OS が保証するため、信頼モードで処理
	var decoder decoders.Decoder = &decoders.EnvDecoder{Trusted: true}
	result, err := decoder.Decode(stream)
	if err != nil {
		return nil, err
	}

	// include/exclude フィルタ適用
	if f.fs.include != nil || f.fs.exclude != nil {
		if m, ok := result.(map[string]any); ok {
			result = filterKeys(m, f.fs.include, f.fs.exclude)
		}
	}

	return result, nil
}

// filterKeys は include/exclude に基づいてマップのキーをフィルタする
// exclude が include より優先される
func filterKeys(m map[string]any, include, exclude map[string]struct{}) map[string]any {
	filtered := make(map[string]any)
	for k, v := range m {
		if exclude != nil {
			if _, excluded := exclude[k]; excluded {
				continue
			}
		}
		if include != nil {
			if _, included := include[k]; !included {
				continue
			}
		}
		filtered[k] = v
	}
	return filtered
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
