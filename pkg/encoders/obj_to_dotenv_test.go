package encoders

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJsonToDotenv(t *testing.T) {
	require := require.New(t)

	encoder := ObjToDotenvEncoder{}
	actual, err := encoder.Marshal(map[string]any{
		"FOO": "bar",
		"BAZ": "hello world",
		"NUM": "123",
	})

	require.Nil(err)
	require.Equal(`BAZ="hello world"
FOO="bar"
NUM="123"
`, string(actual))

}
