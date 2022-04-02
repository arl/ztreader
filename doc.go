// Package zt provides a Reader that transparently decodes a input stream of
// compressed bytes, without specifying the compression algorithm. If the stream
// is not compressed, or compressed with an unknown algorithm, the input stream
// is forwarded as-is.
package zt
