package filesystems

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sasano8/kvtool/pkg/decoders"
)

// 名前なしフィールド（埋め込みフィールド）を定義すると、内側の構造体のメソッドを外側の構造体から呼べる（method promotion）
type nopWriteCloser struct {
	io.Writer
}

func (n nopWriteCloser) Close() error { return nil }

type Path string

type LocalFsConfig struct {
	Timeout   time.Duration
	Root      string
	Transform string // "dotenv", "json", "" など。空文字列の場合はデフォルト（json）
}

type FsLocalFile struct {
	Path    string
	NoMkdir bool
}

type LocalFs struct {
	Ctx       context.Context
	Timeout   time.Duration
	Root      string
	Transform string // ファイル読み込み時の変換形式
}

type LocalFile struct {
	fs   *LocalFs
	path string
}

func GetLocalFs(parent context.Context, config *LocalFsConfig) (*LocalFs, error) {
	root := strings.TrimSpace(config.Root)
	fs := LocalFs{
		Ctx:       parent,
		Timeout:   config.Timeout,
		Root:      root,
		Transform: config.Transform,
	}
	return &fs, nil
}

// GetFile は Filesystem インターフェースを実装
func (fs *LocalFs) GetFile(path string) (File, error) {
	return &LocalFile{
		fs:   fs,
		path: path,
	}, nil
}

func (fs *LocalFs) OpenReader(path string) (io.ReadCloser, error) {
	p := strings.TrimSpace(path)
	if p == "" {
		return nil, fmt.Errorf("path is empty")
	}

	if filepath.IsAbs(p) {
		return nil, fmt.Errorf("path must be relative: %q", p)
	}

	if strings.HasPrefix(p, "~") {
		return nil, fmt.Errorf("path must be relative (no ~ expansion): %q", p)
	}

	root := strings.TrimSpace(string(fs.Root))
	if root == "" {
		root, _ = os.Getwd()
	}
	if root == "" {
		return nil, fmt.Errorf("root directory not specified")
	}

	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	// 連続した / を処理したり、.. を畳んだり、空文字を . にする
	full := filepath.Clean(filepath.Join(root, p))

	// root 以上 .. で遡られないようにチェック
	rel, err := filepath.Rel(root, full)
	if err != nil {
		return nil, fmt.Errorf("rel: %w", err)
	}
	if rel == "." {
		// Root itself - you probably don't want to open a directory as a reader.
		return nil, fmt.Errorf("path points to root directory, not a file: %q", p)
	}
	if strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." {
		return nil, fmt.Errorf("path escapes root: %q", p)
	}

	f, err := os.Open(full)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (fs *LocalFs) LoadAsJson(path string) (any, error) {
	var v any // ジェネリクスにしたい。 LocalFs[T] のような定義にしないといけない

	reader, err := fs.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	dec := json.NewDecoder(reader)
	dec.UseNumber()

	err = dec.Decode(&v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// LocalFile の File インターフェース実装

func (f *LocalFile) LoadAsJson() (any, error) {
	// Transform 設定に基づいて適切なデコーダーを選択
	reader, err := f.OpenReader()
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	switch f.fs.Transform {
	case "dotenv":
		// dotenv 形式として読み込み
		return decoders.DotenvToJson(reader)
	case "json", "":
		// デフォルトは JSON
		var v any
		dec := json.NewDecoder(reader)
		dec.UseNumber()
		if err := dec.Decode(&v); err != nil {
			return nil, err
		}
		return v, nil
	default:
		return nil, fmt.Errorf("unsupported transform format: %s", f.fs.Transform)
	}
}

func (f *LocalFile) OpenReader() (io.ReadCloser, error) {
	return f.fs.OpenReader(f.path)
}

func (fs *FsLocalFile) Exists() bool {
	_, err := os.Stat(fs.Path)
	return err == nil
}

func (fs *FsLocalFile) OpenWriter() (io.WriteCloser, error) {
	if fs.Path == "" {
		return nil, fmt.Errorf("An empty file name was given.: %s ", fs.Path)
	}

	if !fs.NoMkdir {
		if dir := filepath.Dir(fs.Path); dir != "." {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return nil, fmt.Errorf("mkdir %s: %w", dir, err)
			}
		}
	}

	f, err := os.Create(fs.Path) // truncate/create
	if err != nil {
		return nil, fmt.Errorf("open output %s: %w", fs.Path, err)
	}
	return f, nil
}
