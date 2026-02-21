package filesystems

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGitFs_ConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *GitFsConfig
		expectError bool
		errorMsg    string
	}{
		{
			name:        "URL が空の場合エラー",
			config:      &GitFsConfig{},
			expectError: true,
			errorMsg:    "url is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			_, err := NewGitFs(ctx, tt.config)
			if tt.expectError {
				require.Error(err)
				require.Contains(err.Error(), tt.errorMsg)
			} else {
				require.NoError(err)
			}
		})
	}
}

func TestGitFs_CacheDir(t *testing.T) {
	tests := []struct {
		name string
		url  string
		ref  string
	}{
		{
			name: "同じ URL+ref は同じキャッシュディレクトリ",
			url:  "https://github.com/example/repo.git",
			ref:  "main",
		},
		{
			name: "異なる ref は異なるキャッシュディレクトリ",
			url:  "https://github.com/example/repo.git",
			ref:  "develop",
		},
	}

	require := require.New(t)

	// 同じ URL+ref は一貫性がある
	dir1 := GitCacheDir(tests[0].url, tests[0].ref)
	dir2 := GitCacheDir(tests[0].url, tests[0].ref)
	require.Equal(dir1, dir2, "同じ URL+ref は同じキャッシュディレクトリを返すべき")

	// 異なる ref は異なるディレクトリ
	dir3 := GitCacheDir(tests[1].url, tests[1].ref)
	require.NotEqual(dir1, dir3, "異なる ref は異なるキャッシュディレクトリを返すべき")

	// ディレクトリパスに期待される構造が含まれる
	require.Contains(dir1, ".cache/kvtool/git/")
}

func TestGitFs_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	if getTestSkipGit() {
		t.Skip("skipping git tests (SKIP_GIT_TESTS=true)")
	}

	t.Run("ローカル bare リポジトリからの clone と GetFile", func(t *testing.T) {
		require := require.New(t)

		bareRepoPath, cleanup := setupTestGitRepo(t)
		defer cleanup()

		config := &GitFsConfig{
			URL: bareRepoPath,
			Ref: "main",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		fs, err := NewGitFs(ctx, config)
		require.NoError(err)

		// Sync
		err = fs.Sync()
		require.NoError(err)

		// GetFile
		file, err := fs.GetFile("config.json")
		require.NoError(err)

		result, err := file.LoadAsJson()
		require.NoError(err)
		require.NotNil(result)

		// JSON の中身を確認
		m, ok := result.(map[string]interface{})
		require.True(ok)
		require.Equal("test", m["app"])
	})

	t.Run("自動 Sync（inner が nil の場合）", func(t *testing.T) {
		require := require.New(t)

		bareRepoPath, cleanup := setupTestGitRepo(t)
		defer cleanup()

		config := &GitFsConfig{
			URL: bareRepoPath,
			Ref: "main",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		fs, err := NewGitFs(ctx, config)
		require.NoError(err)

		// Sync を呼ばずに GetFile → 自動 Sync される
		file, err := fs.GetFile("config.json")
		require.NoError(err)

		result, err := file.LoadAsJson()
		require.NoError(err)
		require.NotNil(result)
	})
}
