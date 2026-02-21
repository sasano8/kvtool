package filesystems

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLocalFsGetFile(t *testing.T) {
	require := require.New(t)

	// Create temporary test directory
	tmpDir := t.TempDir()

	fs := &LocalFs{
		Root: tmpDir,
	}

	// GetFile should succeed with any path
	file, err := fs.GetFile("test.json")
	require.NoError(err)
	require.NotNil(file)
	require.IsType(&LocalFile{}, file)

	// Verify the file has correct properties
	localFile := file.(*LocalFile)
	require.Equal(fs, localFile.fs)
	require.Equal("test.json", localFile.path)
}

func TestLocalFsLoadAsJson(t *testing.T) {
	require := require.New(t)

	// Create temporary test directory
	tmpDir := t.TempDir()

	// Create test JSON file
	testData := map[string]any{
		"name":  "test",
		"value": 123,
		"nested": map[string]any{
			"key": "value",
		},
	}
	testFile := filepath.Join(tmpDir, "test.json")
	data, err := json.Marshal(testData)
	require.NoError(err)
	err = os.WriteFile(testFile, data, 0644)
	require.NoError(err)

	fs := &LocalFs{
		Root: tmpDir,
	}

	// Test LoadAsJson through File interface
	file, err := fs.GetFile("test.json")
	require.NoError(err)

	result, err := file.LoadAsJson()
	require.NoError(err)
	require.NotNil(result)

	// Verify data
	resultMap, ok := result.(map[string]any)
	require.True(ok, "result should be a map")
	require.Equal("test", resultMap["name"])
	require.Equal(json.Number("123"), resultMap["value"]) // UseNumber() converts to json.Number
}

func TestLocalFsOpenReader(t *testing.T) {
	require := require.New(t)

	// Create temporary test directory
	tmpDir := t.TempDir()

	// Create test file
	testContent := `{"key": "value", "number": 42}`
	testFile := filepath.Join(tmpDir, "data.json")
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	require.NoError(err)

	fs := &LocalFs{
		Root: tmpDir,
	}

	// Test OpenReader through File interface
	file, err := fs.GetFile("data.json")
	require.NoError(err)

	reader, err := file.OpenReader()
	require.NoError(err)
	require.NotNil(reader)
	defer reader.Close()

	// Read content
	content, err := io.ReadAll(reader)
	require.NoError(err)
	require.Equal(testContent, string(content))
}

func TestLocalFsFileNotFound(t *testing.T) {
	require := require.New(t)

	tmpDir := t.TempDir()

	fs := &LocalFs{
		Root: tmpDir,
	}

	// GetFile succeeds (lazy evaluation)
	file, err := fs.GetFile("nonexistent.json")
	require.NoError(err)

	// But LoadAsJson fails
	_, err = file.LoadAsJson()
	require.Error(err)
	require.Contains(err.Error(), "no such file or directory")

	// OpenReader also fails
	_, err = file.OpenReader()
	require.Error(err)
	require.Contains(err.Error(), "no such file or directory")
}

func TestLocalFsEmptyPath(t *testing.T) {
	require := require.New(t)

	tmpDir := t.TempDir()

	fs := &LocalFs{
		Root: tmpDir,
	}

	file, err := fs.GetFile("")
	require.NoError(err)

	// Empty path should fail
	_, err = file.OpenReader()
	require.Error(err)
	require.Contains(err.Error(), "path is empty")
}

func TestLocalFsAbsolutePath(t *testing.T) {
	require := require.New(t)

	tmpDir := t.TempDir()

	fs := &LocalFs{
		Root: tmpDir,
	}

	// Absolute paths should be rejected
	file, err := fs.GetFile("/etc/passwd")
	require.NoError(err)

	_, err = file.OpenReader()
	require.Error(err)
	require.Contains(err.Error(), "path must be relative")
}

func TestLocalFsPathWithTilde(t *testing.T) {
	require := require.New(t)

	tmpDir := t.TempDir()

	fs := &LocalFs{
		Root: tmpDir,
	}

	// Paths with ~ should be rejected
	file, err := fs.GetFile("~/test.json")
	require.NoError(err)

	_, err = file.OpenReader()
	require.Error(err)
	require.Contains(err.Error(), "path must be relative (no ~ expansion)")
}

func TestLocalFsPathTraversal(t *testing.T) {
	// Create temporary test directory
	tmpDir := t.TempDir()

	// Create a file outside tmpDir to attempt accessing
	parentDir := filepath.Dir(tmpDir)
	outsideFile := filepath.Join(parentDir, "outside.json")
	err := os.WriteFile(outsideFile, []byte(`{"outside": true}`), 0644)
	require.NoError(t, err)
	defer os.Remove(outsideFile)

	fs := &LocalFs{
		Root: tmpDir,
	}

	// Attempt path traversal attacks
	testCases := []string{
		"../outside.json",
		"../../outside.json",
		"../../../etc/passwd",
		"subdir/../../outside.json",
		"./../../outside.json",
	}

	for _, testPath := range testCases {
		t.Run(testPath, func(t *testing.T) {
			req := require.New(t)

			file, err := fs.GetFile(testPath)
			req.NoError(err)

			// Should fail with path escapes root error
			_, err = file.OpenReader()
			req.Error(err, "path traversal should be blocked: %s", testPath)
			req.Contains(err.Error(), "path escapes root", "path: %s", testPath)
		})
	}
}

func TestLocalFsRootAsPath(t *testing.T) {
	tmpDir := t.TempDir()

	fs := &LocalFs{
		Root: tmpDir,
	}

	// Attempting to access root directory itself
	testCases := []string{
		".",
		"./",
	}

	for _, testPath := range testCases {
		t.Run(testPath, func(t *testing.T) {
			req := require.New(t)

			file, err := fs.GetFile(testPath)
			req.NoError(err)

			_, err = file.OpenReader()
			req.Error(err)
			req.Contains(err.Error(), "path points to root directory")
		})
	}
}

func TestLocalFsSubdirectory(t *testing.T) {
	require := require.New(t)

	// Create temporary test directory with subdirectory
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	err := os.MkdirAll(subDir, 0755)
	require.NoError(err)

	// Create test file in subdirectory
	testData := map[string]any{"subdir": true}
	testFile := filepath.Join(subDir, "test.json")
	data, err := json.Marshal(testData)
	require.NoError(err)
	err = os.WriteFile(testFile, data, 0644)
	require.NoError(err)

	fs := &LocalFs{
		Root: tmpDir,
	}

	// Should successfully read from subdirectory
	file, err := fs.GetFile("subdir/test.json")
	require.NoError(err)

	result, err := file.LoadAsJson()
	require.NoError(err)

	resultMap := result.(map[string]any)
	require.Equal(true, resultMap["subdir"])
}

func TestLocalFsDefaultRoot(t *testing.T) {
	require := require.New(t)

	// Create test file in current working directory
	cwd, err := os.Getwd()
	require.NoError(err)

	testFile := filepath.Join(cwd, "test_default_root.json")
	testData := map[string]any{"test": "default_root"}
	data, err := json.Marshal(testData)
	require.NoError(err)
	err = os.WriteFile(testFile, data, 0644)
	require.NoError(err)
	defer os.Remove(testFile)

	// Empty root should default to current working directory
	fs := &LocalFs{
		Root: "",
	}

	file, err := fs.GetFile("test_default_root.json")
	require.NoError(err)

	result, err := file.LoadAsJson()
	require.NoError(err)

	resultMap := result.(map[string]any)
	require.Equal("default_root", resultMap["test"])
}
