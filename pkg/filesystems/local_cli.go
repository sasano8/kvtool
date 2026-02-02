package filesystems

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

const StdinPath = "@"
const StdoutPath = "@"

// - が習慣という話もある
// const StdinPath = "-"
// const StdoutPath = "-"

type FsLocalFileCli struct {
	Path    string
	NoMkdir bool
}

func (fs *FsLocalFileCli) OpenWriter() (io.WriteCloser, error) {
	if fs.Path == StdoutPath {
		return nopWriteCloser{os.Stdout}, nil
	}

	local := FsLocalFile{Path: fs.Path, NoMkdir: fs.NoMkdir}
	f, err := local.OpenWriter()
	return f, err
}

type DecoderFunc func(r io.Reader) (any, error)

func (fs *FsLocalFileCli) Load(decoder DecoderFunc) (any, error) {
	rc, err := fs.openReader()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	result, err := decoder(rc)
	return result, err
}

func (fs *FsLocalFileCli) LoadAsJson() (any, error) {
	rc, err := fs.openReader()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	dec := json.NewDecoder(rc)
	dec.UseNumber()

	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}

	// 1つ目のJSON値の後にゴミ（2つ目のJSONなど）がないかチェック
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("extra data after first JSON value")
		}
		return nil, err
	}

	return v, nil
}

type readCloser struct {
	io.Reader
	CloseFn func() error
}

func (rc readCloser) Close() error {
	if rc.CloseFn == nil {
		return nil
	}
	return rc.CloseFn()
}

type nopReadCloser struct{ io.Reader }

func (nopReadCloser) Close() error { return nil }

func (fs *FsLocalFileCli) openReader() (io.ReadCloser, error) {
	if fs.Path == StdinPath {
		reader := bufio.NewReader(os.Stdin)
		_ = consumeUTF8BOM(reader)
		return nopReadCloser{Reader: reader}, nil
	}

	f, err := os.Open(fs.Path)
	if err != nil {
		return nil, err
	}

	// ファイルでも BOM 対策（必要なら）
	r := bufio.NewReader(f)
	_ = consumeUTF8BOM(r)
	reader := readCloser{Reader: r, CloseFn: func() error {
		return f.Close()
	}}
	return reader, nil
}

func consumeUTF8BOM(r *bufio.Reader) error {
	b, err := r.Peek(3)
	if err != nil {
		// Peekが失敗しても（空等）問題ないので無視してOK
		return nil
	}
	if len(b) == 3 && b[0] == 0xEF && b[1] == 0xBB && b[2] == 0xBF {
		_, _ = r.Discard(3)
	}
	return nil
}
