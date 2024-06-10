package envsettings

import (
	"os"
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

type Settings struct {
	Prometheus  Prometheus
	Fluentd     Fluentd
	LogFilePath string
	LogLevel    string
	LogFormat   string
}

func (s *Settings) PrometheusEnabled() bool {
	return s.Prometheus.PushGateway != ""
}

func Get() Settings {
	return Settings{
		LogFilePath: os.Getenv(EnvLogFilePath),
		LogLevel:    os.Getenv(EnvLogLevel),
		LogFormat:   os.Getenv(EnvLogFormat),
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
