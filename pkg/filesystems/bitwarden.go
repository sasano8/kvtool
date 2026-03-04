// Package filesystems provides unified filesystem interfaces.
//
// # Bitwarden Secrets Manager Filesystem
//
// Bitwarden ファイルシステムは、Bitwarden Secrets Manager から
// 公式 Go SDK 経由でシークレットを読み込みます。
//
// ## Driver Name
//
// bitwarden
//
// ## Use Cases
//
// - CI/CD パイプラインでのシークレット取得
// - アプリケーション設定の安全な配布
//
// ## Security
//
// - アクセストークンによる認証
// - SDK がエンドツーエンド暗号化を処理
//
// ## Build Requirements
//
// CGO が必要（Rust FFI）。ビルド前に: sudo apt install musl-tools && go env -w CGO_ENABLED=1 CC=musl-gcc
package filesystems

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	sdk "github.com/bitwarden/sdk-go/v2"
)

// BitwardenFsConfig は Bitwarden Secrets Manager ファイルシステムの設定
type BitwardenFsConfig struct {
	// AccessToken は Bitwarden マシンアカウントのアクセストークン
	AccessToken string `yaml:"access_token" doc:"Bitwarden アクセストークン" required:"true" example:"0.48c78342-..."`

	// OrganizationID は Bitwarden 組織 ID
	OrganizationID string `yaml:"organization_id" doc:"Bitwarden 組織 ID" required:"true" example:"a1b2c3d4-..."`

	// ProjectID はフィルタ対象のプロジェクト ID（省略時は全シークレット）
	ProjectID string `yaml:"project_id" doc:"フィルタ対象のプロジェクト ID" required:"false" default:"" example:"e325ea69-..."`
}

// BitwardenFs は Bitwarden Secrets Manager をファイルシステムとして扱う
type BitwardenFs struct {
	client         sdk.BitwardenClientInterface
	organizationID string
	projectID      string
}

// BitwardenFile は Bitwarden シークレットを表すファイル
type BitwardenFile struct {
	fs   *BitwardenFs
	path string // シークレットの key 名
}

// NewBitwardenFs は Bitwarden ファイルシステムを生成する
func NewBitwardenFs(config *BitwardenFsConfig) (*BitwardenFs, error) {
	if config.AccessToken == "" {
		return nil, fmt.Errorf("bitwarden: access_token is required")
	}
	if config.OrganizationID == "" {
		return nil, fmt.Errorf("bitwarden: organization_id is required")
	}

	client, err := sdk.NewBitwardenClient(nil, nil)
	if err != nil {
		return nil, fmt.Errorf("bitwarden: failed to create client: %w", err)
	}

	if err := client.AccessTokenLogin(config.AccessToken, nil); err != nil {
		client.Close()
		return nil, fmt.Errorf("bitwarden: authentication failed: %w", err)
	}

	return &BitwardenFs{
		client:         client,
		organizationID: config.OrganizationID,
		projectID:      config.ProjectID,
	}, nil
}

// Close は SDK クライアントのリソースを解放する
func (fs *BitwardenFs) Close() {
	if fs.client != nil {
		fs.client.Close()
		fs.client = nil
	}
}

// GetFile は Filesystem インターフェースを実装
// path はシークレットの key 名を指定する
func (fs *BitwardenFs) GetFile(path string) (File, error) {
	return &BitwardenFile{
		fs:   fs,
		path: path,
	}, nil
}

// LoadAsJson はシークレットの value を JSON として返す
// path が空の場合は全シークレットを map[key]value として返す
// path が指定された場合は該当シークレットの value を返す（JSON パースを試み、失敗なら文字列）
func (f *BitwardenFile) LoadAsJson() (any, error) {
	secrets := f.fs.client.Secrets()

	// シークレット一覧を取得（ID + Key のみ）
	identifiers, err := secrets.List(f.fs.organizationID)
	if err != nil {
		return nil, fmt.Errorf("bitwarden: failed to list secrets: %w", err)
	}

	if len(identifiers.Data) == 0 {
		if f.path == "" {
			return map[string]any{}, nil
		}
		return nil, fmt.Errorf("bitwarden: secret with key %q not found", f.path)
	}

	// path が空: 全シークレットを取得
	if f.path == "" {
		ids := make([]string, len(identifiers.Data))
		for i, id := range identifiers.Data {
			ids[i] = id.ID
		}

		fullSecrets, err := secrets.GetByIDS(ids)
		if err != nil {
			return nil, fmt.Errorf("bitwarden: failed to get secrets: %w", err)
		}

		result := make(map[string]any, len(fullSecrets.Data))
		for _, s := range fullSecrets.Data {
			result[s.Key] = s.Value
		}
		return result, nil
	}

	// path 指定: key 名で検索
	for _, id := range identifiers.Data {
		if id.Key == f.path {
			secret, err := secrets.Get(id.ID)
			if err != nil {
				return nil, fmt.Errorf("bitwarden: failed to get secret %q: %w", f.path, err)
			}
			return tryParseJSON(secret.Value), nil
		}
	}

	return nil, fmt.Errorf("bitwarden: secret with key %q not found", f.path)
}

// OpenReader は JSON エンコードされたストリームを返す
func (f *BitwardenFile) OpenReader() (io.ReadCloser, error) {
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

// tryParseJSON は文字列を JSON としてパースを試み、失敗なら文字列をそのまま返す
func tryParseJSON(s string) any {
	var result any
	if err := json.Unmarshal([]byte(s), &result); err != nil {
		return s
	}
	return result
}
