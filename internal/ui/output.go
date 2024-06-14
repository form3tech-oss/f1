package ui

import (
	"log/slog"

	"github.com/form3tech-oss/f1/v2/internal/log"
)

type Outputable interface {
	Print(printer *Printer)
	Log(logger *slog.Logger)
}

type Outputer interface {
	Display(outputable Outputable)
	Logger() *slog.Logger
	Printer() *Printer
}

var _ Outputer = (*Output)(nil)

type Output struct {
	logger  *slog.Logger
	printer *Printer
}

func (o *Output) Display(outputable Outputable) {
	if o.printer.Interactive {
		outputable.Print(o.printer)
		return
	}

	outputable.Log(o.logger)
}

func (o *Output) Logger() *slog.Logger {
	return o.logger
}

func (o *Output) Printer() *Printer {
	return o.printer
}

var _ Outputer = (*ConsoleOutput)(nil)

type ConsoleOutput struct {
	printer *Printer
}

func (o *ConsoleOutput) Display(outputable Outputable) {
	outputable.Print(o.printer)
}

func (o *ConsoleOutput) Logger() *slog.Logger {
	return nil
}

func (o *ConsoleOutput) Printer() *Printer {
	return o.printer
}

func NewConoleOnlyOutput() *ConsoleOutput {
	printer := NewDefaultPrinter()

	return &ConsoleOutput{printer: printer}
}

func NewDiscardOutput() *Output {
	printer := NewDefaultPrinter()
	logger := log.NewDiscardLogger()

	return NewOutput(logger, printer)
}

func NewDefaultOutput(logLevel slog.Level, jsonFormat bool) *Output {
	printer := NewDefaultPrinter()

	config := log.NewConfig().WithLevel(logLevel).WithJSONFormat(jsonFormat)
	logger := log.NewConsoleLogger(config)

	return NewOutput(logger, printer)
}

func NewOutput(logger *slog.Logger, printer *Printer) *Output {
	return &Output{
		logger:  logger,
		printer: printer,
	}
}
