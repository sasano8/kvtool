package kvfs

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"path"
	"strings"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
)

type VaultFS struct {
	client  *vaultapi.Client
	mount   string        // 例: "secret"
	prefix  string        // 例: "app" (任意)
	timeout time.Duration // 例: 10s
}

// NewVaultFS: client は外で作って渡すのがおすすめ（token/addr等の責務分離）
func NewVaultFS(client *vaultapi.Client, mount string, opts ...func(*VaultFS)) *VaultFS {
	v := &VaultFS{
		client:  client,
		mount:   strings.Trim(mount, "/"),
		timeout: 10 * time.Second,
	}
	for _, f := range opts {
		f(v)
	}
	return v
}

func WithPrefix(p string) func(*VaultFS)         { return func(v *VaultFS) { v.prefix = strings.Trim(p, "/") } }
func WithTimeout(d time.Duration) func(*VaultFS) { return func(v *VaultFS) { v.timeout = d } }

// Open implements fs.FS
// name は "foo/bar" みたいな相対パス想定。
// 返す内容は「Vaultの data を JSON にしたもの」（＝KVとして読みやすい）
func (v *VaultFS) Open(name string) (fs.File, error) {
	if v.client == nil {
		return nil, fs.ErrInvalid
	}
	name = strings.TrimSpace(name)
	if !fs.ValidPath(name) {
		return nil, fs.ErrInvalid
	}

	secretPath := v.join(name) // prefix + name
	ctx, cancel := context.WithTimeout(context.Background(), v.timeout)
	defer cancel()

	// KV v2 前提（必要なら v1/v2 切替を追加）
	sec, err := v.client.KVv2(v.mount).Get(ctx, secretPath)
	if err != nil {
		// “存在しない”を fs.ErrNotExist に寄せると fs.ReadFile 等と相性が良い
		if looksNotFound(err) {
			return nil, fs.ErrNotExist
		}
		return nil, err
	}
	if sec == nil || sec.Data == nil {
		return nil, fs.ErrNotExist
	}

	b, err := json.Marshal(sec.Data) // メタデータ捨てて data だけ
	if err != nil {
		return nil, err
	}
	b = append(b, '\n')

	return newMemFile(path.Base(name), b), nil
}

// Exists: ファイルのみ対象（Vaultのキー=ファイル扱い）
func (v *VaultFS) Exists(name string) (bool, error) {
	f, err := v.Open(name)
	if err == nil {
		_ = f.Close()
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func (v *VaultFS) join(name string) string {
	if v.prefix == "" {
		return name
	}
	return v.prefix + "/" + name
}

// --- memFile: Vaultの値を fs.File として返すための擬似ファイル ---

type memFile struct {
	*bytes.Reader
	name string
	size int64
}

func newMemFile(name string, b []byte) *memFile {
	return &memFile{
		Reader: bytes.NewReader(b),
		name:   name,
		size:   int64(len(b)),
	}
}

func (m *memFile) Close() error { return nil }

func (m *memFile) Stat() (fs.FileInfo, error) { return memInfo{name: m.name, size: m.size}, nil }

type memInfo struct {
	name string
	size int64
}

func (i memInfo) Name() string       { return i.name }
func (i memInfo) Size() int64        { return i.size }
func (i memInfo) Mode() fs.FileMode  { return 0o444 } // read-only file
func (i memInfo) ModTime() time.Time { return time.Time{} }
func (i memInfo) IsDir() bool        { return false }
func (i memInfo) Sys() any           { return nil }

// Vaultの not found をどう判定するかは環境で揺れるので雑にしてる。
// 本番は err の型や status code を見る実装に寄せるのが良い。
func looksNotFound(err error) bool {
	// vault/api は内部で *api.ResponseError を返すことがある
	// ここは簡易判定（必要なら厳密化）
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "404") || strings.Contains(s, "not found") || strings.Contains(s, "permission denied")
}

// コンパイル時に満たしているか確認（任意）
var _ fs.FS = (*VaultFS)(nil)
var _ io.Reader = (*memFile)(nil) // bytes.Readerがあるので OK
