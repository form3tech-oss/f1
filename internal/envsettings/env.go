package envsettings

import (
	"os"
)

const (
	EnvPrometheusLabelID     = "PROMETHEUS_LABEL_ID"
	EnvPrometheusNamespace   = "PROMETHEUS_NAMESPACE"
	EnvPrometheusPushGateway = "PROMETHEUS_PUSH_GATEWAY"

	EnvLogFilePath = "LOG_FILE_PATH"

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

type Settings struct {
	Prometheus  Prometheus
	Fluentd     Fluentd
	LogFilePath string
}

func (s *Settings) PrometheusEnabled() bool {
	return s.Prometheus.PushGateway != ""
}

func Get() Settings {
	return Settings{
		LogFilePath: os.Getenv(EnvLogFilePath),
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
