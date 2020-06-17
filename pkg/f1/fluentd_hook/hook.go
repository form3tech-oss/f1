package fluentd_hook

import (
	"fmt"
	"os"
	"strconv"

	"github.com/evalphobia/logrus_fluent"
	log "github.com/sirupsen/logrus"
)

func AddFluentdLoggingHook(scenario string) {
	host := os.Getenv("FLUENTD_HOST")
	port := os.Getenv("FLUENTD_PORT")
	if host == "" || port == "" {
		return
	}
	portNum, err := strconv.Atoi(port)
	if err != nil {
		log.WithError(err).Error("unable to parse fluentd port")
		return
	}
	hook, err := logrus_fluent.NewWithConfig(logrus_fluent.Config{
		Port:                portNum,
		Host:                host,
		DefaultTag:          fmt.Sprintf("f1-%s", scenario),
		TagPrefix:           "service",
		DefaultMessageField: "message",
		AsyncConnect:        false,
		MarshalAsJSON:       true,
	})
	if err != nil {
		panic(err)
	}

	log.AddHook(hook)
}
