# zt the zip-transparent Reader

Package *zt* provides types and functions that allow to transparently handle an
incoming stream of bytes, whether it's compressed – by decompressing it on the
fly – or uncompressed, in which case the bytes are simply forwarded as-is.

One example of use is for CLI programs that wants to support reading compressed
data from standard input. By using *zt* your program can transparently read
compressed and uncompressed data.

Currently supported compression algorithms are:
  - zstandard
  - gzip


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

go run main.go < /some/file.gz
go run main.go < /some/file.zst
go run main.go < /some/file
