package ui

import (
	"fmt"
	"log/slog"

	"github.com/form3tech-oss/f1/v2/internal/log"
)

type ErrorMessage struct {
	Error   error
	Message string
}

var _ Outputable = (*ErrorMessage)(nil)

func (m ErrorMessage) Print(printer *Printer) {
	printer.Error(fmt.Errorf("%s: %w", m.Message, m.Error).Error())
}

func (m ErrorMessage) Log(logger *slog.Logger) {
	logger.Error(m.Message, log.ErrorAttr(m.Error))
}

type WarningMessage struct {
	Message string
}

var _ Outputable = (*WarningMessage)(nil)

func (m WarningMessage) Print(printer *Printer) {
	printer.Warn(m.Message)
}

func (m WarningMessage) Log(logger *slog.Logger) {
	logger.Warn(m.Message)
}

var _ Outputable = (*InteractiveMessage)(nil)

type InteractiveMessage struct {
	Message string
}

func (m InteractiveMessage) Print(printer *Printer) {
	printer.Println(m.Message)
}

func (m InteractiveMessage) Log(*slog.Logger) {
}

var _ Outputable = (*InfoMessage)(nil)

type InfoMessage struct {
	Message string
}

func (m InfoMessage) Print(printer *Printer) {
	printer.Println(m.Message)
}

func (m InfoMessage) Log(logger *slog.Logger) {
	logger.Info(m.Message)
}
