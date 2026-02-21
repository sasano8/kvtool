// Package filesystems provides unified filesystem interfaces.
//
// # Git Filesystem
//
// Git ファイルシステムは、Git リポジトリを clone/pull し、ローカルキャッシュ経由で読み込みます。
// 認証は git コマンドの設定（.gitconfig）に従います。
//
// ## Driver Name
//
// git
//
// ## Use Cases
//
// - Git リポジトリに格納された設定ファイルの読み込み
// - バージョン管理された設定の参照
//
// ## Security
//
// - パストラバーサル攻撃の防止
// - キャッシュディレクトリの自動管理
package filesystems

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// GitFsConfig は Git ファイルシステムの設定
type GitFsConfig struct {
	// URL は Git リポジトリの URL（必須）
	URL string `yaml:"url" doc:"Git リポジトリ URL" required:"true" example:"https://github.com/example/config.git"`

	// Ref はブランチ・タグ名（オプション、デフォルト: main）
	Ref string `yaml:"ref" doc:"ブランチまたはタグ名" required:"false" default:"main" example:"main"`

	// Timeout はコマンドタイムアウト（オプション）
	Timeout time.Duration `yaml:"timeout" doc:"git コマンドのタイムアウト" required:"false" default:"60s" example:"120s"`
}

// GitFs は Git リポジトリベースのファイルシステム
type GitFs struct {
	ctx      context.Context
	cacheDir string
	url      string
	ref      string
	timeout  time.Duration
	inner    *LocalFs
}

// NewGitFs は新しい Git ファイルシステムを作成
func NewGitFs(ctx context.Context, config *GitFsConfig) (*GitFs, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("url is required")
	}

	ref := config.Ref
	if ref == "" {
		ref = "main"
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 60 * time.Second
	}

	cacheDir := gitCacheDir(config.URL, ref)

	fs := &GitFs{
		ctx:      ctx,
		cacheDir: cacheDir,
		url:      config.URL,
		ref:      ref,
		timeout:  timeout,
	}

	// git ls-remote で到達可能か確認
	if err := fs.checkReachable(); err != nil {
		return nil, fmt.Errorf("git repository not reachable: %w", err)
	}

	return fs, nil
}

// gitCacheDir はキャッシュディレクトリのパスを算出する
func gitCacheDir(url, ref string) string {
	h := sha256.Sum256([]byte(url + "@" + ref))
	hex := fmt.Sprintf("%x", h)
	short := hex[:12]

	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = os.TempDir()
	}

	return filepath.Join(homeDir, ".cache", "kvtool", "git", short)
}

// GitCacheDir は外部からキャッシュディレクトリのパスを取得するためのメソッド（テスト用）
func GitCacheDir(url, ref string) string {
	return gitCacheDir(url, ref)
}

// checkReachable は git ls-remote でリポジトリへの到達可能性を確認する
func (fs *GitFs) checkReachable() error {
	ctx, cancel := context.WithTimeout(fs.ctx, fs.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "ls-remote", "--exit-code", fs.url)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git ls-remote failed for %s: %w", fs.url, err)
	}

	return nil
}

// Sync はリモートリポジトリからローカルキャッシュを同期する
func (fs *GitFs) Sync() error {
	if _, err := os.Stat(filepath.Join(fs.cacheDir, ".git")); os.IsNotExist(err) {
		// 初回: clone
		if err := fs.cloneRepo(); err != nil {
			return fmt.Errorf("failed to clone: %w", err)
		}
	} else {
		// 既存: fetch + checkout
		if err := fs.fetchAndCheckout(); err != nil {
			return fmt.Errorf("failed to fetch: %w", err)
		}
	}

	// inner LocalFs を設定
	localFs, err := GetLocalFs(fs.ctx, &LocalFsConfig{
		Root:    fs.cacheDir,
		Timeout: fs.timeout,
	})
	if err != nil {
		return fmt.Errorf("failed to create local fs for cache: %w", err)
	}
	fs.inner = localFs

	return nil
}

func (fs *GitFs) cloneRepo() error {
	ctx, cancel := context.WithTimeout(fs.ctx, fs.timeout)
	defer cancel()

	// 親ディレクトリを作成
	if err := os.MkdirAll(filepath.Dir(fs.cacheDir), 0755); err != nil {
		return fmt.Errorf("failed to create cache parent dir: %w", err)
	}

	cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", "--branch", fs.ref, fs.url, fs.cacheDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %s: %w", strings.TrimSpace(string(output)), err)
	}

	return nil
}

func (fs *GitFs) fetchAndCheckout() error {
	ctx, cancel := context.WithTimeout(fs.ctx, fs.timeout)
	defer cancel()

	// fetch
	fetchCmd := exec.CommandContext(ctx, "git", "-C", fs.cacheDir, "fetch", "origin", fs.ref, "--depth", "1")
	output, err := fetchCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git fetch failed: %s: %w", strings.TrimSpace(string(output)), err)
	}

	// checkout FETCH_HEAD
	checkoutCmd := exec.CommandContext(ctx, "git", "-C", fs.cacheDir, "checkout", "FETCH_HEAD")
	output, err = checkoutCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git checkout failed: %s: %w", strings.TrimSpace(string(output)), err)
	}

	return nil
}

// GetFile は Filesystem インターフェースを実装
func (fs *GitFs) GetFile(path string) (File, error) {
	// inner が nil なら自動 Sync
	if fs.inner == nil {
		if err := fs.Sync(); err != nil {
			return nil, fmt.Errorf("auto sync failed: %w", err)
		}
	}

	return fs.inner.GetFile(path)
}
