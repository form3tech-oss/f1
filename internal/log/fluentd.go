package log

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/fluent/fluent-logger-golang/fluent"
	slogfluentd "github.com/samber/slog-fluentd/v2"
)

func fluentdHandler(host, port, scenario string) (slog.Handler, error) {
	if host == "" || port == "" {
		return nil, nil
	}

	portNum, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("parsing fluentd port: %w", err)
	}

	client, err := fluent.New(fluent.Config{
		FluentHost:    host,
		FluentPort:    portNum,
		FluentNetwork: "tcp",
		MarshalAsJSON: true,
		AsyncConnect:  false,
		TagPrefix:     "service",
	})
	if err != nil {
		return nil, fmt.Errorf("creating fluentd client: %w", err)
	}

	opts := slogfluentd.Option{
		Client: client,
		Tag:    "f1-" + scenario,
	}

	return opts.NewFluentdHandler(), nil
}
