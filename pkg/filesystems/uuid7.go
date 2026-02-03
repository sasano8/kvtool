// Package filesystems provides unified filesystem interfaces.
//
// # UUID7 Filesystem
//
// UUID7 ファイルシステムは、アクセスするたびに新しい UUID7 を生成する仮想ファイルシステムです。
// /dev/random のような生成系ソースとして機能します。
//
// ## Driver Name
//
// uuid7
//
// ## Use Cases
//
// - データベースの主キー生成
// - 一意な識別子の取得
// - 時間順序付きの UUID が必要な場合
// - テスト時の再現可能な UUID 生成（seed オプション）
// - モック用の固定 UUID（fixed オプション）
//
// ## Configuration
//
//   - seed: シード値（int64）。指定すると決定論的な UUID を生成
//   - fixed: 固定 UUID 文字列。指定すると常にこの値を返す（seed より優先）
//
// ## Notes
//
// - パスは無視されます（常に新しい UUID を生成）
// - UUID7 は時間順序付き（RFC 9562 準拠）
// - seed 指定時は UUID7 形式だが時間情報は含まない（テスト用途）
package filesystems

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"regexp"

	"github.com/google/uuid"
)

// UUID7Fs は UUID7 を生成する仮想ファイルシステム
// パスは無視され、アクセスするたびに新しい UUID7 を返す
type UUID7Fs struct {
	Ctx    context.Context
	config *UUID7FsConfig
	rng    *rand.Rand // seed 指定時に使用
}

// UUID7FsConfig は UUID7Fs の設定
type UUID7FsConfig struct {
	// Seed はシード値。指定すると決定論的な UUID を生成する（テスト用）
	Seed *int64 `yaml:"seed" doc:"シード値（指定すると決定論的な UUID を生成）" required:"false"`
	// Fixed は固定 UUID 文字列。指定すると常にこの値を返す（Seed より優先）
	Fixed string `yaml:"fixed" doc:"固定 UUID 文字列（常にこの値を返す）" required:"false" example:"0190a5e4-1234-7abc-8def-0123456789ab"`
}

// uuidRegex は UUID の形式を検証する
var uuidFormatRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// NewUUID7Fs は新しい UUID7Fs インスタンスを作成する
func NewUUID7Fs(ctx context.Context, config *UUID7FsConfig) (*UUID7Fs, error) {
	if config == nil {
		config = &UUID7FsConfig{}
	}

	// Fixed が指定されている場合、形式を検証
	if config.Fixed != "" {
		if !uuidFormatRegex.MatchString(config.Fixed) {
			return nil, fmt.Errorf("invalid UUID format for fixed value: %s", config.Fixed)
		}
	}

	fs := &UUID7Fs{
		Ctx:    ctx,
		config: config,
	}

	// Seed が指定されている場合、RNG を初期化
	if config.Seed != nil {
		fs.rng = rand.New(rand.NewSource(*config.Seed))
	}

	return fs, nil
}

// GetFile は Filesystem インターフェースを実装
// path は無視され、常に新しい UUID7 を生成するファイルを返す
func (fs *UUID7Fs) GetFile(path string) (File, error) {
	return &UUID7File{fs: fs}, nil
}

// UUID7File は UUID7 を生成するファイル
type UUID7File struct {
	fs *UUID7Fs
}

// LoadAsJson は新しい UUID7 を文字列として返す
func (f *UUID7File) LoadAsJson() (any, error) {
	id, err := f.generateUUID()
	if err != nil {
		return nil, err
	}
	return id, nil
}

// OpenReader は新しい UUID7 をストリームとして返す（改行なし）
func (f *UUID7File) OpenReader() (io.ReadCloser, error) {
	id, err := f.generateUUID()
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader([]byte(id))), nil
}

// generateUUID は設定に応じて UUID を生成する
func (f *UUID7File) generateUUID() (string, error) {
	// Fixed が指定されている場合、その値を返す（最優先）
	if f.fs.config.Fixed != "" {
		return f.fs.config.Fixed, nil
	}

	// Seed が指定されている場合、決定論的な UUID を生成
	if f.fs.rng != nil {
		return f.generateSeededUUID(), nil
	}

	// 通常の UUID7 を生成
	id, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

// generateSeededUUID はシードベースの決定論的な UUID を生成する
// UUID7 形式（バージョン 7、バリアント RFC 4122）を維持
func (f *UUID7File) generateSeededUUID() string {
	var buf [16]byte

	// ランダムバイトを生成
	binary.BigEndian.PutUint64(buf[0:8], f.fs.rng.Uint64())
	binary.BigEndian.PutUint64(buf[8:16], f.fs.rng.Uint64())

	// バージョン 7 を設定（4ビット）
	buf[6] = (buf[6] & 0x0f) | 0x70

	// バリアント RFC 4122 を設定（2ビット）
	buf[8] = (buf[8] & 0x3f) | 0x80

	// UUID 形式の文字列に変換
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		buf[0:4],
		buf[4:6],
		buf[6:8],
		buf[8:10],
		buf[10:16])
}

// コンパイル時にインターフェース実装を確認
var _ Filesystem = (*UUID7Fs)(nil)
var _ File = (*UUID7File)(nil)
