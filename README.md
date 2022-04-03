# *zt*: the transparent `io.Reader` for compressed data

Package *zt* provides types and functions that allow to transparently handle an
incoming stream of bytes, whether it's compressed – by decompressing it on the
fly – or uncompressed, in which case the bytes are simply forwarded as-is.

One example of use is for CLI programs that wants to support reading compressed
data from standard input. By using *zt* your program can transparently read
compressed and uncompressed data.

Currently supported compression algorithms are:
  - [Zstandard](https://github.com/facebook/zstd)
  - [Gzip](https://www.gzip.org/)
  - [Bzip2](https://en.wikipedia.org/wiki/Bzip2)
  - [zlib](https://www.zlib.net/)


Example program: a transparent decompressor.

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