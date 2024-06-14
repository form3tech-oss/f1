package log

import "log/slog"

type Loggable interface {
	Log(logger *slog.Logger)
}
