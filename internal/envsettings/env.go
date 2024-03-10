package envsettings

import (
	"os"
)

const (
	EnvPrometheusLabelID     = "PROMETHEUS_LABEL_ID"
	EnvPrometheusNamespace   = "PROMETHEUS_NAMESPACE"
	EnvPrometheusPushGateway = "PROMETHEUS_PUSH_GATEWAY"

	EnvLogFilePath = "LOG_FILE_PATH"

	EnvTrace             = "TRACE"
	EnvTraceEnabledValue = "true"

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
	LogFilePath string

	Trace bool

	Fluentd    Fluentd
	Prometheus Prometheus
}

func Get() Settings {
	return Settings{
		LogFilePath: os.Getenv(EnvLogFilePath),
		Trace:       os.Getenv(EnvTrace) == EnvTraceEnabledValue,
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
