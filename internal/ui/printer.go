package ui

import (
	"fmt"
	"io"
	"os"

	"github.com/mattn/go-isatty"
)

type Printer struct {
	Writer      io.Writer
	ErrWriter   io.Writer
	Interactive bool
}

func NewDefaultPrinter() *Printer {
	return NewPrinter(os.Stdout, os.Stderr, isatty.IsTerminal(os.Stdin.Fd()))
}

func NewPrinter(writer io.Writer, errWriter io.Writer, interactive bool) *Printer {
	return &Printer{
		Writer:      writer,
		ErrWriter:   errWriter,
		Interactive: interactive,
	}
}

func (t *Printer) Println(a ...any) {
	fmt.Fprintln(t.Writer, a...)
}

func (t *Printer) Error(a ...any) {
	fmt.Fprintln(t.ErrWriter, a...)
}

func (t *Printer) Printf(format string, a ...any) {
	fmt.Fprintf(t.Writer, format, a...)
}

func (t *Printer) Warn(a ...any) {
	fmt.Fprint(t.ErrWriter, a...)
}

func (t *Printer) Warnf(format string, a ...any) {
	fmt.Fprintf(t.ErrWriter, format, a...)
}
