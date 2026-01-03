package registry

import (
	"testing"

	"github.com/sasano8/kvtool/internal/testutil"
)

func TestRegistry(t *testing.T) {
	var v, ok any
	t2 := testutil.New(t)

	reg := CreateRegistry[int]()

	// 要素数が0であること
	t2.AssertEqual(0, len(reg))

	v, ok = reg.Create("key1", 1)
	t2.AssertEqual(1, len(reg))
	t2.AssertEqual(true, ok)
	t2.AssertEqual(1, v)

	v, ok = reg.Get("key1")
	t2.AssertEqual(1, v)

	v, ok = reg.Create("key1", 2)
	t2.AssertEqual(1, len(reg))
	t2.AssertEqual(false, ok)

	// 値が変わっていないこと
	v, ok = reg.Get("key1")
	t2.AssertEqual(1, v)

	ok = reg.Delete("key1")
	t2.AssertEqual(0, len(reg))
	t2.AssertEqual(true, ok)
	v, ok = reg.Get("key1")
	t2.AssertEqual(false, ok)

	ok = reg.Delete("key1")
	t2.AssertEqual(0, len(reg))
	t2.AssertEqual(false, ok)

	v = reg.Put("key1", 1)
	v, ok = reg.Get("key1")
	t2.AssertEqual(1, v)

	v = reg.Put("key1", 2)
	v, ok = reg.Get("key1")
	t2.AssertEqual(2, v)
}
