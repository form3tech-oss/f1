package logging

// TODO: describe overall design

// TODO: implement in `pkg/f1-logrus`

// TODO: use https://github.com/mattn/go-isatty for internal verbose StdLogger implementation?

type Logger interface {
	IsTTY() bool
	WithField(key string, value interface{}) Logger
	WithError(err error) Logger

	Debug(args ...interface{}) // TODO: simplify args to `msg string`?
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
}

type MultiLogger struct {
	loggers []Logger
}

func (l MultiLogger) Info(args ...interface{}) {
	for _, logger := range l.loggers {
		logger.Info(args...)
	}
}

// PrettyInfo logs either pretty or nonPretty version of a log message, depending on Logger.IsTTY()
func (l MultiLogger) PrettyInfo(pretty, nonPretty string) {
	for _, logger := range l.loggers {
		if logger.IsTTY() {
			logger.Info(pretty)
		} else {
			logger.Info(nonPretty)
		}
	}
}
