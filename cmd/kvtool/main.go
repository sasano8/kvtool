package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/sasano8/kvtool/internal/commands"
	"github.com/sasano8/kvtool/internal/convert"
)

type cliCommand struct {
	run  func([]string)
	help string
}

var _commands = map[string]cliCommand{
	"json":        {run: jsonCmd, help: "JSON -> JSON"},
	"env2json":    {run: env2jsonCmd, help: "env -> JSON"},
	"dotenv2json": {run: dotenv2jsonCmd, help: ".env -> JSON"},
	"json2env":    {run: json2envCmd, help: "JSON -> .env"},
	"init":        {run: initCmd, help: "init config"},
	"store":       {run: storeCmd, help: "load config and dispatch store"},
	"vault":       {run: commands.VaultCmd, help: "Vault KV -> JSON (data only)"},
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "json":
		jsonCmd(os.Args[2:])
	case "vault":
		_commands["vault"].run(os.Args[2:])
	case "json2env":
		json2envCmd(os.Args[2:])
	case "dotenv2json":
		dotenv2jsonCmd(os.Args[2:])
	case "env2json":
		env2jsonCmd(os.Args[2:])
	case "init":
		initCmd(os.Args[2:])
	case "store":
		storeCmd(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `kvtool - key-value conversion tool

Usage:
  kvtool <command> [options]

Commands:
  json          JSON -> JSON
  env2json      env -> JSON
  dotenv2json   .env -> JSON
  json2env      JSON -> .env
  init
  store

Run "kvtool <command> -h" for command options.
`)
}

type ioOpts struct {
	inPath  string
	outPath string
}

func parseIOFlags(args []string, name string) (*ioOpts, []string) {
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	var inPath string
	fs.StringVar(&inPath, "i", "", "input file (default: stdin)")
	fs.StringVar(&inPath, "input", "", "input file (default: stdin)")

	var outPath string
	fs.StringVar(&outPath, "o", "", "output file (default: stdout)")
	fs.StringVar(&outPath, "output", "", "output file (default: stdout)")

	_ = fs.Parse(args)

	return &ioOpts{
		inPath:  inPath,
		outPath: outPath,
	}, fs.Args()
}

func openInput(path string) (io.ReadCloser, error) {
	if path == "" {
		return io.NopCloser(os.Stdin), nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return f, nil
}

type nopWriteCloser struct{ io.Writer }

func (nwc nopWriteCloser) Close() error { return nil }

func exitErr(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}

func openOutput(path string) (io.WriteCloser, error) {
	if path == "" {
		return nopWriteCloser{os.Stdout}, nil
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func jsonCmd(args []string) {
	ioOpts, _ := parseIOFlags(args, "catjson")

	in, err := openInput(ioOpts.inPath)
	if err != nil {
		exitErr(err)
	}
	defer in.Close()

	out, err := openOutput(ioOpts.outPath)
	if err != nil {
		exitErr(err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		exitErr(err)
	}
}

func json2envCmd(args []string) {
	ioOpts, _ := parseIOFlags(args, "json2env")

	in, err := openInput(ioOpts.inPath)
	if err != nil {
		exitErr(err)
	}
	defer in.Close()

	out, err := openOutput(ioOpts.outPath)
	if err != nil {
		exitErr(err)
	}
	defer out.Close()

	if err := convert.JSONToEnv(in, out); err != nil {
		exitErr(err)
	}
}

func dotenv2jsonCmd(args []string) {
	ioOpts, _ := parseIOFlags(args, "dotenv2json")

	in, err := openInput(ioOpts.inPath)
	if err != nil {
		exitErr(err)
	}
	defer in.Close()

	out, err := openOutput(ioOpts.outPath)
	if err != nil {
		exitErr(err)
	}
	defer out.Close()

	if err := convert.DotenvToJSON(in, out); err != nil {
		exitErr(err)
	}
}

func env2jsonCmd(args []string) {
	ioOpts, _ := parseIOFlags(args, "env2json")
	out, err := openOutput(ioOpts.outPath)
	if err != nil {
		exitErr(err)
	}
	defer out.Close()

	if err := convert.EnvToJSON(out); err != nil {
		exitErr(err)
	}
}

type StoreConfig struct {
	Version    float64                     `json:"version"`
	Namespaces map[string]map[string]Store `json:"namespaces"`
}
type Store struct {
	Type string         `json:"type"`
	Args map[string]any `json:"args"`
}

func initCmd(args []string) {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	outPath := fs.String("out", ".kvtool.json", "output file path (e.g. ./config.json)")
	pretty := fs.Bool("pretty", true, "pretty print JSON")
	force := fs.Bool("force", false, "overwrite if file already exists")

	// エラーメッセージを自前にするなら fs.Usage を上書き
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: mytool init -out <path> [-pretty]")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}

	if *outPath == "" {
		fs.Usage()
		os.Exit(2)
	}

	if !*force {
		if _, err := os.Stat(*outPath); err == nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s already exists (use -force to overwrite)\n", *outPath)
			os.Exit(1)
		} else if !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "ERROR: stat %s: %v\n", *outPath, err)
			os.Exit(1)
		}
	}

	payload := StoreConfig{
		Version: 0.1,
		Namespaces: map[string]map[string]Store{
			"default": {
				".env": {
					Type: ".env",
					Args: map[string]any{
						"input": ".env",
					},
				},
			},
		},
	}

	if err := writeJSONFileAtomic(*outPath, payload, *pretty); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}

	fmt.Println("wrote:", *outPath)
}

