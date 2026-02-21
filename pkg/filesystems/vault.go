package filesystems

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
)

// VaultConfig は HashiCorp Vault ファイルシステムの設定
type VaultConfig struct {
	// Addr は Vault サーバーのアドレス
	Addr string `yaml:"addr" doc:"Vault サーバーのアドレス" required:"true" example:"http://localhost:8200"`

	// Token は Vault 認証トークン
	Token string `yaml:"token" doc:"Vault 認証トークン" required:"true" example:"root"`

	// Namespace は Vault 名前空間（Enterprise 機能）
	Namespace string `yaml:"namespace" doc:"Vault 名前空間（Enterprise 機能）" required:"false" default:"" example:"admin"`

	// Mount は KV シークレットエンジンのマウントパス
	Mount string `yaml:"mount" doc:"KV シークレットエンジンのマウントパス" required:"true" default:"secret" example:"secret"`

	// KvVer は KV シークレットエンジンのバージョン（現在は 2 のみサポート）
	KvVer int `yaml:"kv_ver" doc:"KV シークレットエンジンのバージョン（現在は 2 のみサポート）" required:"false" default:"2"`

	// Version はシークレットのバージョン（0 = 最新）
	Version int `yaml:"version" doc:"取得するシークレットのバージョン（0 = 最新）" required:"false" default:"0"`

	// Field は取得する特定のフィールド名（省略時は全フィールド）
	Field string `yaml:"field" doc:"取得する特定のフィールド名（省略時は全フィールド）" required:"false" default:""`

	// Pretty は JSON 出力を整形するか
	Pretty bool `yaml:"pretty" doc:"JSON 出力を整形するか" required:"false" default:"false"`

	// Timeout はリクエストタイムアウト
	Timeout time.Duration `yaml:"timeout" doc:"リクエストタイムアウト" required:"false" default:"30s" example:"60s"`
}

type VaultFs struct {
	Ctx     context.Context
	Client  *vaultapi.KVv2
	Version int
	Timeout time.Duration
}

type VaultFsFile struct {
	fs   *VaultFs
	Path string
}

func GetVaultFs(parent context.Context, fs *VaultConfig) (*VaultFs, error) {
	cfg := vaultapi.DefaultConfig()
	// 環境変数からは読み込まない（明示的な設定のみ使用）

	cfg.Address = fs.Addr
	client, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("vault client: %w", err)
	}
	client.SetNamespace(fs.Namespace)
	client.SetToken(fs.Token)

	mount := strings.Trim(fs.Mount, "/")

	fs.KvVer = 2
	if fs.KvVer != 2 {
		return nil, fmt.Errorf("Unsupported Version: %v", fs.KvVer)
	}

	vault_client := client.KVv2(mount)
	fs2 := VaultFs{
		Ctx:     parent,
		Client:  vault_client,
		Timeout: fs.Timeout,
		Version: fs.Version,
	}
	return &fs2, nil
}

func (fs *VaultFs) GetFile(path string) (File, error) {
	return &VaultFsFile{
		fs:   fs,
		Path: path,
	}, nil
}

func (file *VaultFsFile) LoadAsJson() (any, error) {
	fs := file.fs
	secretPath := strings.Trim(file.Path, "/")
	ctx, cancel := context.WithTimeout(fs.Ctx, fs.Timeout)
	defer cancel()

	// Version == 0(latest) 以外許容しない
	if fs.Version != 0 {
		return nil, fmt.Errorf("version must be 0 (latest), got %d", fs.Version)
	}

	sec, err := fs.Client.Get(ctx, secretPath)
	if err != nil {
		return nil, err
	}

	if sec == nil || sec.Data == nil {
		return nil, fmt.Errorf("secret has no data (deleted or empty)")
	}
	return sec.Data, nil
}

// OpenReader は JSON エンコードされたストリームを返す
func (file *VaultFsFile) OpenReader() (io.ReadCloser, error) {
	data, err := file.LoadAsJson()
	if err != nil {
		return nil, err
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to encode JSON: %w", err)
	}

	return io.NopCloser(bytes.NewReader(jsonBytes)), nil
}
