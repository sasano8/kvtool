package convert

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// JSONToEnv converts a flat JSON object into .env format.
func JSONToEnv(r io.Reader, w io.Writer) error {
	dec := json.NewDecoder(r)
	var m map[string]interface{}
	if err := dec.Decode(&m); err != nil {
		return err
	}

	for k, v := range m {
		val := fmt.Sprint(v)           // JSON の値を文字列に
		escaped := escapeEnvValue(val) // ダブルクォート用にエスケープ

		if _, err := fmt.Fprintf(w, "%s=\"%s\"\n", k, escaped); err != nil {
			return err
		}
	}
	return nil
}

func EnvToJSON(w io.Writer) error {
	var buf bytes.Buffer

	for _, kv := range os.Environ() {
		if _, err := buf.WriteString(kv + "\n"); err != nil {
			return err
		}
	}
	return DotenvToJSON(&buf, w)
}

// EnvToJSON converts .env-style lines into a JSON object.
func DotenvToJSON(r io.Reader, w io.Writer) error {
	scanner := bufio.NewScanner(r)
	result := make(map[string]string)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		idx := strings.Index(line, "=")
		if idx <= 0 {
			return fmt.Errorf("invalid line: %q", line)
		}

		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		if key == "" {
			return errors.New("empty key")
		}

		// quoted value の処理
		if len(val) >= 2 {
			// ダブルクォート: エスケープ付き
			if val[0] == '"' && val[len(val)-1] == '"' {
				inner := val[1 : len(val)-1]
				val = unescapeEnvValue(inner)
			} else if val[0] == '\'' && val[len(val)-1] == '\'' {
				// シングルクォート: そのまま（エスケープなし）
				val = val[1 : len(val)-1]
			}
		}

		result[key] = val
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
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

// JSONToEnv 側と対になるアンエスケープ
func unescapeEnvValue(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '\\' && i+1 < len(s) {
			n := s[i+1]
			switch n {
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			case 't':
				b.WriteByte('\t')
			case '\\':
				b.WriteByte('\\')
			case '"':
				b.WriteByte('"')
			default:
				// よく分からないエスケープはそのまま残す
				b.WriteByte('\\')
				b.WriteByte(n)
			}
			i++ // 次の1文字を消費済み
			continue
		}
		b.WriteByte(c)
	}
	return b.String()
}
