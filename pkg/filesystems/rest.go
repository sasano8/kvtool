// Package filesystems provides unified filesystem interfaces.
//
// # REST Filesystem
//
// REST ファイルシステムは、任意の REST API からデータを読み込みます。
//
// ## Driver Name
//
// rest
//
// ## Use Cases
//
// - Kubernetes API からの設定取得
// - 外部 REST API からのデータ取得
//
// ## Security
//
// - TLS 証明書検証
// - Bearer/Basic 認証サポート
package filesystems

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

// RestFsConfig は REST ファイルシステムの設定
type RestFsConfig struct {
	BaseURL   string        `yaml:"base_url" doc:"ベース URL" required:"true" example:"https://api.example.com"`
	Root      string        `yaml:"root" doc:"ルートパス" required:"false" default:"" example:"/api/v1/data"`
	AuthType  string        `yaml:"auth_type" doc:"認証タイプ（bearer, basic）" required:"false" example:"bearer"`
	Token     string        `yaml:"token" doc:"Bearer トークン" required:"false"`
	TokenFile string        `yaml:"token_file" doc:"Bearer トークンファイルパス" required:"false" example:"/var/run/secrets/token"`
	Username  string        `yaml:"username" doc:"Basic 認証ユーザー名" required:"false"`
	Password  string        `yaml:"password" doc:"Basic 認証パスワード" required:"false"`
	CAFile    string        `yaml:"ca_file" doc:"CA 証明書ファイルパス" required:"false"`
	Insecure  bool          `yaml:"insecure" doc:"TLS 証明書検証をスキップ" required:"false" default:"false"`
	Timeout   time.Duration `yaml:"timeout" doc:"リクエストタイムアウト" required:"false" default:"30s"`
}

// RestFs は REST API ファイルシステム
type RestFs struct {
	ctx    context.Context
	config *RestFsConfig
	client *http.Client
}

// NewRestFs は新しい REST ファイルシステムを作成
func NewRestFs(ctx context.Context, config *RestFsConfig) (*RestFs, error) {
	if config.BaseURL == "" {
		return nil, fmt.Errorf("base_url is required")
	}

	// デフォルト値
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	// HTTP クライアントの作成
	client, err := createHTTPClient(config)
	if err != nil {
		return nil, err
	}

	return &RestFs{
		ctx:    ctx,
		config: config,
		client: client,
	}, nil
}

func createHTTPClient(config *RestFsConfig) (*http.Client, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: config.Insecure,
	}

	// CA 証明書の読み込み
	if config.CAFile != "" {
		caCert, err := os.ReadFile(config.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA file: %w", err)
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		tlsConfig.RootCAs = caCertPool
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}, nil
}

// GetFile はファイルを取得する
func (fs *RestFs) GetFile(filePath string) (File, error) {
	// パスの検証
	if strings.HasPrefix(filePath, "/") || strings.HasPrefix(filePath, "..") {
		return nil, fmt.Errorf("path must be relative: %s", filePath)
	}

	return &RestFile{
		fs:   fs,
		path: filePath,
	}, nil
}

func (fs *RestFs) buildURL(filePath string) string {
	// base_url + root + path
	url := strings.TrimSuffix(fs.config.BaseURL, "/")
	if fs.config.Root != "" {
		root := fs.config.Root
		if !strings.HasPrefix(root, "/") {
			root = "/" + root
		}
		url += strings.TrimSuffix(root, "/")
	}
	if filePath != "" {
		url += "/" + filePath
	}
	return url
}

func (fs *RestFs) getAuthToken() (string, error) {
	if fs.config.Token != "" {
		return fs.config.Token, nil
	}
	if fs.config.TokenFile != "" {
		data, err := os.ReadFile(fs.config.TokenFile)
		if err != nil {
			return "", fmt.Errorf("failed to read token file: %w", err)
		}
		return strings.TrimSpace(string(data)), nil
	}
	return "", nil
}

func (fs *RestFs) doRequest(filePath string) (*http.Response, error) {
	url := fs.buildURL(filePath)

	req, err := http.NewRequestWithContext(fs.ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// 認証ヘッダーの設定
	switch fs.config.AuthType {
	case "bearer":
		token, err := fs.getAuthToken()
		if err != nil {
			return nil, err
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	case "basic":
		if fs.config.Username != "" {
			req.SetBasicAuth(fs.config.Username, fs.config.Password)
		}
	}

	resp, err := fs.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("not found: %s", url)
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, url)
	}

	return resp, nil
}

// RestFile は REST API ファイル
type RestFile struct {
	fs   *RestFs
	path string
}

// LoadAsJson は JSON としてロードする
func (f *RestFile) LoadAsJson() (any, error) {
	resp, err := f.fs.doRequest(f.path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	return result, nil
}

// OpenReader はリーダーを返す
func (f *RestFile) OpenReader() (io.ReadCloser, error) {
	resp, err := f.fs.doRequest(f.path)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

// Path はファイルパスを返す
func (f *RestFile) Path() string {
	return path.Join(f.fs.config.Root, f.path)
}

// LoadAsBytes はバイト列として読み込む
func (f *RestFile) LoadAsBytes() ([]byte, error) {
	reader, err := f.OpenReader()
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
