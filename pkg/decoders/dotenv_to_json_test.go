package decoders

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDotenvToJSON(t *testing.T) {
	require := require.New(t)
	input := `
FOO=bar
BAZ="hello world"
# comment
NUM=123
`
	reader := strings.NewReader(input)
	actual, err := DotenvToJson(reader)
	require.Nil(err)
	require.Equal(map[string]any{
		"FOO": "bar",
		"BAZ": "hello world",
		"NUM": "123",
	}, actual)
}

func TestDotenvInlineComment(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]any
	}{
		{
			name:  "行末コメント（スペース付き）",
			input: "KEY=value # this is a comment",
			expected: map[string]any{
				"KEY": "value",
			},
		},
		{
			name:  "行末コメント（スペースなし）",
			input: "KEY=value#comment",
			expected: map[string]any{
				"KEY": "value",
			},
		},
		{
			name:  "ダブルクォート内の # はコメントではない",
			input: `KEY="value # not a comment"`,
			expected: map[string]any{
				"KEY": "value # not a comment",
			},
		},
		{
			name:  "シングルクォート内の # はコメントではない",
			input: `KEY='value # not a comment'`,
			expected: map[string]any{
				"KEY": "value # not a comment",
			},
		},
		{
			name:  "クォート後の行末コメント",
			input: `KEY="value" # comment after quote`,
			expected: map[string]any{
				"KEY": "value",
			},
		},
		{
			name:  "複数行",
			input: "KEY1=value1 # comment1\nKEY2=value2 # comment2",
			expected: map[string]any{
				"KEY1": "value1",
				"KEY2": "value2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			reader := strings.NewReader(tt.input)
			actual, err := EnvToJson(reader)
			require.NoError(err)
			require.Equal(tt.expected, actual)
		})
	}
}

func TestDotenvPOSIXKeyValidation(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		errorMsg    string
	}{
		// 有効なキー名
		{
			name:        "英字で始まるキー",
			input:       "FOO=bar",
			expectError: false,
		},
		{
			name:        "小文字で始まるキー",
			input:       "foo=bar",
			expectError: false,
		},
		{
			name:        "アンダースコアで始まるキー",
			input:       "_FOO=bar",
			expectError: false,
		},
		{
			name:        "アンダースコアのみ",
			input:       "_=bar",
			expectError: false,
		},
		{
			name:        "英字と数字とアンダースコアの組み合わせ",
			input:       "FOO_BAR_123=value",
			expectError: false,
		},
		// 無効なキー名
		{
			name:        "数字で始まるキー",
			input:       "1KEY=value",
			expectError: true,
			errorMsg:    "invalid key name",
		},
		{
			name:        "数字のみのキー",
			input:       "123=value",
			expectError: true,
			errorMsg:    "invalid key name",
		},
		{
			name:        "ハイフンを含むキー",
			input:       "FOO-BAR=value",
			expectError: true,
			errorMsg:    "invalid key name",
		},
		{
			name:        "ドットを含むキー",
			input:       "FOO.BAR=value",
			expectError: true,
			errorMsg:    "invalid key name",
		},
		{
			name:        "スペースを含むキー",
			input:       "FOO BAR=value",
			expectError: true,
			errorMsg:    "invalid key name",
		},
		{
			name:        "日本語を含むキー",
			input:       "キー=value",
			expectError: true,
			errorMsg:    "invalid key name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			reader := strings.NewReader(tt.input)
			result, err := EnvToJson(reader)
			if tt.expectError {
				require.Error(err)
				require.Contains(err.Error(), tt.errorMsg)
				require.Nil(result)
			} else {
				require.NoError(err)
				require.NotNil(result)
			}
		})
	}
}

func TestDotenvSingleQuoteInValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]any
	}{
		{
			name:  "シングルクォートを含むクォートなし値（スペースなし）",
			input: "KEY=it's",
			expected: map[string]any{
				"KEY": "it's",
			},
		},
		{
			name:  "ダブルクォートで囲まれたシングルクォート",
			input: `KEY="it's a test"`,
			expected: map[string]any{
				"KEY": "it's a test",
			},
		},
		{
			name:  "ダブルクォートで囲まれた複数のシングルクォート",
			input: `KEY="it's John's book"`,
			expected: map[string]any{
				"KEY": "it's John's book",
			},
		},
		{
			name:  "シングルクォートで囲まれた値（リテラル）",
			input: `KEY='it'\''s a test'`,
			// シングルクォート内はリテラルなので、最初の ' で終了
			// 'it' の後に残りがある形になるが、現在の実装では最初の終端クォートまで
			expected: map[string]any{
				"KEY": "it",
			},
		},
		{
			name:  "シングルクォートのみの値",
			input: "KEY='",
			expected: map[string]any{
				"KEY": "'",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			reader := strings.NewReader(tt.input)
			actual, err := EnvToJson(reader)
			require.NoError(err)
			require.Equal(tt.expected, actual)
		})
	}
}

