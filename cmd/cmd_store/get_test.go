package cmd_store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStoreGetCommand(t *testing.T) {
	require := require.New(t)

	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Create test .env file
	envContent := `APP_NAME=kvtool
APP_ENV=test
DATABASE_URL=postgres://localhost:5432/testdb
`
	envPath := filepath.Join(tmpDir, "test.env")
	err := os.WriteFile(envPath, []byte(envContent), 0644)
	require.NoError(err)

	// Create test config file
	configContent := `version: 0.1
namespaces:
  default:
    "test":
      driver: "local"
      args:
        transform:
          read: dotenv
          write: dotenv
        root: "` + tmpDir + `"
      mount:
        file: ""
`
	configPath := filepath.Join(tmpDir, ".kvtool.yml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(err)

	// Test the command
	cmd := CmdGet
	cmd.SetArgs([]string{"test/test.env", "-c", configPath})

	err = cmd.Execute()
	require.NoError(err)
}
