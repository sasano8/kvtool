package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sasano8/kvtool/internal/convert"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "json2env":
		json2envCmd(os.Args[2:])
	case "dotenv2json":
		dotenv2jsonCmd(os.Args[2:])
	case "env2json":
		env2jsonCmd(os.Args[2:])
	case "init":
		initCmd(os.Args[2:])
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
	in := fs.String("i", "", "input file (default: stdin)")
	out := fs.String("o", "", "output file (default: stdout)")
	_ = fs.Parse(args)

	return &ioOpts{
		inPath:  *in,
		outPath: *out,
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
						"path": ".env",
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
