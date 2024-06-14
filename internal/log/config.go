package log

import (
	"log/slog"
	"time"
)

const logTimeFormat = "2006-01-02T15:04:05.000Z07:00"

type Config struct {
	json  bool
	level slog.Level
}

func NewConfig() *Config {
	return &Config{
		level: slog.LevelInfo,
	}
}

func (c *Config) WithLevel(level slog.Level) *Config {
	c.level = level

	return c
}

func (c *Config) WithJSONFormat(enabled bool) *Config {
	c.json = enabled
	return c
}

func (c Config) IsFormatJSON() bool {
	return c.json
}

func (c *Config) TextHandlerOptions() *slog.HandlerOptions {
	return &slog.HandlerOptions{
		Level: c.level,
	}
}

func (c *Config) JSONHandlerOptions() *slog.HandlerOptions {
	return &slog.HandlerOptions{
		Level: c.level,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.TimeKey:
				a.Key = "@timestamp"
				a.Value = slog.StringValue(time.Now().UTC().Format(logTimeFormat))
			case slog.MessageKey:
				a.Key = "message"
			case slog.LevelKey:
				// logrus compatibility
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
