package filesystems

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestVaultFs(t *testing.T) {
	// Vault が起動していない場合はスキップ
	if testing.Short() || os.Getenv("SKIP_VAULT_TESTS") == "true" {
		t.Skip("Vault tests skipped (-short flag or SKIP_VAULT_TESTS=true)")
	}

	require := require.New(t)

	config := &VaultConfig{
		Addr:      getTestVaultAddr(),
		Token:     getTestVaultToken(),
		Namespace: "admin",
		Mount:     "secret",
		KvVer:     2,
		Version:   0,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	vaultFs, err := GetVaultFs(ctx, config)
	require.NoError(err)
	require.NotNil(vaultFs)

	// Test getting a file that should exist (created by vault-init in docker-compose)
	file, err := vaultFs.GetFile("app/prod")
	require.NoError(err)

	data, err := file.LoadAsJson()
	require.NoError(err)
	require.NotNil(data)

	// Verify the test data
	dataMap, ok := data.(map[string]interface{})
	require.True(ok, "data should be a map")
	require.Equal("alice", dataMap["username"])
	require.Equal("pa55w0rd", dataMap["password"])
}

func TestVaultFsNonExistentPath(t *testing.T) {
	// Vault が起動していない場合はスキップ
	if testing.Short() || os.Getenv("SKIP_VAULT_TESTS") == "true" {
		t.Skip("Vault tests skipped (-short flag or SKIP_VAULT_TESTS=true)")
	}

	require := require.New(t)

	config := &VaultConfig{
		Addr:      getTestVaultAddr(),
		Token:     getTestVaultToken(),
		Namespace: "admin",
		Mount:     "secret",
		KvVer:     2,
		Version:   0,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	vaultFs, err := GetVaultFs(ctx, config)
	require.NoError(err)

	// Test getting a non-existent file
	file, err := vaultFs.GetFile("nonexistent/path")
	require.NoError(err)

	_, err = file.LoadAsJson()
	require.Error(err, "should error when loading non-existent secret")
}
