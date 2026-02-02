package sources

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSourceEnvImplementsInterface(t *testing.T) {
	// Source インターフェースを実装していることを確認
	var _ Source = (*SourceEnv)(nil)
}

func TestSourceEnvLoad(t *testing.T) {
	require := require.New(t)

	// テスト用の環境変数を設定
	os.Setenv("TEST_SOURCE_KEY", "test_value")
	defer os.Unsetenv("TEST_SOURCE_KEY")

	// Source インターフェースとして使用
	var source Source = &SourceEnv{}
	reader, err := source.Load()
	require.NoError(err)
	require.NotNil(reader)

	// データを読み取る
	data, err := io.ReadAll(reader)
	require.NoError(err)

	// TEST_SOURCE_KEY が含まれていることを確認
	content := string(data)
	require.Contains(content, "TEST_SOURCE_KEY=test_value")
}

func TestSourceEnvFormat(t *testing.T) {
	require := require.New(t)

	os.Setenv("KEY1", "value1")
	os.Setenv("KEY2", "value2")
	defer os.Unsetenv("KEY1")
	defer os.Unsetenv("KEY2")

	source := &SourceEnv{}
	reader, err := source.Load()
	require.NoError(err)

	data, err := io.ReadAll(reader)
	require.NoError(err)

	// フォーマットが KEY=VALUE\n であることを確認
	content := string(data)
	lines := strings.Split(content, "\n")

	foundKey1 := false
	foundKey2 := false
	for _, line := range lines {
		if line == "KEY1=value1" {
			foundKey1 = true
		}
		if line == "KEY2=value2" {
			foundKey2 = true
		}
	}

	require.True(foundKey1, "KEY1=value1 should be in output")
	require.True(foundKey2, "KEY2=value2 should be in output")
}

func TestSourceEnvEscapeNewline(t *testing.T) {
	require := require.New(t)

	// 改行を含む環境変数を設定
	os.Setenv("MULTILINE_KEY", "line1\nline2\nline3")
	defer os.Unsetenv("MULTILINE_KEY")

	source := &SourceEnv{}
	reader, err := source.Load()
	require.NoError(err)

	data, err := io.ReadAll(reader)
	require.NoError(err)

	content := string(data)
	lines := strings.Split(content, "\n")

	// エスケープされた形式で出力されることを確認
	found := false
	for _, line := range lines {
		if line == `MULTILINE_KEY="line1\nline2\nline3"` {
			found = true
			break
		}
	}

	require.True(found, "MULTILINE_KEY should be escaped with quotes")
}

func TestSourceEnvEscapeTab(t *testing.T) {
	require := require.New(t)

	os.Setenv("TAB_KEY", "value\twith\ttabs")
	defer os.Unsetenv("TAB_KEY")

	source := &SourceEnv{}
	reader, err := source.Load()
	require.NoError(err)

	data, err := io.ReadAll(reader)
	require.NoError(err)

	content := string(data)
	lines := strings.Split(content, "\n")

	found := false
	for _, line := range lines {
		if line == `TAB_KEY="value\twith\ttabs"` {
			found = true
			break
		}
	}

	require.True(found, "TAB_KEY should be escaped with quotes")
}

func TestSourceEnvEscapeBackslash(t *testing.T) {
	require := require.New(t)

	os.Setenv("BACKSLASH_KEY", `path\to\file`)
	defer os.Unsetenv("BACKSLASH_KEY")

	source := &SourceEnv{}
	reader, err := source.Load()
	require.NoError(err)

	data, err := io.ReadAll(reader)
	require.NoError(err)

	content := string(data)
	lines := strings.Split(content, "\n")

	found := false
	for _, line := range lines {
		if line == `BACKSLASH_KEY="path\\to\\file"` {
			found = true
			break
		}
	}

	require.True(found, "BACKSLASH_KEY should be escaped with quotes")
}

func TestSourceEnvEscapeDoubleQuote(t *testing.T) {
	require := require.New(t)

	os.Setenv("QUOTE_KEY", `value with "quotes"`)
	defer os.Unsetenv("QUOTE_KEY")

	source := &SourceEnv{}
	reader, err := source.Load()
	require.NoError(err)

	data, err := io.ReadAll(reader)
	require.NoError(err)

	content := string(data)
	lines := strings.Split(content, "\n")

	found := false
	for _, line := range lines {
		if line == `QUOTE_KEY="value with \"quotes\""` {
			found = true
			break
		}
	}

	require.True(found, "QUOTE_KEY should be escaped with quotes")
}

func TestSourceEnvEscapeMultipleSpecialChars(t *testing.T) {
	require := require.New(t)

	// 複数の特殊文字を含む環境変数
	os.Setenv("COMPLEX_KEY", "line1\nline2\t\"quoted\"\r\\backslash")
	defer os.Unsetenv("COMPLEX_KEY")

	source := &SourceEnv{}
	reader, err := source.Load()
	require.NoError(err)

	data, err := io.ReadAll(reader)
	require.NoError(err)

	content := string(data)
	lines := strings.Split(content, "\n")

	found := false
	for _, line := range lines {
		if line == `COMPLEX_KEY="line1\nline2\t\"quoted\"\r\\backslash"` {
			found = true
			break
		}
	}

	require.True(found, "COMPLEX_KEY should be escaped with all special characters")
}
