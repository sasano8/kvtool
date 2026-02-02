package sources

import (
	"os"
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
