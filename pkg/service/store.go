package service

import (
	"context"
	"fmt"

	"github.com/sasano8/kvtool/pkg/config"
	"github.com/sasano8/kvtool/pkg/filesystems"
)

// StoreService はストア操作のビジネスロジックを提供する
type StoreService interface {
	// Get retrieves content from a store
	Get(ctx context.Context, opts GetOptions) (any, error)
}

// GetOptions は Get メソッドのオプション
type GetOptions struct {
	ConfigPath string // 設定ファイルのパス
	Namespace  string // 使用する namespace
	StorePath  string // "store_name/file_path" 形式
}

// storeService は StoreService の実装
type storeService struct {
	configLoader ConfigLoader
	fsFactory    FilesystemFactory
}

// ConfigLoader は設定ファイルのロードを担当
type ConfigLoader interface {
	LoadConfig(path string) (*config.KvtoolConfig, error)
}

// FilesystemFactory は Filesystem の作成を担当
type FilesystemFactory interface {
	Create(ctx context.Context, storeInfo *config.StoreInfo) (filesystems.Filesystem, error)
}

// NewStoreService creates a new StoreService
func NewStoreService() StoreService {
	return &storeService{
		configLoader: &defaultConfigLoader{},
		fsFactory:    &defaultFilesystemFactory{},
	}
}

// Get implements StoreService.Get
func (s *storeService) Get(ctx context.Context, opts GetOptions) (any, error) {
	// 1. Parse store path
	parsed, err := config.ParseStorePath(opts.StorePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse store path: %w", err)
	}

	// 2. Load config
	cfg, err := s.configLoader.LoadConfig(opts.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// 3. Get store info
	storeInfo, err := cfg.GetStore(opts.Namespace, parsed.StoreName)
	if err != nil {
		return nil, fmt.Errorf("failed to get store info: %w", err)
	}

	// 4. Create filesystem
	fs, err := s.fsFactory.Create(ctx, storeInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to create filesystem: %w", err)
	}

	// 5. Get file content
	content, err := getFileContent(fs, storeInfo, parsed.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file content: %w", err)
	}

	return content, nil
}

// getFileContent retrieves file content from a filesystem
func getFileContent(fs filesystems.Filesystem, storeInfo *config.StoreInfo, filePath string) (interface{}, error) {
	// mount.dir が設定されている場合、パスにプレフィックスを付与
	if storeInfo.Mount != nil && storeInfo.Mount.Dir != nil && *storeInfo.Mount.Dir != "" {
		filePath = *storeInfo.Mount.Dir + "/" + filePath
	}

	// ファイルを取得して JSON として読み込む
	// Transform は LocalFs 内で自動的に適用される
	file, err := fs.GetFile(filePath)
	if err != nil {
		return nil, err
	}

	return file.LoadAsJson()
}

// defaultConfigLoader は ConfigLoader のデフォルト実装
type defaultConfigLoader struct{}

func (l *defaultConfigLoader) LoadConfig(path string) (*config.KvtoolConfig, error) {
	return config.LoadConfigAuto(path)
}

// defaultFilesystemFactory は FilesystemFactory のデフォルト実装
type defaultFilesystemFactory struct{}

func (f *defaultFilesystemFactory) Create(ctx context.Context, storeInfo *config.StoreInfo) (filesystems.Filesystem, error) {
	factory := filesystems.NewFilesystemFactory(ctx)

	// Convert config.StoreInfo to filesystems.StoreInfo
	fsStoreInfo := &filesystems.StoreInfo{
		Driver: storeInfo.Driver,
		Args:   storeInfo.Args,
	}

	// Propagate context info
	if storeInfo.Context != nil {
		fsStoreInfo.Context = &filesystems.ContextInfo{
			Timeout: storeInfo.Context.Timeout,
		}
	}

	return factory.Create(fsStoreInfo)
}
