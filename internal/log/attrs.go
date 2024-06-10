package log

import "log/slog"

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
