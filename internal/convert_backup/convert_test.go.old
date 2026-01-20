package convert

import (
	"bytes"
	"strings"
	"testing"
)

func TestJSONToEnv(t *testing.T) {
	input := `{
		"FOO": "bar",
		"BAZ": "hello world",
		"NUM": 123
	}`

	var buf bytes.Buffer
	if err := JSONToEnv(strings.NewReader(input), &buf); err != nil {
		t.Fatalf("JSONToEnv error: %v", err)
	}

	out := buf.String()

	// map の順序は未定なので Contains で判定するのは OK
	if !strings.Contains(out, "FOO=\"bar\"\n") {
		t.Errorf("expected line FOO=\"bar\", got:\n%s", out)
	}
	if !strings.Contains(out, "BAZ=\"hello world\"\n") {
		t.Errorf("expected line BAZ=\"hello world\", got:\n%s", out)
	}
	if !strings.Contains(out, "NUM=\"123\"\n") {
		t.Errorf("expected line NUM=\"123\", got:\n%s", out)
	}
}

func TestDotenvToJSON(t *testing.T) {
	input := `
FOO=bar
BAZ="hello world"
# comment
NUM=123
`

	var buf bytes.Buffer
	if err := DotenvToJSON(strings.NewReader(input), &buf); err != nil {
		t.Fatalf("EnvToJSON error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, `"FOO": "bar"`) {
		t.Errorf("expected FOO in json, got:\n%s", out)
	}
	if !strings.Contains(out, `"BAZ": "hello world"`) {
		t.Errorf("expected BAZ in json, got:\n%s", out)
	}
	if !strings.Contains(out, `"NUM": "123"`) {
		t.Errorf("expected NUM in json, got:\n%s", out)
	}
}
