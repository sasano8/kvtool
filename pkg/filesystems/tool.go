// Package filesystems provides unified filesystem interfaces.
//
// # Tool Filesystem
//
// Tool ファイルシステムは、UUID、タイムスタンプ、乱数などの値を動的に生成します。
//
// ## Driver Name
//
// tool
//
// ## Use Cases
//
// - テンプレートでの動的値生成
// - シークレット生成
// - 一意な識別子の生成
//
// ## Available Paths
//
// - uuid7: UUID v7（時刻ベース）
// - uuid4: UUID v4（ランダム）
// - now: 現在時刻（RFC3339 形式）
// - timestamp: Unix タイムスタンプ
// - random/hex/{bytes}: 16進数ランダム値
// - random/base64/{bytes}: Base64 ランダム値
// - password: 16文字のランダムパスワード
package filesystems

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ToolFsConfig は Tool ファイルシステムの設定
type ToolFsConfig struct {
	// 設定項目なし（将来の拡張用）
}

// ToolFs は動的値生成ファイルシステム
type ToolFs struct {
	ctx context.Context
}

// NewToolFs は新しい Tool ファイルシステムを作成
func NewToolFs(ctx context.Context, config *ToolFsConfig) (*ToolFs, error) {
	return &ToolFs{
		ctx: ctx,
	}, nil
}

// GetFile は指定されたパスの値を生成
func (fs *ToolFs) GetFile(path string) (File, error) {
	// パスの検証
	if strings.HasPrefix(path, "/") || strings.HasPrefix(path, "..") {
		return nil, fmt.Errorf("path must be relative: %s", path)
	}

	return &ToolFile{
		fs:   fs,
		path: path,
	}, nil
}

// ToolFile は Tool ファイルシステムのファイル
type ToolFile struct {
	fs   *ToolFs
	path string
}

// generateValue はパスに基づいて値を生成（値をそのまま返す）
func (f *ToolFile) generateValue() (any, error) {
	parts := strings.Split(f.path, "/")
	toolName := parts[0]

	switch toolName {
	case "uuid7":
		id, err := uuid.NewV7()
		if err != nil {
			return nil, err
		}
		return id.String(), nil

	case "uuid4":
		return uuid.New().String(), nil

	case "now":
		return time.Now().UTC().Format(time.RFC3339), nil

	case "timestamp":
		// int64 として返す
		return time.Now().Unix(), nil

	case "random":
		if len(parts) < 3 {
			return nil, fmt.Errorf("usage: random/{hex|base64}/{bytes}")
		}
		format := parts[1]
		byteCount, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil, fmt.Errorf("invalid byte count: %s", parts[2])
		}
		if byteCount <= 0 || byteCount > 1024 {
			return nil, fmt.Errorf("bytes must be between 1 and 1024")
		}

		data := make([]byte, byteCount)
		if _, err := rand.Read(data); err != nil {
			return nil, err
		}

		switch format {
		case "hex":
			return hex.EncodeToString(data), nil
		case "base64":
			return base64.StdEncoding.EncodeToString(data), nil
		default:
			return nil, fmt.Errorf("unknown format: %s (use hex or base64)", format)
		}

	case "password":
		const length = 16
		const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"

		password := make([]byte, length)
		randomBytes := make([]byte, length)
		if _, err := rand.Read(randomBytes); err != nil {
			return nil, err
		}

		for i := 0; i < length; i++ {
			password[i] = charset[int(randomBytes[i])%len(charset)]
		}
		return string(password), nil

	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}

// LoadAsJson は値をそのまま返す（ラップしない）
func (f *ToolFile) LoadAsJson() (any, error) {
	return f.generateValue()
}

// OpenReader は JSON エンコードされたストリームを返す
func (f *ToolFile) OpenReader() (io.ReadCloser, error) {
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