func writeJSONFileAtomic(path string, v any, pretty bool) error {
	// 親ディレクトリ作成
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", dir, err)
		}
	}

	// JSON生成
	var (
		b   []byte
		err error
	)
	if pretty {
		b, err = json.MarshalIndent(v, "", "  ")
	} else {
		b, err = json.Marshal(v)
	}
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	b = append(b, '\n')

	// atomic write（同一FS上なら rename は原子的）
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return fmt.Errorf("write tmp %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename to %s: %w", path, err)
	}
	return nil
}

func storeCmd(args []string) {
	fs := flag.NewFlagSet("read", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	configPath := fs.String("config", ".kvtool.json", "config file path")
	ns := fs.String("ns", "default", "namespace name")

	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, `Usage: mytool read -config <path> [-ns default] [-store ".env"]`)
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}

	rest := fs.Args()
	var storeKey string
	if len(rest) >= 1 {
		storeKey = rest[0]
	}
	if len(rest) >= 2 {
		fmt.Fprintln(os.Stderr, "ERROR: too many args. expected at most 1 storeKey")
		os.Exit(2)
	}

	cfg, err := loadConfig(*configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}

	k, store, err := getStoreKV(cfg, *ns, storeKey)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}

	dispatchStore(k, store)
}

func dispatchStore(storeKey string, st Store) {
	switch st.Type {
	case "env":
		env2jsonCmd(mapToFlagArgs(st.Args))
	case ".env":
		dotenv2jsonCmd(mapToFlagArgs(st.Args))
	case "vault":
		// _commands["vault"].run(mapToFlagArgs(st.Args))  // 循環参照になってしまう
		commands.VaultCmd(mapToFlagArgs(st.Args))
	// case "json":
	default:
		fmt.Fprintf(os.Stderr, "unknown store type: %q (key=%q)\n", st.Type, storeKey)
		os.Exit(1)
	}
}

func loadConfig(path string) (StoreConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return StoreConfig{}, fmt.Errorf("read %s: %w", path, err)
	}

	var cfg StoreConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
		return StoreConfig{}, fmt.Errorf("parse json %s: %w", path, err)
	}
	if cfg.Namespaces == nil {
		cfg.Namespaces = map[string]map[string]Store{}
	}
	return cfg, nil
}

// nsName が空なら "default" 扱い、storeKey が空なら「その namespace に1件だけならそれを採用」
func getStoreKV(cfg StoreConfig, nsName, storeKey string) (string, Store, error) {
	if nsName == "" {
		nsName = "default"
	}

	ns, ok := cfg.Namespaces[nsName]
	if !ok {
		return "", Store{}, fmt.Errorf("namespace %q not found", nsName)
	}
	if len(ns) == 0 {
		return "", Store{}, fmt.Errorf("namespace %q has no stores", nsName)
	}

	// storeKey 指定があるならそれを取りに行く
	if storeKey != "" {
		st, ok := ns[storeKey]
		if !ok {
			return "", Store{}, fmt.Errorf("store %q not found in namespace %q", storeKey, nsName)
		}
		return storeKey, st, nil
	}

	// storeKey 未指定なら「1件だけならそれを採用」、複数ならエラー
	if len(ns) == 1 {
		for k, st := range ns {
			return k, st, nil
		}
	}

	return "", Store{}, errors.New(`multiple stores exist; specify -store (e.g. -store ".env")`)
}

func mapToFlagArgs(m map[string]any) []string {
	if len(m) == 0 {
		return nil
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys) // あるとデバッグが楽

	out := make([]string, 0, len(m)*2)
	for _, k := range keys {
		v := m[k]
		flagName := "-" + k

		switch x := v.(type) {
		case bool:
			// bool は -flag=true/false が楽
			out = append(out, fmt.Sprintf("%s=%t", flagName, x))
		case string:
			out = append(out, flagName, x)
		case float64:
			// 0.1 なども来るのでそのまま文字列化
			out = append(out, flagName, fmt.Sprint(x))
		case nil:
			// 無視
		default:
			// map/array 等は JSON 文字列として渡す（受け側で解釈するなら）
			b, _ := json.Marshal(x)
			out = append(out, flagName, string(b))
		}
	}
	return out
}
