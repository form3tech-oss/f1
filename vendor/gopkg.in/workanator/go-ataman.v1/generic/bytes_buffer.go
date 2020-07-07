package generic

import "bytes"

// bytesBuffer is a proxy for bytes.Buffer.
type bytesBuffer struct {
	bytes.Buffer
}

func (buf *bytesBuffer) Len() int {
	return buf.Buffer.Len()
}

func (buf *bytesBuffer) WriteString(s string) (n int, err error) {
	return buf.Buffer.WriteString(s)
}

func (buf *bytesBuffer) WriteANSISequence(s string) (n int, err error) {
	return buf.Buffer.WriteString(s)
}
