package convert

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
		val := fmt.Sprint(v)
		// simple quoting: if contains spaces or '#', wrap in double quotes
		if strings.ContainsAny(val, " #") {
			val = `"` + escapeDoubleQuotes(val) + `"`
		}
		if _, err := fmt.Fprintf(w, "%s=%s\n", k, val); err != nil {
			return err
		}
	}
	return nil
}

// EnvToJSON converts .env-style lines into a JSON object.
func EnvToJSON(r io.Reader, w io.Writer) error {
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

		// Strip quotes if quoted
		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') ||
				(val[0] == '\'' && val[len(val)-1] == '\'') {
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

func escapeDoubleQuotes(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}
