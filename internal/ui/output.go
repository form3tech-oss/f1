package ui

import (
	"log/slog"
	"os"

	"github.com/mattn/go-isatty"

	"github.com/form3tech-oss/f1/v2/internal/log"
)

// Outputable may be a type of message (like [ErrorMessage], [InfoMessage], etc) or
// [github.com/form3tech-oss/f1/v2/internal/run/views.ViewContext]
type Outputable interface {
	Print(printer *Printer)
	Log(logger *slog.Logger)
}

type Output struct {
	Logger        *slog.Logger
	Printer       *Printer
	Interactive   bool
	AllowPrinting bool
}

func NewOutput(logger *slog.Logger, printer *Printer, interactive bool, allowPrinting bool) *Output {
	return &Output{
		Logger:        logger,
		Printer:       printer,
		Interactive:   interactive,
		AllowPrinting: allowPrinting,
	}
}

func (o *Output) Display(outputable Outputable) {
	if o.AllowPrinting && o.Interactive {
		outputable.Print(o.Printer)
		return
	}

	outputable.Log(o.Logger)
}

func NewDiscardOutput() *Output {
	printer := NewDiscardPrinter()
	logger := log.NewDiscardLogger()

	return NewOutput(logger, printer, false, false)
}

func NewDefaultOutput(logLevel slog.Level, jsonFormat bool) *Output {
	printer := NewDefaultPrinter()

	config := log.NewConfig().WithLevel(logLevel).WithJSONFormat(jsonFormat)
	logger := log.NewConsoleLogger(config)

	interactive := isatty.IsTerminal(os.Stdin.Fd())

	return NewOutput(logger, printer, interactive, true)
}
