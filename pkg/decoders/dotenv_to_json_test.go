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
