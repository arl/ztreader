# *zt*: the transparent `io.Reader` for compressed data

Package *zt* provides a type implementing the `io.ReadCloser` interface,
that transparently uncompresses a stream of compressed bytes. *zt* detects
the compression algorithm from the stream header, creates that appropriate
decompressor. In case the incoming data is not compressed, or if the compression
algorithm is unknown or unsupported, bytes are simply forwarded as-is.

Supported compression algorithms are:
  - [Zstandard](https://github.com/facebook/zstd)
  - [Gzip](https://www.gzip.org/)
  - [Bzip2](https://en.wikipedia.org/wiki/Bzip2)
  - [zlib](https://www.zlib.net/)

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
