package zt

import (
	"bytes"
	_ "embed"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var (
	//go:embed testdata/lorem.txt.golden
	lorem []byte

	//go:embed testdata/lorem.txt.golden.gz
	loremGzip []byte

	//go:embed testdata/lorem.txt.golden.zst
	loremZstd []byte
)

func testReader(r io.Reader) func(t *testing.T) {
	return func(t *testing.T) {
		r, err := NewReader(r)
		if err != nil {
			t.Fatalf("Reader returns %v", err)
		}
		got, err := io.ReadAll(r)
		if err != nil {
			t.Fatalf("couldn't read all: %v", err)
		}

		if !bytes.Equal(got, lorem) {
			t.Errorf("read content doesn't match. Writing buffer to temp file")
			tmpName := strings.ReplaceAll(t.Name(), string(filepath.Separator), "_") + "_"
			f, err := os.CreateTemp("", tmpName)
			if err != nil {
				t.Fatalf("couldn't create temp file: %v", err)
			}
			if _, err := io.Copy(f, bytes.NewReader(got)); err != nil {
				t.Fatalf("couldn't write to temp file: %v", err)
			}
			t.Logf("the content has been written to %q", f.Name())
		}
	}
}

func testReaders(newReader func([]byte) io.Reader) func(t *testing.T) {
	return func(t *testing.T) {
		t.Run("uncompressed", testReader(newReader(lorem)))
		t.Run("gzip", testReader(newReader(loremGzip)))
		t.Run("zstd", testReader(newReader(loremZstd)))
	}
}

func TestReader(t *testing.T) {
	t.Run("bytes", testReaders(func(buf []byte) io.Reader { return bytes.NewReader(buf) }))
}

func TestReader1Byte(t *testing.T) {
	r, err := NewReader(strings.NewReader("0"))
	if err != nil {
		t.Fatal(err)
	}
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != "0" {
		t.Errorf("got b = %q, want %q", b, "0")
	}
}
