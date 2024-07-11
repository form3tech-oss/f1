package ui

import (
	"fmt"
	"io"
	"os"
)

type Printer struct {
	Writer    io.Writer
	ErrWriter io.Writer
}

func NewDefaultPrinter() *Printer {
	return NewPrinter(os.Stdout, os.Stderr)
}

func NewDiscardPrinter() *Printer {
	return NewPrinter(io.Discard, io.Discard)
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

func (t *Printer) Error(a ...any) {
	fmt.Fprintln(t.ErrWriter, a...)
}

func (t *Printer) Warn(a ...any) {
	fmt.Fprintln(t.ErrWriter, a...)
}
