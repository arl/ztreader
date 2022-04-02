package zt

import (
	"bytes"
	"fmt"
	"io"

	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zstd"
)

const (
	// Magic numbers for the supported compressors.
	gzipHeader = "\x1f\x8b"
	zstdHeader = "\x28\xb5\x2f\xfd"
)

// NewReader returns an io.ReadCloser that reads from r, whether r is a reader
// over compressed data or not. The returned Reader decodes compressed data
// (currently supporting gzip and zst), or fowards data as-is if it's not
// compressed, or compressed using a non supported algorithm.
//
// Note: NewReader is an utility provided as a best effort, it's certainly
// possible to trick it, intentionally or not, into thinking a stream of bytes
// contains zstd or gzip compressed data, while in fact it's not.
func NewReader(r io.Reader) (io.ReadCloser, error) {
	var hdr [4]byte
	n, err := r.Read(hdr[:])
	switch {
	case err == io.EOF || n < len(hdr):
		return io.NopCloser(bytes.NewReader(hdr[:n])), nil
	case err != nil:
		return nil, fmt.Errorf("zt.NewReader: error from underlying reader: %v", err)
	}

	var rc io.ReadCloser

	// Prefill a reader with the header containing the magic number
	pfr := newPrefilledReader(r, hdr[:])
	switch {
	case bytes.Equal(hdr[:2], []byte(gzipHeader)):
		r, err := gzip.NewReader(pfr)
		if err != nil {
			return nil, fmt.Errorf("zt.NewReader: error from underlying gzip reader: %v", err)
		}
		rc = r
	case bytes.Equal(hdr[:4], []byte(zstdHeader)):
		r, err := zstd.NewReader(pfr)
		if err != nil {
			return nil, fmt.Errorf("zt.NewReader: error from underlying zstd reader: %v", err)
		}
		rc = newReadCloser(r, func() error { r.Close(); return nil })
	default:
		rc = io.NopCloser(pfr)
	}

	return rc, nil
}

type prefilledReader struct {
	r   io.Reader
	hdr []byte
	off int // track offset for next read on 'hdr'
}

// newPrefilledReader returns an io.Reader that first reads the provided header,
// after what successive calls to Read are forwarded to r.
func newPrefilledReader(r io.Reader, hdr []byte) *prefilledReader {
	return &prefilledReader{
		r:   r,
		hdr: hdr,
	}
}

func (r *prefilledReader) readFromHdr(p []byte) (n int, err error) {
	n = copy(p, r.hdr[r.off:])
	r.off += n
	switch {
	case r.off < len(r.hdr):
		return n, nil
	case r.off == len(r.hdr):
		// Now that the whole header has been read, forward next calls to r.
		r.hdr = nil
		return n, nil
	}
	panic(fmt.Sprintf("unexpected n=%d r.off=%d len(header)=%d", n, r.off, len(r.hdr)))
}

func (r *prefilledReader) Read(p []byte) (n int, err error) {
	if r.hdr != nil {
		return r.readFromHdr(p)
	}

	return r.r.Read(p)
}

type readCloser struct {
	io.Reader
	close func() error
}

// newReadCloser makes a io.ReadCloser from a reader and a close function.
func newReadCloser(r io.Reader, close func() error) io.ReadCloser {
	return &readCloser{Reader: r, close: close}
}

func (rc *readCloser) Close() error {
	if err := rc.close(); err != nil {
		return fmt.Errorf("zt.Reader: error closing: %v", err)
	}
	return nil
}
