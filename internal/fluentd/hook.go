package fluentd

import (
	"fmt"
	"strconv"

	"github.com/evalphobia/logrus_fluent"
	log "github.com/sirupsen/logrus"

	"github.com/form3tech-oss/f1/v2/internal/logging"
)

func LoggingHook(host, port string) logging.RegisterLogHookFunc {
	return func(scenario string) error {
		if host == "" || port == "" {
			return nil
		}

		portNum, err := strconv.Atoi(port)
		if err != nil {
			return fmt.Errorf("parsing fluentd port: %w", err)
		}

		hook, err := logrus_fluent.NewWithConfig(logrus_fluent.Config{
			Port:                portNum,
			Host:                host,
			DefaultTag:          "f1-" + scenario,
			TagPrefix:           "service",
			DefaultMessageField: "message",
			AsyncConnect:        false,
			MarshalAsJSON:       true,
		})
		if err != nil {
			return fmt.Errorf("creating fluent config: %w", err)
		}

		log.AddHook(hook)
		return nil
	}
}
