package decoders

import (
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
