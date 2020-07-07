package generic

// stringWriter abstracts string building. The interface is used interbally
// to
type stringWriter interface {
	Len() int
	WriteString(string) (int, error)
	WriteANSISequence(string) (int, error)
}
