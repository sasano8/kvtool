package repository

import (
	"testing"

	"github.com/sasano8/kvtool/internal/core/testutil"
)

func TestRepository(t *testing.T) {
	var v any
	var err error

	t2 := testutil.New(t)

	reg := New[int]()

	// 要素数が0であること
	t2.AssertEqual(0, len(reg))

	v, err = reg.Create("key1", 1)
	t2.AssertEqual(1, len(reg))
	t2.IsNil(err)
	t2.AssertEqual(1, v)

	v, err = reg.Get("key1")
	t2.AssertEqual(1, v)

	v, err = reg.Create("key1", 2)
	t2.AssertEqual(1, len(reg))
	t2.IsError(err)

	// 値が変わっていないこと
	v, err = reg.Get("key1")
	t2.AssertEqual(1, v)

	err = reg.Delete("key1")
	t2.AssertEqual(0, len(reg))
	t2.IsNil(err)
	v, err = reg.Get("key1")
	t2.IsError(err)

	err = reg.Delete("key1")
	t2.AssertEqual(0, len(reg))
	t2.IsError(err)

	v, err = reg.Put("key1", 1)
	v, err = reg.Get("key1")
	t2.AssertEqual(1, v)

	v, err = reg.Put("key1", 2)
	v, err = reg.Get("key1")
	t2.AssertEqual(2, v)
}
