package filesystems

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
)

type FilesystemConfig struct {
	Type   string         `json:"type"`
	Kwargs map[string]any `json:"kwargs"`
}

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

type VaultFile struct {
	fs   *FilesystemConfig
	Path string
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

func (fs *VaultFs) GetFile(path string) (VaultFsFile, error) {
	return VaultFsFile{
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
		return nil, fmt.Errorf("Error version.")
		// fmt.Printf("aaa")
		// sec, err = vault_client.GetVersion(ctx, secretPath, fs.Version)
	}

	sec, err := fs.Client.Get(ctx, secretPath)
	if err != nil {
		return nil, err
	}

	// KVv2 helper が古いVaultで metadata 互換が崩れるケースがあるので、
	// 失敗したら HTTP API の /data を直接叩くフォールバックに落とす（data だけ取る）
	// if err != nil {
	// 	fmt.Printf("asdfasd")
	// 	raw, err2 := readVaultKVv2Raw(ctx, client, mount, secretPath, fs.Version)
	// 	if err2 == nil {
	// 		return raw, nil
	// 	}
	// 	return nil, err
	// }

	if sec == nil || sec.Data == nil {
		return nil, fmt.Errorf("secret has no data (deleted or empty)")
	}
	return sec.Data, nil
}

// KV v2 の HTTP API を直接叩いて data だけ抜くフォールバック
// v2 の Read API は /<mount>/data/<path> を使う :contentReference[oaicite:1]{index=1}
func readVaultKVv2Raw(ctx context.Context, client *vaultapi.Client, mount, secretPath string, version int) (any, error) {
	apiPath := fmt.Sprintf("%s/data/%s", strings.Trim(mount, "/"), strings.Trim(secretPath, "/"))
	q := map[string][]string{}
	if version > 0 {
		q["version"] = []string{strconv.Itoa(version)}
	}

	sec, err := client.Logical().ReadWithDataWithContext(ctx, apiPath, q)
	if err != nil {
		return nil, err
	}
	if sec == nil || sec.Data == nil {
		return nil, fmt.Errorf("secret not found")
	}

	// KV v2 のレスポンスは Data["data"] に本体が入る :contentReference[oaicite:2]{index=2}
	rawData, ok := sec.Data["data"].(map[string]interface{})
	if !ok || rawData == nil {
		return nil, fmt.Errorf("unexpected KV v2 response format at %s", apiPath)
	}
	// fmt.Printf("%v", rawData)
	return rawData, nil
}
