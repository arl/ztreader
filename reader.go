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

// NewReader returns an io.ReadCloser that reads from r, either decoding the
// compressed stream (in case the algorithm used to compress it is both
// supported and detected), or just forwarding it.
//
// In order for NewReader to succeeds, at least 4 bytes must be read from r, in
// order to detect the presence of a compression algorithm and its type.
//
// Note: this utility provided as a best effort, it's certainly possible to
// trick it into thinking a stream of bytes contains zstd or gzip compressed
// data, while in fact it's not.
func NewReader(r io.Reader) (io.ReadCloser, error) {
	buf := make([]byte, 4)
	n, err := io.ReadAtLeast(r, buf, 4)
	switch {
	case err == io.EOF || n < len(buf):
		return io.NopCloser(bytes.NewReader(buf[:n])), nil
	case err != nil:
		return nil, fmt.Errorf("zt.NewReader: error from underlying reader: %v", err)
	}

	var rc io.ReadCloser

	// Prefill a reader with the header containing the magic number.
	pfr := newPrefilledReader(r, buf[:])
	switch {
	case bytes.Equal(buf[:2], []byte(gzipHeader)):
		r, err := gzip.NewReader(pfr)
		if err != nil {
			return nil, fmt.Errorf("zt.NewReader: error from underlying gzip reader: %v", err)
		}
		rc = r
	case bytes.Equal(buf[:4], []byte(zstdHeader)):
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

func (r *prefilledReader) Read(p []byte) (n int, err error) {
	if r.hdr != nil {
		n = copy(p, r.hdr[r.off:])
		r.off += n
		switch {
		case r.off < len(r.hdr):
			return n, nil
		case r.off == len(r.hdr):
			// Now that the header has been read, forward next calls to r.
			r.hdr = nil
			return n, nil
		}
		panic(fmt.Sprintf("unexpected n=%d r.off=%d len(header)=%d", n, r.off, len(r.hdr)))
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