func TestDotenvUTF8BOM(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected map[string]any
	}{
		{
			name:  "BOM 付きファイル",
			input: []byte{0xEF, 0xBB, 0xBF, 'F', 'O', 'O', '=', 'b', 'a', 'r'},
			expected: map[string]any{
				"FOO": "bar",
			},
		},
		{
			name:  "BOM なしファイル",
			input: []byte("FOO=bar"),
			expected: map[string]any{
				"FOO": "bar",
			},
		},
		{
			name:  "BOM 付き複数行",
			input: append([]byte{0xEF, 0xBB, 0xBF}, []byte("FOO=bar\nBAZ=qux")...),
			expected: map[string]any{
				"FOO": "bar",
				"BAZ": "qux",
			},
		},
		{
			name:  "BOM 付きクォート値",
			input: append([]byte{0xEF, 0xBB, 0xBF}, []byte(`KEY="hello world"`)...),
			expected: map[string]any{
				"KEY": "hello world",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			reader := bytes.NewReader(tt.input)
			actual, err := EnvToJson(reader)
			require.NoError(err)
			require.Equal(tt.expected, actual)
		})
	}
}

func TestDotenvNULLCharacter(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "NULL 文字を含む値（クォートなし）",
			input:       "KEY=hello\x00world",
			expectError: true,
			errorMsg:    "NULL character",
		},
		{
			name:        "NULL 文字を含む値（ダブルクォート）",
			input:       "KEY=\"hello\x00world\"",
			expectError: true,
			errorMsg:    "NULL character",
		},
		{
			name:        "NULL 文字を含む値（シングルクォート）",
			input:       "KEY='hello\x00world'",
			expectError: true,
			errorMsg:    "NULL character",
		},
		{
			name:        "NULL 文字のみの値",
			input:       "KEY=\x00",
			expectError: true,
			errorMsg:    "NULL character",
		},
		{
			name:        "NULL 文字を含まない値",
			input:       "KEY=helloworld",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			reader := strings.NewReader(tt.input)
			result, err := EnvToJson(reader)
			if tt.expectError {
				require.Error(err)
				require.Contains(err.Error(), tt.errorMsg)
				require.Nil(result)
			} else {
				require.NoError(err)
				require.NotNil(result)
			}
		})
	}
}

func TestDotenvUnquotedSpaceError(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		errorMsg string
	}{
		{
			name:     "クォートなしでスペースを含む値",
			input:    "KEY=hello world",
			errorMsg: "unquoted value contains space",
		},
		{
			name:     "クォートなしで複数スペースを含む値",
			input:    "KEY=hello world today",
			errorMsg: "unquoted value contains space",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			reader := strings.NewReader(tt.input)
			result, err := EnvToJson(reader)
			require.Error(err)
			require.Contains(err.Error(), tt.errorMsg)
			require.Nil(result)
		})
	}
}

func TestDotenvQuotedSpaceOK(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]any
	}{
		{
			name:  "ダブルクォートでスペースを含む値",
			input: `KEY="hello world"`,
			expected: map[string]any{
				"KEY": "hello world",
			},
		},
		{
			name:  "シングルクォートでスペースを含む値",
			input: `KEY='hello world'`,
			expected: map[string]any{
				"KEY": "hello world",
			},
		},
		{
			name:  "複数のスペースを含む値",
			input: `KEY="hello   world"`,
			expected: map[string]any{
				"KEY": "hello   world",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			reader := strings.NewReader(tt.input)
			actual, err := EnvToJson(reader)
			require.NoError(err)
			require.Equal(tt.expected, actual)
		})
	}
}

func TestEnvToJsonWithTrusted(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]any
	}{
		{
			name:  "非POSIXキー名が信頼モードで許可される",
			input: "1KEY=value",
			expected: map[string]any{
				"1KEY": "value",
			},
		},
		{
			name:  "クォートなしスペースが信頼モードで許可される",
			input: "KEY=hello world",
			expected: map[string]any{
				"KEY": "hello world",
			},
		},
		{
			name:  "NULL文字が信頼モードで許可される",
			input: "KEY=hello\x00world",
			expected: map[string]any{
				"KEY": "hello\x00world",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			reader := strings.NewReader(tt.input)
			actual, err := EnvToJson(reader, WithTrusted())
			require.NoError(err)
			require.Equal(tt.expected, actual)
		})
	}
}

func TestEnvToJsonTrustedStillParses(t *testing.T) {
	require := require.New(t)

	// 信頼モードでもパース処理は正常に動作することを確認
	input := `KEY1="hello world"
KEY2='literal $value'
KEY3=simple`
	reader := strings.NewReader(input)
	actual, err := EnvToJson(reader, WithTrusted())
	require.NoError(err)
	require.Equal(map[string]any{
		"KEY1": "hello world",
		"KEY2": "literal $value",
		"KEY3": "simple",
	}, actual)
}
