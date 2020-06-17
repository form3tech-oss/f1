package generic

// mockBuffer implements stringWriter interface to allow the rendering engine
// determine the length of the result string without memory allocations.
type mockBuffer int

func (buf *mockBuffer) Len() int {
	return int(*buf)
}

func (buf *mockBuffer) WriteString(s string) (n int, err error) {
	n = len(s)
	*buf += mockBuffer(n)

	return n, nil
}

func (buf *mockBuffer) WriteANSISequence(s string) (n int, err error) {
	return len(s), nil
}
