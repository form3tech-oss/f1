package f1

import "github.com/form3tech-oss/f1/v3/internal/envsettings"

// WithPrometheusPushGateway overrides the PROMETHEUS_PUSH_GATEWAY env var.
func WithPrometheusPushGateway(url string) Option {
	return func(f *F1) {
		f.settings.Prometheus.PushGateway = url
	}
}

// WithPrometheusNamespace overrides the PROMETHEUS_NAMESPACE env var.
func WithPrometheusNamespace(ns string) Option {
	return func(f *F1) {
		f.settings.Prometheus.Namespace = ns
	}
}

// WithPrometheusLabelID overrides the PROMETHEUS_LABEL_ID env var.
func WithPrometheusLabelID(id string) Option {
	return func(f *F1) {
		f.settings.Prometheus.LabelID = id
	}
}

// WithLogFilePath overrides the LOG_FILE_PATH env var.
func WithLogFilePath(path string) Option {
	return func(f *F1) {
		f.settings.Log.FilePath = path
	}
}

// WithLogLevel overrides the F1_LOG_LEVEL env var.
// Accepts "debug", "info", "warn", "error" (case-insensitive).
// Has no effect when WithLogger is also used.
func WithLogLevel(level string) Option {
	return func(f *F1) {
		f.settings.Log.Level = level
	}
}

// WithLogFormat overrides the F1_LOG_FORMAT env var.
// Accepts "text" or "json" (case-insensitive).
// Has no effect when WithLogger is also used.
func WithLogFormat(format string) Option {
	return func(f *F1) {
		f.settings.Log.Format = format
	}
}

// WithoutEnvSettings ignores all environment variables; settings start from
// zero values (info level, text format, no prometheus). Must precede other
// settings options in the option list so they are not overwritten.
func WithoutEnvSettings() Option {
	return func(f *F1) {
		f.settings = envsettings.Settings{}
	}
}
