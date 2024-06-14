package log

import (
	"context"
	"io"
	"log/slog"

	"github.com/sirupsen/logrus"
)

var _ logrus.Hook = (*slogHook)(nil)

// NewSlogLogrusLogger returns a logrus.Logger that will use slog as logging backend.
func NewSlogLogrusLogger(logger *slog.Logger) *logrus.Logger {
	l := logrus.New()
	l.AddHook(newSlogHook(logger))
	l.SetOutput(io.Discard)

	return l
}

// slogHook converts logurs entries to slog
//
// This is needed for backwards compatibility with externally exposed logrus logger.
type slogHook struct {
	logger *slog.Logger
}

func newSlogHook(logger *slog.Logger) *slogHook {
	return &slogHook{
		logger: logger,
	}
}

func (*slogHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *slogHook) Fire(entry *logrus.Entry) error {
	level := convertLevel(entry.Level)
	msg := entry.Message

	fields := make([]slog.Attr, 0, len(entry.Data))
	for k, v := range entry.Data {
		fields = append(fields, slog.Any(k, v))
	}

	h.logger.LogAttrs(context.Background(), level, msg, fields...)
	return nil
}

func convertLevel(l logrus.Level) slog.Level {
	switch l {
	case logrus.TraceLevel, logrus.DebugLevel:
		return slog.LevelDebug
	case logrus.InfoLevel:
		return slog.LevelInfo
	case logrus.WarnLevel:
		return slog.LevelWarn
	case logrus.ErrorLevel:
		return slog.LevelError
	case logrus.FatalLevel, logrus.PanicLevel:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
