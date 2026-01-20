package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// package １ディレクトリ＝１パッケージ

type StoreConfig struct {
	Version float64 `json:"version"`
	// Namespaces map[string]map[string]Store `json:"namespaces"`
}

type KwargsOptions struct {
	AllowDashes bool // key が "--x" のように来てもOKにするか（通常は先頭の-を剥がす）
	AllowDup    bool // 同じkeyが複数回出たとき、上書き許可するか
}

func Map2NormalizedMap[T any](m map[string]T) (map[string]string, error) {
	normalizeKey := func(k string) string {
		k = strings.TrimSpace(k)
		k = strings.TrimLeft(k, "-") // "--a" -> "a", "-a" -> "a"
		return k
	}

	// keep=false のときは「キーを削除（出力に入れない）」
	normalizeValue := func(v any) (s string, keep bool, err error) {
		isNilLike := func(v any) bool {
			if v == nil {
				return true
			}
			rv := reflect.ValueOf(v)
			switch rv.Kind() {
			case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.Interface, reflect.Slice:
				return rv.IsNil()
			default:
				return false
			}
		}

		if isNilLike(v) {
			return "", false, nil // ★ nil は drop
		}

		// JSON互換スカラーを優先
		switch x := v.(type) {
		case string:
			return x, true, nil
		case bool:
			return strconv.FormatBool(x), true, nil
		case int:
			return strconv.Itoa(x), true, nil
		case int8, int16, int32, int64:
			return strconv.FormatInt(reflect.ValueOf(x).Int(), 10), true, nil
		case uint, uint8, uint16, uint32, uint64, uintptr:
			return strconv.FormatUint(reflect.ValueOf(x).Uint(), 10), true, nil
		case float32:
			return strconv.FormatFloat(float64(x), 'g', -1, 32), true, nil
		case float64:
			return strconv.FormatFloat(x, 'g', -1, 64), true, nil
		case json.Number:
			return x.String(), true, nil
		}

		// pointer/interface を剥がす（型付きnil対策も兼ねる）
		rv := reflect.ValueOf(v)
		for rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
			if rv.IsNil() {
				return "", false, nil // ★ nil は drop
			}
			rv = rv.Elem()
		}

		switch rv.Kind() {
		case reflect.Map, reflect.Slice, reflect.Array, reflect.Struct:
			return "", false, fmt.Errorf("non-scalar value type %T", v)
		default:
			return fmt.Sprint(rv.Interface()), true, nil
		}
	}

	out := make(map[string]string, len(m))

	for rawK, rawV := range m {
		k := normalizeKey(rawK)
		if k == "" {
			return nil, fmt.Errorf("empty key after normalize: %q", rawK)
		}

		s, keep, err := normalizeValue(any(rawV))
		if err != nil {
			return nil, fmt.Errorf("key %q: %w", k, err)
		}
		if !keep {
			continue // ★ nil は出力しない（キー削除）
		}

		if _, exists := out[k]; exists {
			return nil, fmt.Errorf("duplicate key after normalize: %q (from %q)", k, rawK)
		}
		out[k] = s
	}
	return out, nil
}

func Map2Cli(m map[string]string) ([]string, error) {
	if m == nil {
		return nil, nil
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		if strings.TrimSpace(k) == "" {
			return nil, fmt.Errorf("empty key")
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	args := make([]string, 0, len(keys)*2)
	for _, k := range keys {
		// Map2NormalizedMap で key は正規化済み、value も string 化済み
		args = append(args, "--"+k, m[k])
	}
	return args, nil
}

func Cli2Map(args []string) (map[string]string, error) {
	opt := KwargsOptions{
		AllowDup: false,
	}
	if opt.AllowDup {
		panic("Not allow option")
	}

	out := make(map[string]string, len(args)/2)

	set := func(k, v string) error {
		if !opt.AllowDup {
			if _, exists := out[k]; exists {
				return fmt.Errorf("duplicate key: %q", k)
			}
		}
		out[k] = v
		return nil
	}

	normalizeKey := func(k string) string {
		if opt.AllowDashes {
			return k
		}
		// "--key" / "-k" みたいなのを key として使うなら先頭の "-" を剥がす
		return strings.TrimLeft(k, "-")
	}

	for i := 0; i < len(args); i++ {
		a := args[i]

		// 1) key=value / --key=value
		if eq := strings.IndexByte(a, '='); eq >= 0 {
			k := strings.TrimSpace(a[:eq])
			v := a[eq+1:] // 値はそのまま（空文字も許可）
			k = normalizeKey(k)
			if k == "" {
				return nil, fmt.Errorf("empty key in %q", a)
			}
			if err := set(k, v); err != nil {
				return nil, err
			}
			continue
		}

		// 2) --key value / key value（次要素を値にする）
		k := normalizeKey(strings.TrimSpace(a))
		if k == "" {
			return nil, fmt.Errorf("empty key at args[%d]=%q", i, a)
		}
		if i+1 >= len(args) {
			return nil, fmt.Errorf("missing value for key %q", k)
		}
		v := args[i+1]
		i++ // value を消費

		if err := set(k, v); err != nil {
			return nil, err
		}
	}

	return out, nil
}

func Map2Struct[T any](m map[string]any) (T, error) {
	var out T

	// map -> JSON -> struct 変換（json tag が効く）
	b, err := json.Marshal(m)
	if err != nil {
		return out, fmt.Errorf("marshal map: %w", err)
	}

	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields() // 不要なら外してOK
	if err := dec.Decode(&out); err != nil {
		return out, fmt.Errorf("unmarshal into %T: %w", out, err)
	} else {
		return out, nil
	}
}
