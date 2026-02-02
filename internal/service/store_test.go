package service

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/sasano8/kvtool/pkg/filesystems"
	"github.com/sasano8/kvtool/internal/config"
	"github.com/stretchr/testify/require"
)

func TestStoreServiceGet(t *testing.T) {
	require := require.New(t)

	// Create temporary test directory
	tmpDir := t.TempDir()

	// Create test JSON file
	jsonContent := `{"name": "test", "value": 123}`
	jsonFile := filepath.Join(tmpDir, "test.json")
	err := os.WriteFile(jsonFile, []byte(jsonContent), 0644)
	require.NoError(err)

	// Create test config file
	configContent := `version: 0.1
namespaces:
  default:
    "test":
      driver: "local"
      args:
        root: "` + tmpDir + `"
      mount:
        file: ""
`
	configPath := filepath.Join(tmpDir, ".kvtool.yml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(err)

	// Test service
	ctx := context.Background()
	service := NewStoreService()

	result, err := service.Get(ctx, GetOptions{
		ConfigPath: configPath,
		Namespace:  "default",
		StorePath:  "test/test.json",
	})
	require.NoError(err)
	require.NotNil(result)

	// Verify result
	resultMap, ok := result.(map[string]any)
	require.True(ok, "result should be map[string]any")
	require.Equal("test", resultMap["name"])
	require.Equal("123", resultMap["value"].(json.Number).String())
}

func TestStoreServiceGetWithTransform(t *testing.T) {
	require := require.New(t)

	// Create temporary test directory
	tmpDir := t.TempDir()

	// Create test .env file
	envContent := `APP_NAME=myapp
DATABASE_URL=postgres://localhost/db
`
	envFile := filepath.Join(tmpDir, "test.env")
	err := os.WriteFile(envFile, []byte(envContent), 0644)
	require.NoError(err)

	// Create test config file with transform
	configContent := `version: 0.1
namespaces:
  default:
    "env":
      driver: "local"
      args:
        root: "` + tmpDir + `"
        transform:
          read: "dotenv"
      mount:
        file: ""
`
	configPath := filepath.Join(tmpDir, ".kvtool.yml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(err)

	// Test service
	ctx := context.Background()
	service := NewStoreService()

	result, err := service.Get(ctx, GetOptions{
		ConfigPath: configPath,
		Namespace:  "default",
		StorePath:  "env/test.env",
	})
	require.NoError(err)
	require.NotNil(result)

	// Verify result
	resultMap, ok := result.(map[string]any)
	require.True(ok, "result should be map[string]any")
	require.Equal("myapp", resultMap["APP_NAME"])
	require.Equal("postgres://localhost/db", resultMap["DATABASE_URL"])
}

func TestStoreServiceGetInvalidStorePath(t *testing.T) {
	require := require.New(t)

	ctx := context.Background()
	service := NewStoreService()

	_, err := service.Get(ctx, GetOptions{
		ConfigPath: "/nonexistent",
		Namespace:  "default",
		StorePath:  "", // Invalid: empty path
	})
	require.Error(err)
	require.Contains(err.Error(), "failed to parse store path")
}

func TestStoreServiceGetConfigNotFound(t *testing.T) {
	require := require.New(t)

	ctx := context.Background()
	service := NewStoreService()

	_, err := service.Get(ctx, GetOptions{
		ConfigPath: "/nonexistent/.kvtool.yml",
		Namespace:  "default",
		StorePath:  "test/file",
	})
	require.Error(err)
	require.Contains(err.Error(), "failed to load config")
}

func TestStoreServiceGetStoreNotFound(t *testing.T) {
	require := require.New(t)

	// Create temporary test directory
	tmpDir := t.TempDir()

	// Create test config file without the requested store
	configContent := `version: 0.1
namespaces:
  default:
    "existing":
      driver: "local"
      args:
        root: "."
`
	configPath := filepath.Join(tmpDir, ".kvtool.yml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(err)

	ctx := context.Background()
	service := NewStoreService()

	_, err = service.Get(ctx, GetOptions{
		ConfigPath: configPath,
		Namespace:  "default",
		StorePath:  "nonexistent/file.json",
	})
	require.Error(err)
	require.Contains(err.Error(), "failed to get store info")
}

// Mock implementations for testing

type mockConfigLoader struct {
	cfg *config.KvtoolConfig
	err error
}

func (m *mockConfigLoader) LoadConfig(path string) (*config.KvtoolConfig, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.cfg, nil
}

type mockFilesystemFactory struct {
	fs  filesystems.Filesystem
	err error
}

func (m *mockFilesystemFactory) Create(ctx context.Context, storeInfo *config.StoreInfo) (filesystems.Filesystem, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.fs, nil
}

func TestStoreServiceWithMocks(t *testing.T) {
	require := require.New(t)

	// Create mock filesystem
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "test.json")
	err := os.WriteFile(jsonFile, []byte(`{"mocked": true}`), 0644)
	require.NoError(err)

	mockFs, err := filesystems.GetLocalFs(context.Background(), &filesystems.LocalFsConfig{
		Root: tmpDir,
	})
	require.NoError(err)

	// Create mock config
	mockCfg := &config.KvtoolConfig{
		Version: 0.1,
		Namespaces: map[string]map[string]config.StoreInfo{
			"default": {
				"test": {
					Driver: "local",
					Args:   map[string]interface{}{"root": tmpDir},
				},
			},
		},
	}

	// Create service with mocks
	service := &storeService{
		configLoader: &mockConfigLoader{cfg: mockCfg},
		fsFactory:    &mockFilesystemFactory{fs: mockFs},
	}

	ctx := context.Background()
	result, err := service.Get(ctx, GetOptions{
		ConfigPath: "/any/path",
		Namespace:  "default",
		StorePath:  "test/test.json",
	})
	require.NoError(err)

	resultMap := result.(map[string]any)
	require.True(resultMap["mocked"].(bool))
}
