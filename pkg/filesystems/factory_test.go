package filesystems

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilesystemFactory(t *testing.T) {
	ctx := context.Background()
	factory := NewFilesystemFactory(ctx)
	require.NotNil(t, factory)
}

func TestFactoryCreateLocalFs(t *testing.T) {
	require := require.New(t)

	ctx := context.Background()
	factory := NewFilesystemFactory(ctx)

	storeInfo := &StoreInfo{
		Driver: "local",
		Args: map[string]interface{}{
			"root": "/tmp",
		},
	}

	fs, err := factory.Create(storeInfo)
	require.NoError(err)
	require.NotNil(fs)
}

func TestFactoryCreateEnvFs(t *testing.T) {
	require := require.New(t)

	ctx := context.Background()
	factory := NewFilesystemFactory(ctx)

	storeInfo := &StoreInfo{
		Driver: "env",
		Args:   map[string]interface{}{},
	}

	fs, err := factory.Create(storeInfo)
	require.NoError(err)
	require.NotNil(fs)

	// Verify it's an env filesystem
	_, ok := fs.(*FsEnvFilesystem)
	require.True(ok, "should be FsEnvFilesystem")
}

func TestFactoryCreateUnsupportedDriver(t *testing.T) {
	require := require.New(t)

	ctx := context.Background()
	factory := NewFilesystemFactory(ctx)

	storeInfo := &StoreInfo{
		Driver: "unsupported",
		Args:   map[string]interface{}{},
	}

	fs, err := factory.Create(storeInfo)
	require.Error(err)
	require.Nil(fs)
	require.Contains(err.Error(), "unsupported driver")
}

func TestFactoryCreateLocalFsWithDefaults(t *testing.T) {
	require := require.New(t)

	ctx := context.Background()
	factory := NewFilesystemFactory(ctx)

	// No root specified - should default to "."
	storeInfo := &StoreInfo{
		Driver: "local",
		Args:   map[string]interface{}{},
	}

	fs, err := factory.Create(storeInfo)
	require.NoError(err)
	require.NotNil(fs)

	localFs, ok := fs.(*LocalFs)
	require.True(ok, "should be LocalFs")
	require.Equal(".", localFs.Root)
}

func TestFactoryCreateToolFs(t *testing.T) {
	require := require.New(t)

	ctx := context.Background()
	factory := NewFilesystemFactory(ctx)

	storeInfo := &StoreInfo{
		Driver: "tool",
		Args:   map[string]interface{}{},
	}

	fs, err := factory.Create(storeInfo)
	require.NoError(err)
	require.NotNil(fs)

	// uuid7 を取得
	file, err := fs.GetFile("uuid7")
	require.NoError(err)

	data, err := file.LoadAsJson()
	require.NoError(err)

	// 値がそのまま返される
	value, ok := data.(string)
	require.True(ok)
	require.Len(value, 36) // UUID format
}

func TestFactoryCreateEnvFsWithIncludeExclude(t *testing.T) {
	require := require.New(t)

	t.Setenv("ALLOWED_VAR", "allowed")
	t.Setenv("BLOCKED_VAR", "blocked")
	t.Setenv("OTHER_VAR", "other")

	ctx := context.Background()
	factory := NewFilesystemFactory(ctx)

	storeInfo := &StoreInfo{
		Driver: "env",
		Args: map[string]interface{}{
			"include": []interface{}{"ALLOWED_VAR", "BLOCKED_VAR"},
			"exclude": []interface{}{"BLOCKED_VAR"},
		},
	}

	fs, err := factory.Create(storeInfo)
	require.NoError(err)
	require.NotNil(fs)

	file, err := fs.GetFile("")
	require.NoError(err)

	data, err := file.LoadAsJson()
	require.NoError(err)

	dataMap := data.(map[string]any)
	require.Equal("allowed", dataMap["ALLOWED_VAR"])
	require.NotContains(dataMap, "BLOCKED_VAR", "exclude が優先")
	require.NotContains(dataMap, "OTHER_VAR", "include に含まれない")
}

func TestGetFilesystemConvenienceFunction(t *testing.T) {
	require := require.New(t)

	ctx := context.Background()
	storeInfo := &StoreInfo{
		Driver: "env",
		Args:   map[string]interface{}{},
	}

	fs, err := GetFilesystem(ctx, storeInfo)
	require.NoError(err)
	require.NotNil(fs)

	_, ok := fs.(*FsEnvFilesystem)
	require.True(ok, "should be FsEnvFilesystem")
}
