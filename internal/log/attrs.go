package log

import (
	"log/slog"
	"time"
)

func ErrorAttr(err error) slog.Attr {
	return ErrorStringAttr(err.Error())
}

func ErrorAnyAttr(err any) slog.Attr {
	return slog.Any("error", err)
}

func ErrorStringAttr(errMessage string) slog.Attr {
	return slog.String("error", errMessage)
}

func StackTraceAttr(stack []byte) slog.Attr {
	return slog.String("stack_trace", string(stack))
}

func ScenarioAttr(scenarioName string) slog.Attr {
	return slog.String("scenario", scenarioName)
}

func IterationAttr(iteration string) slog.Attr {
	return slog.String("iteration", iteration)
}

func DurationAttr(duration time.Duration) slog.Attr {
	return slog.Duration("duration", duration)
}

func IterationStatsGroup(started, successful, failed, dropped uint64, period time.Duration) slog.Attr {
	if started == 0 {
		started = successful + failed + dropped
	}
	return slog.Group("iteration_stats",
		slog.Uint64("started", started),
		slog.Uint64("successful", successful),
		slog.Uint64("failed", failed),
		slog.Uint64("dropped", dropped),
		slog.Duration("period", period),
	)
}
