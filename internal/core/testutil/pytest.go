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

func (t *pytest) IsError(actual error) {
	if actual == nil {
		t.Helper()
		t.Fatalf("expect=error actual=%v", actual)
	}
}

func (t *pytest) IsNil(actual any) {
	if actual != nil {
		t.Helper()
		t.Fatalf("expect=nil actual=%v", actual)
	}
}

func (t *pytest) Assert(actual bool) {
	if !actual {
		t.Helper()
		t.Fatalf("expect=true actual=%v", actual)
	}
}

// func (t *pytest) AssertNot(actual bool) {
// 	if actual {
// 		t.Helper()
// 		t.Fatalf("expect=false actual=%v", actual)
// 	}
// }

func (t *pytest) AssertEqual(expect, actual any) {
	if !reflect.DeepEqual(expect, actual) {
		t.Helper() // トレース情報を呼び出し元にする
		t.Fatalf("expect=%v actual=%v", expect, actual)
	}
}
