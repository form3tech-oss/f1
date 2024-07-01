package run

import (
	"log/slog"

	"github.com/form3tech-oss/f1/v2/internal/ui"
)

type Output struct {
	logger    *slog.Logger
	printer   *ui.Printer
	logToFile bool
}

var _ ui.Outputer = (*Output)(nil)

func (o *Output) Display(outputable ui.Outputable) {
	if o.printer.Interactive && o.logToFile {
		outputable.Print(o.printer)
		return
	}

	outputable.Log(o.logger)
}

func (o *Output) Logger() *slog.Logger {
	return o.logger
}

func (o *Output) Printer() *ui.Printer {
	return o.printer
}

func NewOutput(logger *slog.Logger, printer *ui.Printer, logToFile bool) *Output {
	return &Output{
		logger:    logger,
		printer:   printer,
		logToFile: logToFile,
	}
}
