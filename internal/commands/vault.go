package commands

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
)

func VaultCmd(args []string) {
	fs := flag.NewFlagSet("vault", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	// 出力
	var outPath string
	fs.StringVar(&outPath, "o", "", "output file (default: stdout)")
	fs.StringVar(&outPath, "output", "", "output file (default: stdout)")

	// Vault 接続/認証
	addr := fs.String("addr", "", "Vault address (default: VAULT_ADDR)")
	token := fs.String("token", "", "Vault token (default: VAULT_TOKEN)  ※tokenは可能なら環境変数推奨")
	namespace := fs.String("namespace", "", "Vault namespace (default: VAULT_NAMESPACE) (Enterprise)")

	// KV 指定
	mount := fs.String("mount", "secret", "KV mount path (e.g. secret)")
	pathFlag := fs.String("path", "", "secret path under mount (e.g. app/prod)")
	kvVer := fs.Int("kv", 2, "KV engine version: 1 or 2")
	version := fs.Int("version", 0, "KV v2 version (0=latest)")

	// 出力制御
	field := fs.String("field", "", "output only this key from secret (optional)")
	pretty := fs.Bool("pretty", true, "pretty print JSON")
	timeout := fs.Duration("timeout", 10*time.Second, "request timeout")

	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, `Usage:
  kvtool vault -mount <mount> -path <path> [-kv 1|2] [-version N] [-field KEY] [-o <file>]

Examples:
  # KV v2 (latest) を JSON で
  kvtool vault -mount secret -path app/prod

  # KV v2 の特定バージョン
  kvtool vault -mount secret -path app/prod -kv 2 -version 3

  # 1キーだけ取り出す
  kvtool vault -mount secret -path app/prod -field password

Env:
  VAULT_ADDR, VAULT_TOKEN, VAULT_NAMESPACE, VAULT_* TLS vars are supported by vault/api config.
`)
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}

	secretPath := *pathFlag
	rest := fs.Args()
	if secretPath == "" {
		// -path が無い場合は位置引数 1つを許容
		if len(rest) == 1 {
			secretPath = rest[0]
		} else {
			fs.Usage()
			os.Exit(2)
		}
	} else {
		// -path 指定時に余計な位置引数があればエラー
		if len(rest) != 0 {
			fmt.Fprintln(os.Stderr, "ERROR: too many args")
			os.Exit(2)
		}
	}

	out, err := openOutput(outPath)
	if err != nil {
		exitErr(err)
	}
	defer out.Close()

	data, err := readVaultKV(
		context.Background(),
		*addr,
		*token,
		*namespace,
		*mount,
		secretPath,
		*kvVer,
		*version,
		*timeout,
	)
	if err != nil {
		exitErr(err)
	}

	// -field 指定ならその値だけ
	if *field != "" {
		v, ok := data[*field]
		if !ok {
			exitErr(fmt.Errorf("field %q not found in secret (available keys: %v)", *field, sortedKeys(data)))
		}
		// 文字列は素直に1行、それ以外はJSONで
		switch x := v.(type) {
		case string:
			_, _ = fmt.Fprintln(out, x)
		default:
			enc := json.NewEncoder(out)
			if *pretty {
				enc.SetIndent("", "  ")
			}
			if err := enc.Encode(x); err != nil {
				exitErr(err)
			}
		}
		return
	}

	// Data だけを JSON 出力
	enc := json.NewEncoder(out)
	if *pretty {
		enc.SetIndent("", "  ")
	}
	if err := enc.Encode(data); err != nil {
		exitErr(err)
	}
}

func readVaultKV(
	parent context.Context,
	addr, token, ns, mount, secretPath string,
	kvVer int,
	version int,
	timeout time.Duration,
) (map[string]any, error) {
	cfg := vaultapi.DefaultConfig()
	// VAULT_ADDR や TLS 系環境変数（VAULT_CACERT等）を反映
	_ = cfg.ReadEnvironment()

	client, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("vault client: %w", err)
	}

	if addr != "" {
		if err := client.SetAddress(addr); err != nil {
			return nil, fmt.Errorf("set addr: %w", err)
		}
	}

	// namespace は flag > env
	if ns == "" {
		ns = os.Getenv("VAULT_NAMESPACE")
	}
	if ns != "" {
		client.SetNamespace(ns)
	}

	// token は flag > env（vault/api は token を自動では拾わないので明示）
	if token == "" {
		token = os.Getenv("VAULT_TOKEN")
	}
	if token == "" {
		return nil, fmt.Errorf("missing Vault token: set -token or VAULT_TOKEN")
	}
	client.SetToken(token)

	mount = strings.Trim(mount, "/")
	secretPath = strings.Trim(secretPath, "/")

	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	switch kvVer {
	case 1:
		sec, err := client.KVv1(mount).Get(ctx, secretPath)
		if err != nil {
			return nil, err
		}
		if sec == nil || sec.Data == nil {
			return nil, fmt.Errorf("secret has no data (not found or empty)")
		}
		return toAnyMap(sec.Data), nil

	case 2:
		kv := client.KVv2(mount)
		var sec *vaultapi.KVSecret

		if version > 0 {
			sec, err = kv.GetVersion(ctx, secretPath, version)
		} else {
			sec, err = kv.Get(ctx, secretPath)
		}

		// KVv2 helper が古いVaultで metadata 互換が崩れるケースがあるので、
		// 失敗したら HTTP API の /data を直接叩くフォールバックに落とす（data だけ取る）
		if err != nil {
			raw, err2 := readVaultKVv2Raw(ctx, client, mount, secretPath, version)
			if err2 == nil {
				return raw, nil
			}
			return nil, err
		}

		if sec == nil || sec.Data == nil {
			return nil, fmt.Errorf("secret has no data (deleted or empty)")
		}
		return toAnyMap(sec.Data), nil

	default:
		return nil, fmt.Errorf("invalid -kv %d (must be 1 or 2)", kvVer)
	}
}

// KV v2 の HTTP API を直接叩いて data だけ抜くフォールバック
// v2 の Read API は /<mount>/data/<path> を使う :contentReference[oaicite:1]{index=1}
func readVaultKVv2Raw(ctx context.Context, client *vaultapi.Client, mount, secretPath string, version int) (map[string]any, error) {
	apiPath := fmt.Sprintf("%s/data/%s", strings.Trim(mount, "/"), strings.Trim(secretPath, "/"))
	q := map[string][]string{}
	if version > 0 {
		q["version"] = []string{strconv.Itoa(version)}
	}

	sec, err := client.Logical().ReadWithDataWithContext(ctx, apiPath, q)
	if err != nil {
		return nil, err
	}
	if sec == nil || sec.Data == nil {
		return nil, fmt.Errorf("secret not found")
	}

	// KV v2 のレスポンスは Data["data"] に本体が入る :contentReference[oaicite:2]{index=2}
	rawData, ok := sec.Data["data"].(map[string]interface{})
	if !ok || rawData == nil {
		return nil, fmt.Errorf("unexpected KV v2 response format at %s", apiPath)
	}
	return toAnyMap(rawData), nil
}

func toAnyMap(m map[string]interface{}) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
