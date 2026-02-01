package cmd_store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestStoreInitCommand(t *testing.T) {
	require := require.New(t)

	// Create a temporary directory
	tmpDir := t.TempDir()

	// Test local init
	t.Run("local init", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, ".kvtool.yml")

		cmd := CmdInit
		cmd.SetArgs([]string{configPath})

		err := cmd.Execute()
		require.NoError(err)

		// Verify file was created
		_, err = os.Stat(configPath)
		require.NoError(err)

		// Verify content is valid YAML
		data, err := os.ReadFile(configPath)
		require.NoError(err)

		var config map[string]interface{}
		err = yaml.Unmarshal(data, &config)
		require.NoError(err)
		require.NotNil(config["namespaces"])
	})

	// Test overwrite protection
	t.Run("overwrite protection", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "existing.yml")

		// Create existing file
		err := os.WriteFile(configPath, []byte("test"), 0644)
		require.NoError(err)

		cmd := CmdInit
		cmd.SetArgs([]string{configPath})

		err = cmd.Execute()
		require.Error(err)
		require.Contains(err.Error(), "already exists")
	})

	// Test force overwrite
	t.Run("force overwrite", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "force.yml")

		// Create existing file
		err := os.WriteFile(configPath, []byte("test"), 0644)
		require.NoError(err)

		cmd := CmdInit
		cmd.SetArgs([]string{configPath, "--force"})

		err = cmd.Execute()
		require.NoError(err)

		// Verify content was overwritten
		data, err := os.ReadFile(configPath)
		require.NoError(err)
		require.NotEqual("test", string(data))
	})
}
