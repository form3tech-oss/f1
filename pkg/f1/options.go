package f1

import (
	"log/slog"

	"github.com/form3tech-oss/f1/v3/internal/ui"
)

// Option configures an F1 instance at construction.
type Option func(*F1)

// WithSettings replaces the settings baseline entirely. By default, settings
// are loaded from environment variables (see DefaultSettings). Pass Settings{}
// to start from zero values and ignore all environment variables.
//
// Individual field options (WithLogLevel, WithPrometheusPushGateway, etc.)
// still apply when placed after WithSettings in the option list.
func WithSettings(s Settings) Option {
	return func(f *F1) {
		f.settings = s
	}
}

// WithLogger specifies the logger for internal and scenario logs.
// When used, logging settings (WithLogLevel, WithLogFormat, F1_LOG_LEVEL,
// F1_LOG_FORMAT) have no effect because the caller controls the logger.
func WithLogger(logger *slog.Logger) Option {
	return func(f *F1) {
		f.options.output = ui.NewDefaultOutputWithLogger(logger)
		f.options.loggerExplicit = true
	}
}

// WithStaticMetrics registers additional labels with fixed values for f1 metrics.
func WithStaticMetrics(labels map[string]string) Option {
	return func(f *F1) {
		f.options.staticMetrics = labels
	}
}

// WithPrometheusPushGateway sets the Prometheus push gateway URL,
// overriding the PROMETHEUS_PUSH_GATEWAY environment variable.
func WithPrometheusPushGateway(url string) Option {
	return func(f *F1) {
		f.settings.Prometheus.PushGateway = url
	}
}

// WithPrometheusNamespace sets the Prometheus namespace label,
// overriding the PROMETHEUS_NAMESPACE environment variable.
func WithPrometheusNamespace(ns string) Option {
	return func(f *F1) {
		f.settings.Prometheus.Namespace = ns
	}
}

// WithPrometheusLabelID sets the Prometheus label ID,
// overriding the PROMETHEUS_LABEL_ID environment variable.
func WithPrometheusLabelID(id string) Option {
	return func(f *F1) {
		f.settings.Prometheus.LabelID = id
	}
}

// WithLogFilePath sets the log file path,
// overriding the LOG_FILE_PATH environment variable.
func WithLogFilePath(path string) Option {
	return func(f *F1) {
		f.settings.Logging.FilePath = path
	}
}

// WithLogLevel sets the log level for the default logger.
// Has no effect when WithLogger is also used.
func WithLogLevel(level slog.Level) Option {
	return func(f *F1) {
		f.settings.Logging.Level = level
	}
}

// WithLogFormat sets the log output format for the default logger.
// Has no effect when WithLogger is also used.
func WithLogFormat(format LogFormat) Option {
	return func(f *F1) {
		f.settings.Logging.Format = format
	}
}
