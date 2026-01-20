package decoders

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

// EnvToJson converts .env-style lines into a JSON object.
func EnvToJson(r io.Reader) (map[string]any, error) {
	scanner := bufio.NewScanner(r) // １行ずつ返す
	result := make(map[string]any)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		idx := strings.Index(line, "=")
		if idx <= 0 {
			return nil, fmt.Errorf("invalid line: %q", line)
		}

		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		if key == "" {
			return nil, errors.New("empty key")
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
		return nil, err
	}
	return result, nil
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
