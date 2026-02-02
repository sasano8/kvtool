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

type VaultConfig struct {
	Addr      string        `json:"addr"`
	Token     string        `json:"token"`
	Namespace string        `json:"namespace"`
	Mount     string        `json:"mount"`
	KvVer     int           `json:"kv_ver"`
	Version   int           `json:"version"`
	Field     string        `json:"field"`
	Pretty    bool          `json:"pretty"`
	Timeout   time.Duration `json:"timeout"`
}

type VaultFs struct {
	Ctx     context.Context
	Client  *vaultapi.KVv2
	Version int
	Timeout time.Duration
	Root    string
}

type VaultFsFile struct {
	fs   *VaultFs
	Path string
}

func GetVaultFs(parent context.Context, fs *VaultConfig) (*VaultFs, error) {
	cfg := vaultapi.DefaultConfig()
	// VAULT_ADDR や TLS 系環境変数（VAULT_CACERT等）を反映
	_ = cfg.ReadEnvironment()

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
