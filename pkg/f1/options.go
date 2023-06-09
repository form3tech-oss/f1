package f1

import log "github.com/sirupsen/logrus"

// Option represents a configuration option. See `f1.WithX` functions.
type Option interface {
	Apply(*F1)
}

type withLoggerOpt struct {
	logger *log.Logger // TODO: have a wrapper around logger to detect TTY and skip colors in such case?
}

func (o *withLoggerOpt) Apply(f1 *F1) {
	f1.logger = o.logger
}

func WithLogger(logger *log.Logger) Option {
	// TODO: drop dependency on logrus?
	return &withLoggerOpt{logger}
}
