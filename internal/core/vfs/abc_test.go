package repository

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCli2Map(t *testing.T) {
	t2 := require.New(t)
	var args []string
	var expect map[string]string
	var result map[string]string
	var err error

	expect, err = Map2NormalizedMap(map[string]any{"a": "1", "-b": "2", "--c": "3", "d": nil})
	fmt.Println(expect)
	args, err = Map2Cli(expect)
	result, err = Cli2Map(args)

	t2.Equal(expect, result)
	t2.Nil(err)
}
