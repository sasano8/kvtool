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
		if !isValidPOSIXKeyName(key) {
			return nil, fmt.Errorf("invalid key name (must match POSIX [a-zA-Z_][a-zA-Z0-9_]*): %q", key)
		}

		// quoted value の処理
		if len(val) >= 2 {
			// ダブルクォート: エスケープ付き
			if val[0] == '"' {
				// 終端のダブルクォートを探す
				endIdx := findClosingQuote(val[1:], '"')
				if endIdx >= 0 {
					inner := val[1 : endIdx+1]
					val = unescapeEnvValue(inner)
				}
			} else if val[0] == '\'' {
				// シングルクォート: そのまま（エスケープなし）
				endIdx := findClosingQuote(val[1:], '\'')
				if endIdx >= 0 {
					val = val[1 : endIdx+1]
				}
			} else {
				// クォートなし: 行末コメントを処理
				val = stripInlineComment(val)
			}
		} else {
			// 短い値: 行末コメントを処理
			val = stripInlineComment(val)
		}

		result[key] = val
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// findClosingQuote はクォートの終端位置を探す（エスケープを考慮）
// 戻り値は元の文字列（クォート開始後）における終端クォートの位置
func findClosingQuote(s string, quote byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			i++ // エスケープシーケンスをスキップ
			continue
		}
		if s[i] == quote {
			return i
		}
	}
	return -1 // 終端が見つからない
}

// stripInlineComment はクォートで囲まれていない値から行末コメントを除去する
func stripInlineComment(val string) string {
	idx := strings.Index(val, " #")
	if idx >= 0 {
		return strings.TrimSpace(val[:idx])
	}
	// スペースなしの # もチェック（値の直後）
	idx = strings.Index(val, "#")
	if idx > 0 {
		return strings.TrimSpace(val[:idx])
	}
	return val
}

// isValidPOSIXKeyName は POSIX シェル変数名のルールに従うかチェックする
// ルール: [a-zA-Z_][a-zA-Z0-9_]*
func isValidPOSIXKeyName(key string) bool {
	if len(key) == 0 {
		return false
	}
	// 最初の文字: 英字またはアンダースコア
	first := key[0]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return false
	}
	// 残りの文字: 英字、数字、アンダースコア
	for i := 1; i < len(key); i++ {
		c := key[i]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	return true
}

// unescapeEnvValue はダブルクォート内のエスケープシーケンスを処理する
// サポートするエスケープ: \n (改行), \r (CR), \" (ダブルクォート)
// Node.js dotenv 互換: \t, \\ は非サポート（そのまま残す）
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
			case '"':
				b.WriteByte('"')
			default:
				// 非サポートのエスケープはそのまま残す（Node.js dotenv 互換）
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
