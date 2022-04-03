package zt

import (
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"

	"github.com/klauspost/compress/zstd"
)

// Minimum number of bytes required for compression detection.
const minBytes = 4

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
	buf := make([]byte, minBytes)
	n, err := io.ReadAtLeast(r, buf, minBytes)
	switch {
	case err == io.EOF || n < len(buf):
		return io.NopCloser(bytes.NewReader(buf[:n])), nil
	case err != nil:
		return nil, fmt.Errorf("zt.NewReader: error from underlying reader: %v", err)
	}

	return newReader(r, buf)
}

func newReader(r io.Reader, buf []byte) (rc io.ReadCloser, err error) {
	comp := detectCompression(buf)

	// Create a reader prefilled with the header we've already had to read in
	// order to detect compression.
	r = &prefilledReader{
		r:   r,
		hdr: buf,
	}

	switch comp {
	case gzipCompression:
		// This already returns a ReadCloser.
		rc, err = gzip.NewReader(r)
		if err != nil {
			return nil, fmt.Errorf("zt.NewReader: error from underlying gzip reader: %v", err)
		}
	case zstdCompression:
		zr, err := zstd.NewReader(r)
		if err != nil {
			return nil, fmt.Errorf("zt.NewReader: error from underlying zstd reader: %v", err)
		}
		// zr is not a ReadCloser since zr.Close doesn't return an error, so we
		// make an actual ReadCloser out of it.
		rc = newReadCloser(zr, func() error { zr.Close(); return nil })
	case bzip2Compression:
		// bzip2.NewReader returns a simple Reader.
		rc = io.NopCloser(bzip2.NewReader(r))
	case noCompression:
		rc = io.NopCloser(r)
	}

	return rc, nil
}

type compressionType int

const (
	noCompression compressionType = iota
	zstdCompression
	gzipCompression
	bzip2Compression
)

func detectCompression(buf []byte) compressionType {
	const (
		gzipMagic  = "\x1f\x8b"
		zstdMagic  = "\x28\xb5\x2f\xfd"
		bzip2Magic = "BZh"
	)

	if bytes.Equal(buf[:2], []byte(gzipMagic)) {
		return gzipCompression
	}
	if bytes.Equal(buf[:4], []byte(zstdMagic)) {
		return zstdCompression
	}
	if bytes.Equal(buf[:3], []byte(bzip2Magic)) {
		// Check compression level
		if buf[3] >= '1' && buf[3] <= '9' {
			return bzip2Compression
		}
	}

	return noCompression
}

type prefilledReader struct {
	r   io.Reader
	hdr []byte
	off int // track offset for next read on 'hdr'
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
