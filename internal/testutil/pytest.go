package testutil

import (
	"reflect"
	"testing"
)

type pytest struct {
	*testing.T
}

func New(t *testing.T) *pytest {
	return &pytest{t}
}

func (t *pytest) AssertEqual(expect, actual any) {
	if !reflect.DeepEqual(expect, actual) {
		t.Helper() // トレース情報を呼び出し元にする
		t.Fatalf("expect=%v actual=%v", expect, actual)
	}
}
