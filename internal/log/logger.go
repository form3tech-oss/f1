package log

import (
	"io"
	"log/slog"
	"os"
)

func NewLogger(output io.Writer, config *Config) *slog.Logger {
	var handler slog.Handler
	if config.IsFormatJSON() {
		handler = slog.NewJSONHandler(output, config.JSONHandlerOptions())
	} else {
		handler = slog.NewTextHandler(output, config.TextHandlerOptions())
	}

	return slog.New(handler)
}

func NewConsoleLogger(config *Config) *slog.Logger {
	return NewLogger(os.Stdout, config)
}

func NewDiscardLogger() *slog.Logger {
	return NewLogger(io.Discard, NewConfig())
}

func NewTestLogger(writer io.Writer) *slog.Logger {
	opts := &slog.HandlerOptions{
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			// remove time key for deterministic test assertions
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	}

	handler := slog.NewTextHandler(writer, opts)

	return slog.New(handler)
}
