package commands

import (
	"fmt"
	"io"
	"os"
)

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
