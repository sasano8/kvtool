package main

import (
	"flag"
	"fmt"
	"io"
	"os"

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

func exitErr(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
