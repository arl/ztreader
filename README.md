# *zt*: the transparent `io.Reader` for compressed data

Package *zt* provides a type implementing the `io.ReadCloser` interface,
that allows to transparently handle an incoming stream of bytes, whether
it's compressed – by decompressing it on the fly – or uncompressed, in
which case the bytes are simply forwarded as-is.

Supported compression algorithms are:
  - [Zstandard](https://github.com/facebook/zstd)
  - [Gzip](https://www.gzip.org/)
  - [Bzip2](https://en.wikipedia.org/wiki/Bzip2)
  - [zlib](https://www.zlib.net/)

One example of use is for CLI programs that transparently support reading from
standard input data, whether it's compressed or not, and without requiring the
user to specify the compression algorithm.

#### Example, a transparent decompressor.

```go
package main

import (
	"io"
	"os"

	"github.com/arl/zt"
)

func main() {
	r, err := zt.NewReader(os.Stdin)
	if err != nil { /* handle error */ }
	defer r.Close()

	_, err = io.Copy(os.Stdout, r)
	if err != nil { /* handle error */ }
}
```

```sh
go run main.go < /some/file.gz  # decompress gzip-compressed file to stdout
go run main.go < /some/file.zst # decompress zstandard-compressed file to stdout
go run main.go < /some/file.bz2 # decompress bzip2-compressed file to stdout
go run main.go < /some/file.zz  # decompress zlib-compressed file to stdout
go run main.go < /some/file     # print non-compressed file to stdout
```
