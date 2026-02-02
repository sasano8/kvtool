package filesystems

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLocalFsTransformDotenv(t *testing.T) {
	require := require.New(t)

	// Create temporary test directory
	tmpDir := t.TempDir()

	// Create a .env file
	envContent := `APP_NAME=myapp
DATABASE_URL=postgres://localhost/db
DEBUG=true
`
	envFile := filepath.Join(tmpDir, "test.env")
	err := os.WriteFile(envFile, []byte(envContent), 0644)
	require.NoError(err)

	// Create LocalFs with dotenv transform
	fs, err := GetLocalFs(context.Background(), &LocalFsConfig{
		Root:      tmpDir,
		Transform: "dotenv",
	})
	require.NoError(err)

	// Get file and load as JSON
	file, err := fs.GetFile("test.env")
	require.NoError(err)

	result, err := file.LoadAsJson()
	require.NoError(err)

	// Verify the result is a map
	resultMap, ok := result.(map[string]any)
	require.True(ok, "result should be map[string]any")

	// Verify the parsed values
	require.Equal("myapp", resultMap["APP_NAME"])
	require.Equal("postgres://localhost/db", resultMap["DATABASE_URL"])
	require.Equal("true", resultMap["DEBUG"])
}

func TestLocalFsTransformJson(t *testing.T) {
	require := require.New(t)

	// Create temporary test directory
	tmpDir := t.TempDir()

	// Create a JSON file
	jsonContent := `{"name": "test", "value": 123}`
	jsonFile := filepath.Join(tmpDir, "test.json")
	err := os.WriteFile(jsonFile, []byte(jsonContent), 0644)
	require.NoError(err)

	// Create LocalFs with json transform (explicit)
	fs, err := GetLocalFs(context.Background(), &LocalFsConfig{
		Root:      tmpDir,
		Transform: "json",
	})
	require.NoError(err)

	// Get file and load as JSON
	file, err := fs.GetFile("test.json")
	require.NoError(err)

	result, err := file.LoadAsJson()
	require.NoError(err)

	// Verify the result
	resultMap := result.(map[string]any)
	require.Equal("test", resultMap["name"])
	require.Equal("123", resultMap["value"].(json.Number).String())
}

func TestLocalFsTransformDefault(t *testing.T) {
	require := require.New(t)

	// Create temporary test directory
	tmpDir := t.TempDir()

	// Create a JSON file
	jsonContent := `{"name": "default", "value": 456}`
	jsonFile := filepath.Join(tmpDir, "test.json")
	err := os.WriteFile(jsonFile, []byte(jsonContent), 0644)
	require.NoError(err)

	// Create LocalFs without transform (should default to JSON)
	fs, err := GetLocalFs(context.Background(), &LocalFsConfig{
		Root:      tmpDir,
		Transform: "", // Empty string should default to JSON
	})
	require.NoError(err)

	// Get file and load as JSON
	file, err := fs.GetFile("test.json")
	require.NoError(err)

	result, err := file.LoadAsJson()
	require.NoError(err)

	// Verify the result
	resultMap := result.(map[string]any)
	require.Equal("default", resultMap["name"])
	require.Equal("456", resultMap["value"].(json.Number).String())
}

func TestLocalFsTransformUnsupported(t *testing.T) {
	require := require.New(t)

	// Create temporary test directory
	tmpDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(err)

	// Create LocalFs with unsupported transform
	fs, err := GetLocalFs(context.Background(), &LocalFsConfig{
		Root:      tmpDir,
		Transform: "unsupported",
	})
	require.NoError(err)

	// Get file
	file, err := fs.GetFile("test.txt")
	require.NoError(err)

	// LoadAsJson should return error for unsupported transform
	_, err = file.LoadAsJson()
	require.Error(err)
	require.Contains(err.Error(), "unsupported transform format")
}

func TestFactoryWithTransform(t *testing.T) {
	require := require.New(t)

	// Create temporary test directory
	tmpDir := t.TempDir()

	// Create a .env file
	envContent := `KEY1=value1
KEY2=value2
`
	envFile := filepath.Join(tmpDir, "test.env")
	err := os.WriteFile(envFile, []byte(envContent), 0644)
	require.NoError(err)

	// Create filesystem using factory with transform config
	ctx := context.Background()
	factory := NewFilesystemFactory(ctx)

	storeInfo := &StoreInfo{
		Driver: "local",
		Args: map[string]interface{}{
			"root": tmpDir,
			"transform": map[string]interface{}{
				"read": "dotenv",
			},
		},
	}

	fs, err := factory.Create(storeInfo)
	require.NoError(err)

	// Verify filesystem was created correctly
	localFs, ok := fs.(*LocalFs)
	require.True(ok, "should be LocalFs")
	require.Equal("dotenv", localFs.Transform)

	// Test loading file with transform
	file, err := fs.GetFile("test.env")
	require.NoError(err)

	result, err := file.LoadAsJson()
	require.NoError(err)

	resultMap := result.(map[string]any)
	require.Equal("value1", resultMap["KEY1"])
	require.Equal("value2", resultMap["KEY2"])
}
