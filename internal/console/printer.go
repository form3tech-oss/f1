package console

import (
	"fmt"
	"io"
)

type Printer struct {
	Writer    io.Writer
	ErrWriter io.Writer
}

func NewPrinter(writer io.Writer, errWriter io.Writer) *Printer {
	return &Printer{
		Writer:    writer,
		ErrWriter: errWriter,
	}
}

func (t *Printer) Println(a ...any) {
	fmt.Fprintln(t.Writer, a...)
}

func (t *Printer) Printf(format string, a ...any) {
	fmt.Fprintf(t.Writer, format, a...)
}

func (t *Printer) Print(a ...any) {
	fmt.Fprint(t.Writer, a...)
}

func (t *Printer) Warn(a ...any) {
	fmt.Fprint(t.ErrWriter, a...)
}

func (t *Printer) Warnf(format string, a ...any) {
	fmt.Fprintf(t.ErrWriter, format, a...)
}
