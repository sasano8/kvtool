package repository

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRepository(t *testing.T) {
	var v any
	var err error

	// r と assert がある。
	// r は即時中断、assert は bool を返す。
	r := require.New(t)

	reg := New[int]()

	// 要素数が0であること
	r.Equal(0, len(reg))

	v, err = reg.Create("key1", 1)
	r.Equal(1, len(reg))
	r.Nil(err)
	r.Equal(1, v)

	v, err = reg.Get("key1")
	r.Equal(1, v)

	v, err = reg.Create("key1", 2)
	r.Equal(1, len(reg))
	r.Error(err)

	// 値が変わっていないこと
	v, err = reg.Get("key1")
	r.Equal(1, v)

	err = reg.Delete("key1")
	r.Equal(0, len(reg))
	r.Nil(err)
	v, err = reg.Get("key1")
	r.Error(err)

	err = reg.Delete("key1")
	r.Equal(0, len(reg))
	r.Error(err)

	v, err = reg.Put("key1", 1)
	v, err = reg.Get("key1")
	r.Equal(1, v)

	v, err = reg.Put("key1", 2)
	v, err = reg.Get("key1")
	r.Equal(2, v)
}
