// Package server provides HTTP server for kvtool stores.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/sasano8/kvtool/pkg/config"
	"github.com/sasano8/kvtool/pkg/filesystems"
)

// Server は kvtool の HTTP サーバー
type Server struct {
	config     *config.KvtoolConfig
	factory    *filesystems.FilesystemFactory
	filesCache map[string]filesystems.Filesystem
}

// NewServer は新しいサーバーインスタンスを作成する
func NewServer(ctx context.Context, cfg *config.KvtoolConfig) *Server {
	return &Server{
		config:     cfg,
		factory:    filesystems.NewFilesystemFactory(ctx),
		filesCache: make(map[string]filesystems.Filesystem),
	}
}

// URL パターン:
// /v1/namespaces/{namespace}/kv/{store}/{path...}
// /v1/namespaces/{namespace}/storage/{store}/{path...}

var (
	// /v1/namespaces/{namespace}/kv/{store}/{path...}
	nsKVPattern = regexp.MustCompile(`^/v1/namespaces/([^/]+)/kv/(.+)$`)
	// /v1/namespaces/{namespace}/storage/{store}/{path...}
	nsStoragePattern = regexp.MustCompile(`^/v1/namespaces/([^/]+)/storage/(.+)$`)
)

// Handler は HTTP ハンドラーを返す
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// /v1/namespaces/{namespace}/...
	mux.HandleFunc("/v1/namespaces/", s.handleNamespaced)

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	return mux
}

// handleNamespaced は /v1/namespaces/{namespace}/... を処理する
func (s *Server) handleNamespaced(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Path

	// /v1/namespaces/{namespace}/kv/{store}/{path...}
	if matches := nsKVPattern.FindStringSubmatch(path); matches != nil {
		namespace := matches[1]
		rest := matches[2]
		s.serveKV(w, namespace, rest)
		return
	}

	// /v1/namespaces/{namespace}/storage/{store}/{path...}
	if matches := nsStoragePattern.FindStringSubmatch(path); matches != nil {
		namespace := matches[1]
		rest := matches[2]
		s.serveStorage(w, namespace, rest)
		return
	}

	http.Error(w, "invalid path", http.StatusBadRequest)
}

// serveKV は JSON レスポンスを返す
func (s *Server) serveKV(w http.ResponseWriter, namespace, storePath string) {
	storeName, path, err := parseStorePath(storePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// ファイルシステムを取得
	fs, err := s.getFilesystem(namespace, storeName)
	if err != nil {
		http.Error(w, fmt.Sprintf("store not found: %s", err), http.StatusNotFound)
		return
	}

	// ファイル取得
	file, err := fs.GetFile(path)
	if err != nil {
		http.Error(w, fmt.Sprintf("file not found: %s", err), http.StatusNotFound)
		return
	}

	// JSON として読み込み
	data, err := file.LoadAsJson()
	if err != nil {
		if isNotFoundError(err) {
			http.Error(w, fmt.Sprintf("file not found: %s", err), http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("failed to load as JSON: %s", err), http.StatusInternalServerError)
		return
	}

	// JSON レスポンス
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode response: %s", err), http.StatusInternalServerError)
		return
	}
}

// serveStorage は生データを返す
func (s *Server) serveStorage(w http.ResponseWriter, namespace, storePath string) {
	storeName, path, err := parseStorePath(storePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// ファイルシステムを取得
	fs, err := s.getFilesystem(namespace, storeName)
	if err != nil {
		http.Error(w, fmt.Sprintf("store not found: %s", err), http.StatusNotFound)
		return
	}

	// ファイル取得
	file, err := fs.GetFile(path)
	if err != nil {
		http.Error(w, fmt.Sprintf("file not found: %s", err), http.StatusNotFound)
		return
	}

	// ストリームとして読み込み
	reader, err := file.OpenReader()
	if err != nil {
		if isNotFoundError(err) {
			http.Error(w, fmt.Sprintf("file not found: %s", err), http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("failed to open file: %s", err), http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	// バイナリレスポンス
	w.Header().Set("Content-Type", "application/octet-stream")
	io.Copy(w, reader)
}

// parseStorePath は store/path 形式をパースする
func parseStorePath(storePath string) (storeName, path string, err error) {
	parts := strings.SplitN(storePath, "/", 2)
	storeName = parts[0]
	if storeName == "" {
		return "", "", fmt.Errorf("store name is required")
	}

	if len(parts) > 1 {
		path = parts[1]
	}

	return storeName, path, nil
}

// isNotFoundError はエラーがファイル不在を示すか判定する
func isNotFoundError(err error) bool {
	return os.IsNotExist(err) ||
		strings.Contains(err.Error(), "no such file") ||
		strings.Contains(err.Error(), "not found")
}

// getFilesystem は指定された store のファイルシステムを取得する
func (s *Server) getFilesystem(namespace, storeName string) (filesystems.Filesystem, error) {
	cacheKey := fmt.Sprintf("%s/%s", namespace, storeName)

	// キャッシュを確認
	if fs, ok := s.filesCache[cacheKey]; ok {
		return fs, nil
	}

	// 設定から store 情報を取得
	storeInfo, err := s.config.GetStore(namespace, storeName)
	if err != nil {
		return nil, err
	}

	// ファイルシステムを作成
	fsStoreInfo := &filesystems.StoreInfo{
		Driver: storeInfo.Driver,
		Args:   storeInfo.Args,
	}
	fs, err := s.factory.Create(fsStoreInfo)
	if err != nil {
		return nil, err
	}

	// キャッシュに保存
	s.filesCache[cacheKey] = fs

	return fs, nil
}
