package encoders

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

type ObjToDotenvEncoder struct {
}

func (encoder *ObjToDotenvEncoder) Marshal(data any) ([]byte, error) {
	obj, ok := data.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected JSON object (map[string]any), got %T", data)
	}

	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Buffer は []bytes のラッパーみたいな感じ
	var buf bytes.Buffer

	for _, k := range keys {
		val := fmt.Sprint(obj[k])
		escaped := escapeEnvValue(val)
		if _, err := fmt.Fprintf(&buf, "%s=\"%s\"\n", k, escaped); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

// .env のダブルクォート値として安全になるようにエスケープ
func escapeEnvValue(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString(`\\`)
		case '"':
			b.WriteString(`\"`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
