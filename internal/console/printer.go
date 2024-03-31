package console

import (
	"fmt"
	"io"
)

type Printer struct {
	writer io.Writer
}

func NewPrinter(writer io.Writer) *Printer {
	return &Printer{
		writer: writer,
	}
}

func (t *Printer) Println(a ...any) {
	fmt.Fprintln(t.writer, a...)
}

func (t *Printer) Printf(format string, a ...any) {
	fmt.Fprintf(t.writer, format, a...)
}

func (t *Printer) Print(a ...any) {
	fmt.Fprint(t.writer, a...)
}
