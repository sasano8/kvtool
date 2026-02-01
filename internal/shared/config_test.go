package shared

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	require := require.New(t)

	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".kvtool.yml")

	configContent := `version: 0.1
namespaces:
  default:
    ".env":
      driver: "local"
      args:
        ext: ".env"
        transform:
          read: dotenv
          write: dotenv
        root: "."
      mount:
        file: ""
    "vault":
      driver: vault
      args:
        conn:
          addr: "http://localhost:8200"
          token: root
          mount: secret
        transform:
          read: null
          write: null
        root: app/prod
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(err)

	// Load config
	config, err := LoadConfig(configPath)
	require.NoError(err)
	require.NotNil(config)
	require.Equal(0.1, config.Version)

	// Test GetStore
	store, err := config.GetStore("default", ".env")
	require.NoError(err)
	require.NotNil(store)
	require.Equal("local", store.Driver)

	vaultStore, err := config.GetStore("default", "vault")
	require.NoError(err)
	require.NotNil(vaultStore)
	require.Equal("vault", vaultStore.Driver)

	// Test non-existent namespace
	_, err = config.GetStore("nonexistent", ".env")
	require.Error(err)

	// Test non-existent store
	_, err = config.GetStore("default", "nonexistent")
	require.Error(err)
}

func TestParseStorePath(t *testing.T) {
	require := require.New(t)

	// Test valid paths
	tests := []struct {
		input     string
		storeName string
		filePath  string
		wantErr   bool
	}{
		{
			input:     ".env/APP_NAME",
			storeName: ".env",
			filePath:  "APP_NAME",
			wantErr:   false,
		},
		{
			input:     "vault/app/prod/db_password",
			storeName: "vault",
			filePath:  "app/prod/db_password",
			wantErr:   false,
		},
		{
			input:     ".env",
			storeName: ".env",
			filePath:  "",
			wantErr:   false,
		},
		{
			input:     "vault/config",
			storeName: "vault",
			filePath:  "config",
			wantErr:   false,
		},
		{
			input:   "",
			wantErr: true,
		},
		{
			input:     ".env/",
			storeName: ".env",
			filePath:  "",
			wantErr:   false,
		},
		{
			input:     "/.env/APP_NAME",
			storeName: ".env",
			filePath:  "APP_NAME",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseStorePath(tt.input)
			if tt.wantErr {
				require.Error(err)
				return
			}

			require.NoError(err)
			require.NotNil(result)
			require.Equal(tt.storeName, result.StoreName)
			require.Equal(tt.filePath, result.FilePath)
		})
	}
}
