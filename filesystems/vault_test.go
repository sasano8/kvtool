package filesystems

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestVaultFs(t *testing.T) {
	// Skip if VAULT_ADDR is not set (Vault not running)
	vaultAddr := os.Getenv("VAULT_ADDR")
	if vaultAddr == "" {
		t.Skip("VAULT_ADDR not set, skipping Vault tests")
	}

	vaultToken := os.Getenv("VAULT_TOKEN")
	if vaultToken == "" {
		vaultToken = "root" // default for dev mode
	}

	require := require.New(t)

	config := &VaultConfig{
		Addr:      vaultAddr,
		Token:     vaultToken,
		Namespace: "admin",
		Mount:     "secret",
		KvVer:     2,
		Version:   0,
		Timeout:   10 * time.Second,
	}

	ctx := context.Background()
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
	// Skip if VAULT_ADDR is not set
	vaultAddr := os.Getenv("VAULT_ADDR")
	if vaultAddr == "" {
		t.Skip("VAULT_ADDR not set, skipping Vault tests")
	}

	vaultToken := os.Getenv("VAULT_TOKEN")
	if vaultToken == "" {
		vaultToken = "root"
	}

	require := require.New(t)

	config := &VaultConfig{
		Addr:      vaultAddr,
		Token:     vaultToken,
		Namespace: "admin",
		Mount:     "secret",
		KvVer:     2,
		Version:   0,
		Timeout:   10 * time.Second,
	}

	ctx := context.Background()
	vaultFs, err := GetVaultFs(ctx, config)
	require.NoError(err)

	// Test getting a non-existent file
	file, err := vaultFs.GetFile("nonexistent/path")
	require.NoError(err)

	_, err = file.LoadAsJson()
	require.Error(err, "should error when loading non-existent secret")
}
