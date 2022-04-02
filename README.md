# zt the zip-transparent Reader

```go
// Package zt provides a Reader that help to transparently handle an incomping
// bytes stream, whether it's compressed (by decompressing it on the fly), or
// uncompressed, in which case it's forwarded as-i.
```

This is especially useful for command line utilities for example, when you want
to transparently deal with data coming from standard input, whether it is:
  - uncompressed
  - compressed in zstandard
  - compressed in gzip


Example of use: a transparent decompressor.

```go
package main

import (
    ...
	"github.com/arl/zt"
)

func main() {
	r, err := zt.NewReader(os.Stdin)
	if err != nil { /* handle error */ }

	n, err := io.Copy(os.Stdout, r)
	if err != nil { /* handle error */ }
}
```

    go run main.go < /some/file.gz
    go run main.go < /some/file.zst
    go run main.go < /some/file