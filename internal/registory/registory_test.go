package registory

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegistory(t *testing.T) {
	require := require.New(t)
	// t2 := testutil.New(t)

	require.NotNil(Commands)

	// t2.Assert(Commands != nil)

	_, err := Commands.Get("VaultCmd")
	require.NotNil(err)
	// t2.Assert(ok)
}
