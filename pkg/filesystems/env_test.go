package filesystems

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFsEnvFilesystem(t *testing.T) {
	require := require.New(t)

	// Set test environment variables
	t.Setenv("TEST_VAR_1", "value1")
	t.Setenv("TEST_VAR_2", "value2")

	fs, err := NewEnvFs(context.Background(), &EnvFsConfig{})
	require.NoError(err)

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
	t.Setenv("TEST_VAR", "test_value")

	fs, err := NewEnvFs(context.Background(), &EnvFsConfig{})
	require.NoError(err)

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

func TestFsEnvFilesystem_Include(t *testing.T) {
	require := require.New(t)

	t.Setenv("APP_NAME", "myapp")
	t.Setenv("APP_PORT", "8080")
	t.Setenv("SECRET_KEY", "supersecret")

	fs, err := NewEnvFs(context.Background(), &EnvFsConfig{
		Include: []string{"APP_NAME", "APP_PORT"},
	})
	require.NoError(err)

	file, err := fs.GetFile("")
	require.NoError(err)

	data, err := file.LoadAsJson()
	require.NoError(err)

	dataMap := data.(map[string]any)
	require.Equal("myapp", dataMap["APP_NAME"])
	require.Equal("8080", dataMap["APP_PORT"])
	require.NotContains(dataMap, "SECRET_KEY", "include に含まれないキーは返らない")
}

func TestFsEnvFilesystem_Exclude(t *testing.T) {
	require := require.New(t)

	t.Setenv("APP_NAME", "myapp")
	t.Setenv("SECRET_KEY", "supersecret")
	t.Setenv("PASSWORD", "pass123")

	fs, err := NewEnvFs(context.Background(), &EnvFsConfig{
		Exclude: []string{"SECRET_KEY", "PASSWORD"},
	})
	require.NoError(err)

	file, err := fs.GetFile("")
	require.NoError(err)

	data, err := file.LoadAsJson()
	require.NoError(err)

	dataMap := data.(map[string]any)
	require.Equal("myapp", dataMap["APP_NAME"])
	require.NotContains(dataMap, "SECRET_KEY", "exclude されたキーは返らない")
	require.NotContains(dataMap, "PASSWORD", "exclude されたキーは返らない")
}

func TestFsEnvFilesystem_ExcludeOverridesInclude(t *testing.T) {
	require := require.New(t)

	t.Setenv("APP_NAME", "myapp")
	t.Setenv("APP_SECRET", "secret")
	t.Setenv("OTHER_VAR", "other")

	fs, err := NewEnvFs(context.Background(), &EnvFsConfig{
		Include: []string{"APP_NAME", "APP_SECRET"},
		Exclude: []string{"APP_SECRET"},
	})
	require.NoError(err)

	file, err := fs.GetFile("")
	require.NoError(err)

	data, err := file.LoadAsJson()
	require.NoError(err)

	dataMap := data.(map[string]any)
	require.Equal("myapp", dataMap["APP_NAME"])
	require.NotContains(dataMap, "APP_SECRET", "exclude が include より優先")
	require.NotContains(dataMap, "OTHER_VAR", "include に含まれないキーは返らない")
}
