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

	// Node.js dotenv 互換: タブはエスケープしない（そのまま出力）
	os.Setenv("TAB_KEY", "value\twith\ttabs")
	defer os.Unsetenv("TAB_KEY")

	source := &SourceEnv{}
	reader, err := source.Load()
	require.NoError(err)

	data, err := io.ReadAll(reader)
	require.NoError(err)

	content := string(data)
	// タブはエスケープされずそのまま出力される
	require.Contains(content, "TAB_KEY=value\twith\ttabs", "TAB_KEY should not be escaped")
}

func TestSourceEnvEscapeBackslash(t *testing.T) {
	require := require.New(t)

	// Node.js dotenv 互換: バックスラッシュはエスケープしない（そのまま出力）
	os.Setenv("BACKSLASH_KEY", `path\to\file`)
	defer os.Unsetenv("BACKSLASH_KEY")

	source := &SourceEnv{}
	reader, err := source.Load()
	require.NoError(err)

	data, err := io.ReadAll(reader)
	require.NoError(err)

	content := string(data)
	// バックスラッシュはエスケープされずそのまま出力される
	require.Contains(content, `BACKSLASH_KEY=path\to\file`, "BACKSLASH_KEY should not be escaped")
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
	// \n, \r, " はエスケープされる
	// \t, \ はエスケープされない（Node.js dotenv 互換）
	os.Setenv("COMPLEX_KEY", "line1\nline2\t\"quoted\"\r\\backslash")
	defer os.Unsetenv("COMPLEX_KEY")

	source := &SourceEnv{}
	reader, err := source.Load()
	require.NoError(err)

	data, err := io.ReadAll(reader)
	require.NoError(err)

	content := string(data)
	// \n, \r, " はエスケープ、\t と \ はそのまま
	// 期待値: COMPLEX_KEY="line1\nline2<TAB>\"quoted\"\r\backslash"
	expectedLine := "COMPLEX_KEY=\"line1\\nline2\t\\\"quoted\\\"\\r\\backslash\""
	require.Contains(content, expectedLine, "COMPLEX_KEY should escape \\n, \\r, \\\" but not \\t, \\\\")
}

func TestSourceEnvEmptyValue(t *testing.T) {
	require := require.New(t)

	// 空の値を持つ環境変数
	os.Setenv("EMPTY_VALUE_KEY", "")
	defer os.Unsetenv("EMPTY_VALUE_KEY")

	source := &SourceEnv{}
	reader, err := source.Load()
	require.NoError(err)

	data, err := io.ReadAll(reader)
	require.NoError(err)

	content := string(data)
	lines := strings.Split(content, "\n")

	found := false
	for _, line := range lines {
		if line == "EMPTY_VALUE_KEY=" {
			found = true
			break
		}
	}

	require.True(found, "EMPTY_VALUE_KEY= should be in output")
}

func TestSourceEnvValueWithEquals(t *testing.T) {
	require := require.New(t)

	// 値に = を含む環境変数
	os.Setenv("EQUALS_KEY", "VALUE=VALUE2=VALUE3")
	defer os.Unsetenv("EQUALS_KEY")

	source := &SourceEnv{}
	reader, err := source.Load()
	require.NoError(err)

	data, err := io.ReadAll(reader)
	require.NoError(err)

	content := string(data)
	lines := strings.Split(content, "\n")

	found := false
	for _, line := range lines {
		if line == "EQUALS_KEY=VALUE=VALUE2=VALUE3" {
			found = true
			break
		}
	}

	require.True(found, "EQUALS_KEY=VALUE=VALUE2=VALUE3 should be in output")
}

func TestSourceEnvLongValue(t *testing.T) {
	require := require.New(t)

	// 非常に長い値（10KB以上）を持つ環境変数
	longValue := strings.Repeat("a", 10*1024) // 10KB
	os.Setenv("LONG_VALUE_KEY", longValue)
	defer os.Unsetenv("LONG_VALUE_KEY")

	source := &SourceEnv{}
	reader, err := source.Load()
	require.NoError(err)

	data, err := io.ReadAll(reader)
	require.NoError(err)

	content := string(data)
	expectedLine := "LONG_VALUE_KEY=" + longValue
	require.Contains(content, expectedLine, "Long value should be preserved correctly")
}

func TestSourceEnvUnicodeValue(t *testing.T) {
	require := require.New(t)

	// Unicode 文字を含む環境変数（絵文字、日本語など）
	os.Setenv("UNICODE_KEY", "こんにちは世界🌍🚀日本語テスト")
	defer os.Unsetenv("UNICODE_KEY")

	source := &SourceEnv{}
	reader, err := source.Load()
	require.NoError(err)

	data, err := io.ReadAll(reader)
	require.NoError(err)

	content := string(data)
	lines := strings.Split(content, "\n")

	found := false
	for _, line := range lines {
		if line == "UNICODE_KEY=こんにちは世界🌍🚀日本語テスト" {
			found = true
			break
		}
	}

	require.True(found, "UNICODE_KEY should preserve Unicode characters including emoji and Japanese")
}

func TestSourceEnvSingleQuoteInValue(t *testing.T) {
	require := require.New(t)

	// シングルクォートを含む値（it's のような文字列）
	// Node.js dotenv 互換: シングルクォートはエスケープ不要
	// ただし、スペースを含む場合はクォートされる
	os.Setenv("SINGLE_QUOTE_KEY", "it's a test")
	defer os.Unsetenv("SINGLE_QUOTE_KEY")

	source := &SourceEnv{}
	reader, err := source.Load()
	require.NoError(err)

	data, err := io.ReadAll(reader)
	require.NoError(err)

	content := string(data)
	lines := strings.Split(content, "\n")

	found := false
	for _, line := range lines {
		// スペースを含むので、ダブルクォートで囲まれて出力される
		// シングルクォート自体はエスケープ不要
		if line == `SINGLE_QUOTE_KEY="it's a test"` {
			found = true
			break
		}
	}

	require.True(found, "SINGLE_QUOTE_KEY should be quoted due to space (single quotes preserved)")
}

func TestSourceEnvSingleQuoteNoSpace(t *testing.T) {
	require := require.New(t)

	// スペースを含まないシングルクォートの値はクォート不要
	os.Setenv("SINGLE_QUOTE_NOSPACE", "it's")
	defer os.Unsetenv("SINGLE_QUOTE_NOSPACE")

	source := &SourceEnv{}
	reader, err := source.Load()
	require.NoError(err)

	data, err := io.ReadAll(reader)
	require.NoError(err)

	content := string(data)
	lines := strings.Split(content, "\n")

	found := false
	for _, line := range lines {
		// スペースがないので、クォートなしでそのまま出力される
		if line == "SINGLE_QUOTE_NOSPACE=it's" {
			found = true
			break
		}
	}

	require.True(found, "SINGLE_QUOTE_NOSPACE should not be quoted (no space, no newline, no double quote)")
}
