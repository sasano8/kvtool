package encoders

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJsonToJson(t *testing.T) {
	require := require.New(t)
	var actual map[string]string

	enc := ObjToJsonEncoder{}
	expect := map[string]string{
		"1": "a",
		"2": "b",
		"3": "c",
	}
	actual_bytes, err := enc.Marshal(expect)
	require.Nil(err)
	require.True(json.Valid(actual_bytes))

	err = json.Unmarshal(actual_bytes, &actual)
	require.Nil(err)

	require.Equal(expect, actual)
}
