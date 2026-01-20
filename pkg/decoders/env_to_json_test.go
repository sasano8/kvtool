package decoders

import (
	"strings"
	"testing"

	"github.com/sasano8/kvtool/pkg/sources"
	"github.com/stretchr/testify/require"
)

func TestEnvToJsonFromSource(t *testing.T) {
	var err error
	require := require.New(t)

	source := sources.SourceEnv{}
	r, err := source.Load()
	require.Nil(err)

	actual, err := EnvToJson(r)
	require.Nil(err)
	require.IsType(map[string]any{}, actual)
}

func TestEnvToJSON(t *testing.T) {
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
