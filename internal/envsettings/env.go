package envsettings

import (
	"log/slog"
	"os"
	"strings"
)

const (
	EnvPrometheusLabelID     = "PROMETHEUS_LABEL_ID"
	EnvPrometheusNamespace   = "PROMETHEUS_NAMESPACE"
	EnvPrometheusPushGateway = "PROMETHEUS_PUSH_GATEWAY"

	EnvLogFilePath = "LOG_FILE_PATH"
	EnvLogFormat   = "LOG_FORMAT"
	EnvLogLevel    = "LOG_LEVEL"

	EnvFluentdHost = "FLUENTD_HOST"
	EnvFluentdPort = "FLUENTD_PORT"
)

type Prometheus struct {
	LabelID     string
	Namespace   string
	PushGateway string
}

type Fluentd struct {
	Host string
	Port string
}

func (f Fluentd) Present() bool {
	return f.Host != "" || f.Port != ""
}

type Log struct {
	FilePath string
	Level    string
	Format   string
}

func (l Log) SlogLevel() slog.Level {
	lvl := slog.LevelInfo
	switch strings.ToLower(l.Level) {
	case "panic", "fatal", "error":
		lvl = slog.LevelError
	case "warn", "warning":
		lvl = slog.LevelWarn
	case "debug", "trace":
		lvl = slog.LevelDebug
	}

	return lvl
}

func (l Log) IsFormatJSON() bool {
	return strings.EqualFold(l.Format, "json")
}

type Settings struct {
	Prometheus Prometheus
	Fluentd    Fluentd
	Log        Log
}

func (s *Settings) PrometheusEnabled() bool {
	return s.Prometheus.PushGateway != ""
}

func Get() Settings {
	return Settings{
		Log: Log{
			FilePath: os.Getenv(EnvLogFilePath),
			Level:    os.Getenv(EnvLogLevel),
			Format:   os.Getenv(EnvLogFormat),
		},
		Fluentd: Fluentd{
			Host: os.Getenv(EnvFluentdHost),
			Port: os.Getenv(EnvFluentdPort),
		},
		Prometheus: Prometheus{
			LabelID:     os.Getenv(EnvPrometheusLabelID),
			Namespace:   os.Getenv(EnvPrometheusNamespace),
			PushGateway: os.Getenv(EnvPrometheusPushGateway),
		},
	}
}
