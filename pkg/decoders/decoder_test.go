package decoders

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnvDecoderImplementsInterface(t *testing.T) {
	// Decoder インターフェースを実装していることを確認
	var _ Decoder = (*EnvDecoder)(nil)
}

func TestEnvDecoderDecode(t *testing.T) {
	require := require.New(t)

	input := strings.NewReader(`KEY1=value1
KEY2=value2
# コメント
KEY3=value3
`)

	// Decoder インターフェースとして使用
	var decoder Decoder = &EnvDecoder{}
	result, err := decoder.Decode(input)
	require.NoError(err)
	require.NotNil(result)

	// map[string]any として型アサーション
	resultMap, ok := result.(map[string]any)
	require.True(ok, "result should be map[string]any")
	require.Equal("value1", resultMap["KEY1"])
	require.Equal("value2", resultMap["KEY2"])
	require.Equal("value3", resultMap["KEY3"])
}

func TestEnvDecoderWithQuotes(t *testing.T) {
	require := require.New(t)

	input := strings.NewReader(`QUOTED="quoted value"
SINGLE='single quoted'
ESCAPED="line1\nline2"
`)

	decoder := &EnvDecoder{}
	result, err := decoder.Decode(input)
	require.NoError(err)

	resultMap := result.(map[string]any)
	require.Equal("quoted value", resultMap["QUOTED"])
	require.Equal("single quoted", resultMap["SINGLE"])
	require.Equal("line1\nline2", resultMap["ESCAPED"])
}

func TestEnvDecoderEmptyLines(t *testing.T) {
	require := require.New(t)

	input := strings.NewReader(`KEY1=value1

KEY2=value2

# コメント

KEY3=value3
`)

	decoder := &EnvDecoder{}
	result, err := decoder.Decode(input)
	require.NoError(err)

	resultMap := result.(map[string]any)
	require.Len(resultMap, 3)
	require.Equal("value1", resultMap["KEY1"])
	require.Equal("value2", resultMap["KEY2"])
	require.Equal("value3", resultMap["KEY3"])
}
