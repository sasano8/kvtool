package sources

import (
	"os"
	"strings"
	"testing"

	"github.com/sasano8/kvtool/pkg/decoders"
	"github.com/stretchr/testify/require"
)

// TestSourceEnvWithDecoderIntegration は SourceEnv と EnvDecoder の統合テスト
// 改行を含む環境変数が正しくエスケープされ、デコードされることを確認
func TestSourceEnvWithDecoderIntegration(t *testing.T) {
	tests := []struct {
		name     string
		envKey   string
		envValue string
		expected string
	}{
		{
			name:     "改行を含む値",
			envKey:   "TEST_NEWLINE",
			envValue: "line1\nline2\nline3",
			expected: "line1\nline2\nline3",
		},
		{
			name:     "タブを含む値",
			envKey:   "TEST_TAB",
			envValue: "value\twith\ttabs",
			expected: "value\twith\ttabs",
		},
		{
			name:     "バックスラッシュを含む値",
			envKey:   "TEST_BACKSLASH",
			envValue: `path\to\file`,
			expected: `path\to\file`,
		},
		{
			name:     "ダブルクォートを含む値",
			envKey:   "TEST_QUOTE",
			envValue: `value with "quotes"`,
			expected: `value with "quotes"`,
		},
		{
			name:     "複数の特殊文字を含む値",
			envKey:   "TEST_COMPLEX",
			envValue: "line1\nline2\t\"quoted\"",
			expected: "line1\nline2\t\"quoted\"",
		},
		{
			name:     "空の値",
			envKey:   "TEST_EMPTY",
			envValue: "",
			expected: "",
		},
		{
			name:     "値に = を含む",
			envKey:   "TEST_EQUALS",
			envValue: "VALUE=VALUE2=VALUE3",
			expected: "VALUE=VALUE2=VALUE3",
		},
		{
			name:     "Unicode 文字（日本語）",
			envKey:   "TEST_JAPANESE",
			envValue: "こんにちは世界",
			expected: "こんにちは世界",
		},
		{
			name:     "Unicode 文字（絵文字）",
			envKey:   "TEST_EMOJI",
			envValue: "Hello 🌍🚀 World",
			expected: "Hello 🌍🚀 World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			// 環境変数を設定
			os.Setenv(tt.envKey, tt.envValue)
			defer os.Unsetenv(tt.envKey)

			// SourceEnv でデータを取得
			var source Source = &SourceEnv{}
			reader, err := source.Load()
			require.NoError(err)

			// EnvDecoder でデコード
			var decoder decoders.Decoder = &decoders.EnvDecoder{}
			result, err := decoder.Decode(reader)
			require.NoError(err)

			// 結果を確認
			dataMap, ok := result.(map[string]any)
			require.True(ok, "result should be map[string]any")

			// 環境変数が正しくデコードされていることを確認
			actualValue, exists := dataMap[tt.envKey]
			require.True(exists, "key should exist in result")
			require.Equal(tt.expected, actualValue, "value should match expected")
		})
	}
}

// TestSourceEnvWithDecoderLongValue は非常に長い値（10KB以上）の統合テスト
func TestSourceEnvWithDecoderLongValue(t *testing.T) {
	require := require.New(t)

	// 非常に長い値（10KB以上）
	longValue := strings.Repeat("a", 10*1024) // 10KB
	os.Setenv("TEST_LONG_VALUE", longValue)
	defer os.Unsetenv("TEST_LONG_VALUE")

	// SourceEnv でデータを取得
	var source Source = &SourceEnv{}
	reader, err := source.Load()
	require.NoError(err)

	// EnvDecoder でデコード
	var decoder decoders.Decoder = &decoders.EnvDecoder{}
	result, err := decoder.Decode(reader)
	require.NoError(err)

	// 結果を確認
	dataMap, ok := result.(map[string]any)
	require.True(ok, "result should be map[string]any")

	actualValue, exists := dataMap["TEST_LONG_VALUE"]
	require.True(exists, "key should exist in result")
	require.Equal(longValue, actualValue, "long value should be preserved correctly")
	require.Len(actualValue, 10*1024, "value length should be 10KB")
}

// TestSourceEnvInjectionResistance は環境変数インジェクション耐性のテスト
// 改行を含む値が別の KEY=VALUE として解釈されないことを確認
func TestSourceEnvInjectionResistance(t *testing.T) {
	tests := []struct {
		name        string
		envKey      string
		envValue    string
		injectedKey string // インジェクションされないことを確認するキー
	}{
		{
			name:        "改行で KEY=VALUE インジェクション試行",
			envKey:      "SAFE_KEY",
			envValue:    "safe_value\nINJECTED_KEY=malicious_value",
			injectedKey: "INJECTED_KEY",
		},
		{
			name:        "CR+LF で KEY=VALUE インジェクション試行",
			envKey:      "SAFE_KEY2",
			envValue:    "safe_value\r\nINJECTED_KEY2=malicious_value",
			injectedKey: "INJECTED_KEY2",
		},
		{
			name:        "複数改行でのインジェクション試行",
			envKey:      "SAFE_KEY3",
			envValue:    "line1\nINJECT1=bad\nINJECT2=worse\nline4",
			injectedKey: "INJECT1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			// 環境変数を設定（改行を含む悪意のある値）
			os.Setenv(tt.envKey, tt.envValue)
			defer os.Unsetenv(tt.envKey)

			// SourceEnv でデータを取得
			var source Source = &SourceEnv{}
			reader, err := source.Load()
			require.NoError(err)

			// EnvDecoder でデコード（信頼モード）
			var decoder decoders.Decoder = &decoders.EnvDecoder{Trusted: true}
			result, err := decoder.Decode(reader)
			require.NoError(err)

			dataMap, ok := result.(map[string]any)
			require.True(ok, "result should be map[string]any")

			// 元のキーが存在し、改行を含む値全体が保持されていることを確認
			actualValue, exists := dataMap[tt.envKey]
			require.True(exists, "original key should exist")
			require.Equal(tt.envValue, actualValue, "original value should be preserved including newlines")

			// インジェクションされたキーが存在しないことを確認
			_, injected := dataMap[tt.injectedKey]
			require.False(injected, "injected key should NOT exist - injection attempt failed (good!)")
		})
	}
}
