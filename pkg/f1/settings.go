package f1

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/form3tech-oss/f1/v3/internal/envsettings"
)

// LogFormat specifies the output format for the default logger.
type LogFormat uint8

const (
	// LogFormatText selects plain-text log output (default).
	LogFormatText LogFormat = iota
	// LogFormatJSON selects JSON-structured log output.
	LogFormatJSON
)

const (
	logFormatTextStr = "text"
	logFormatJSONStr = "json"
	logLevelInfoStr  = "info"
)

// String returns "text" or "json".
func (f LogFormat) String() string {
	if f == LogFormatJSON {
		return logFormatJSONStr
	}

	return logFormatTextStr
}

// ParseLogLevel parses a log level string into slog.Level.
// Accepted values (case-insensitive): "debug", "info", "warn", "error".
// Also accepts legacy aliases: "trace" (→ Debug), "warning" (→ Warn),
// "fatal"/"panic" (→ Error). Empty string defaults to Info.
func ParseLogLevel(s string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug", "trace":
		return slog.LevelDebug, nil
	case logLevelInfoStr, "":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error", "fatal", "panic":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("unknown log level %q: use debug, info, warn, or error", s)
	}
}

// ParseLogFormat parses a log format string into LogFormat.
// Accepted values (case-insensitive): "text", "json".
// Empty string defaults to LogFormatText.
func ParseLogFormat(s string) (LogFormat, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case logFormatTextStr, "":
		return LogFormatText, nil
	case logFormatJSONStr:
		return LogFormatJSON, nil
	default:
		return 0, fmt.Errorf("unknown log format %q: use %s or %s", s, logFormatTextStr, logFormatJSONStr)
	}
}

// LoggingSettings configures the default logger built by f1.
// These settings have no effect when WithLogger is used.
type LoggingSettings struct {
	FilePath string
	Level    slog.Level
	Format   LogFormat
}

// PrometheusSettings configures Prometheus metrics push.
type PrometheusSettings struct {
	PushGateway string
	Namespace   string
	LabelID     string
}

// Settings configures f1 infrastructure (logging, Prometheus).
// Use DefaultSettings to obtain the env-backed baseline,
// or construct a zero-value Settings{} to start from scratch.
type Settings struct {
	Prometheus PrometheusSettings
	Logging    LoggingSettings
}

// DefaultSettings returns settings loaded from environment variables.
// This is the baseline used by New when no WithSettings option is provided.
//
// Environment variables read:
//
//	PROMETHEUS_PUSH_GATEWAY, PROMETHEUS_NAMESPACE, PROMETHEUS_LABEL_ID
//	LOG_FILE_PATH, F1_LOG_LEVEL, F1_LOG_FORMAT
func DefaultSettings() Settings {
	es := envsettings.Get()

	return Settings{
		Prometheus: PrometheusSettings{
			PushGateway: es.Prometheus.PushGateway,
			Namespace:   es.Prometheus.Namespace,
			LabelID:     es.Prometheus.LabelID,
		},
		Logging: LoggingSettings{
			FilePath: es.Log.FilePath,
			Level:    es.Log.SlogLevel(),
			Format:   logFormatFromEnv(es.Log.Format),
		},
	}
}

func (s Settings) toInternal() envsettings.Settings {
	var format string
	if s.Logging.Format == LogFormatJSON {
		format = logFormatJSONStr
	}

	return envsettings.Settings{
		Prometheus: envsettings.Prometheus{
			PushGateway: s.Prometheus.PushGateway,
			Namespace:   s.Prometheus.Namespace,
			LabelID:     s.Prometheus.LabelID,
		},
		Log: envsettings.Log{
			FilePath: s.Logging.FilePath,
			Level:    slogLevelToString(s.Logging.Level),
			Format:   format,
		},
	}
}

func logFormatFromEnv(s string) LogFormat {
	if strings.EqualFold(s, logFormatJSONStr) {
		return LogFormatJSON
	}

	return LogFormatText
}

func slogLevelToString(level slog.Level) string {
	switch level {
	case slog.LevelDebug:
		return "debug"
	case slog.LevelInfo:
		return logLevelInfoStr
	case slog.LevelWarn:
		return "warn"
	case slog.LevelError:
		return "error"
	default:
		return logLevelInfoStr
	}
}
