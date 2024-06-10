package log

import (
	"io"
	"log/slog"
	"strings"
	"time"
)

const logTimeFormat = "2006-01-02T15:04:05.000Z07:00"

type Config struct {
	format string
	level  string
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) WithLevel(level string) *Config {
	c.level = level

	return c
}

func (c *Config) Level() slog.Level {
	lvl := slog.LevelInfo
	switch strings.ToLower(c.level) {
	case "panic", "fatal", "error":
		lvl = slog.LevelError
	case "warn", "warning":
		lvl = slog.LevelWarn
	case "debug", "trace":
		lvl = slog.LevelDebug
	}

	return lvl
}

func (c *Config) WithFormat(format string) *Config {
	c.format = format

	return c
}

func (c Config) IsFormatJSON() bool {
	return strings.EqualFold(c.format, "json")
}

func (c *Config) TextHandlerOptions() *slog.HandlerOptions {
	return &slog.HandlerOptions{
		Level: c.Level(),
	}
}

func (c *Config) JSONHandlerOptions() *slog.HandlerOptions {
	return &slog.HandlerOptions{
		Level: c.Level(),
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.TimeKey:
				a.Key = "@timestamp"
				a.Value = slog.StringValue(time.Now().UTC().Format(logTimeFormat))
			case slog.MessageKey:
				a.Key = "message"
			case slog.LevelKey:
				// This might not be needed, but ensures we have full compatibility with logrus.
				if l, ok := a.Value.Any().(slog.Level); ok {
					switch l {
					case slog.LevelDebug:
						a.Value = slog.StringValue("debug")
					case slog.LevelInfo:
						a.Value = slog.StringValue("info")
					case slog.LevelWarn:
						a.Value = slog.StringValue("warning")
					case slog.LevelError:
						a.Value = slog.StringValue("error")
					}
				}
			}
			return a
		},
	}
}

func NewLogger(output io.Writer, config *Config) *slog.Logger {
	var handler slog.Handler
	if config.IsFormatJSON() {
		handler = slog.NewJSONHandler(output, config.JSONHandlerOptions())
	} else {
		handler = slog.NewTextHandler(output, config.TextHandlerOptions())
	}

	return slog.New(handler)
}
