package filesystems

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFsEnvFilesystem(t *testing.T) {
	require := require.New(t)

	// Set test environment variables
	os.Setenv("TEST_VAR_1", "value1")
	os.Setenv("TEST_VAR_2", "value2")
	defer os.Unsetenv("TEST_VAR_1")
	defer os.Unsetenv("TEST_VAR_2")

	fs := &FsEnvFilesystem{}

	// GetFile should work with any path (path is ignored)
	file1, err := fs.GetFile("")
	require.NoError(err)
	require.NotNil(file1)

	file2, err := fs.GetFile("ignored/path")
	require.NoError(err)
	require.NotNil(file2)

	// LoadAsJson should return environment variables
	data, err := file1.LoadAsJson()
	require.NoError(err)
	require.NotNil(data)

	dataMap, ok := data.(map[string]any)
	require.True(ok, "data should be a map")
	require.Contains(dataMap, "TEST_VAR_1")
	require.Equal("value1", dataMap["TEST_VAR_1"])
	require.Contains(dataMap, "TEST_VAR_2")
	require.Equal("value2", dataMap["TEST_VAR_2"])
}

func TestFsEnvFileOpenReader(t *testing.T) {
	require := require.New(t)

	// Set test environment variable
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	fs := &FsEnvFilesystem{}
	file, err := fs.GetFile("")
	require.NoError(err)

	// OpenReader should return JSON-encoded stream
	reader, err := file.OpenReader()
	require.NoError(err)
	require.NotNil(reader)
	defer reader.Close()

	// Read the stream
	content, err := io.ReadAll(reader)
	require.NoError(err)
	require.NotEmpty(content)

	// Verify it's valid JSON containing our test var
	require.Contains(string(content), "TEST_VAR")
	require.Contains(string(content), "test_value")
}