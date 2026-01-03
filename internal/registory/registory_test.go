package registory

import (
	"testing"

	"github.com/sasano8/kvtool/internal/core/testutil"
)

func TestRegistory(t *testing.T) {
	t2 := testutil.New(t)

	t2.Assert(Commands != nil)

	_, ok := Commands.Get("VaultCmd")
	t2.Assert(ok)
}
