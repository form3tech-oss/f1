package run_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"

	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	log "github.com/sirupsen/logrus"

	"github.com/form3tech-oss/f1/v2/internal/support/errorh"
)

type FakePrometheus struct {
	server         *http.Server
	Port           interface{}
	hasMetrics     int32
	metricFamilies sync.Map
}

func (f *FakePrometheus) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	parseGroupedLabels := func() []*io_prometheus_client.LabelPair {
		// labels added via push.Grouping are passed through URI
		// example: /metrics/job/f1-f94b1fd3-1a08-4829-896e-792397ccdbfd/namespace/test-namespace/abc/cde

		var labels []*io_prometheus_client.LabelPair
		parts := strings.Split(request.RequestURI, "/")[4:]

		var labelName string
		for i, part := range parts {
			if i%2 == 0 {
				labelName = part
				continue
			}

			name := labelName
			value := part
			labels = append(labels, &io_prometheus_client.LabelPair{
				Name:  &name,
				Value: &value,
			})
		}

		return labels
	}

	if request != nil && request.Body != nil {
		defer errorh.SafeClose(request.Body)
		metricFamily := &io_prometheus_client.MetricFamily{}
		expfmt.NewDecoder(request.Body, expfmt.ResponseFormat(request.Header)).Decode(metricFamily)
		mf, ok := f.metricFamilies.Load(metricFamily.GetName())

		if metricFamily.GetMetric() != nil {
			groupedLabels := parseGroupedLabels()
			for _, m := range metricFamily.GetMetric() {
				m.Label = append(m.GetLabel(), groupedLabels...)
			}
		}

		if !ok {
			f.metricFamilies.Store(metricFamily.GetName, metricFamily)
		} else {
			value, ok := mf.(*io_prometheus_client.MetricFamily)
			if !ok {
				response.WriteHeader(http.StatusInternalServerError)
				return
			}
			value.Metric = append(value.Metric, metricFamily.GetMetric()...)
		}
	}
	f.setHasMetrics()
	response.WriteHeader(http.StatusAccepted)
}

func (f *FakePrometheus) StartServer() {
	f.metricFamilies = sync.Map{}
	f.server = &http.Server{Addr: fmt.Sprintf("localhost:%d", f.Port)}
	f.server.Handler = f
	go func() {
		if err := f.server.ListenAndServe(); err != nil {
			log.WithError(err).Error("error starting server")
		}
	}()
	log.Infof("Fake prometheus started on %s", f.server.Addr)
}

func (f *FakePrometheus) setHasMetrics() {
	atomic.StoreInt32(&f.hasMetrics, 1)
}

func (f *FakePrometheus) StopServer() {
	if err := f.server.Shutdown(context.Background()); err != nil {
		log.WithError(err).Error("error shutting down server")
	}
}

func (f *FakePrometheus) HasMetrics() bool {
	return atomic.LoadInt32(&f.hasMetrics) != 0
}

func (f *FakePrometheus) GetIterationDuration(scenario string, q float64) float64 {
	value, ok := f.metricFamilies.Load("form3_loadtest_iteration")
	if !ok {
		return 0.0
	}
	metrics, ok := value.(*io_prometheus_client.MetricFamily)
	if !ok {
		return 0.0
	}
	for _, metric := range metrics.GetMetric() {
		if metric.GetSummary() == nil {
			continue
		}
		test := ""
		for _, label := range metric.GetLabel() {
			if label.GetName() == "test" {
				test = label.GetValue()
			}
		}
		if test != scenario {
			continue
		}
		for _, quantile := range metric.GetSummary().GetQuantile() {
			if (*quantile).GetQuantile() == q {
				return quantile.GetValue()
			}
		}
	}
	return 0.0
}

func (f *FakePrometheus) GetMetricNames() []string {
	names := []string{}
	f.metricFamilies.Range(func(key, value interface{}) bool {
		name, ok := key.(string)
		if ok {
			names = append(names, name)
		}
		return true
	})
	return names
}

func (f *FakePrometheus) GetMetricFamily(name string) *io_prometheus_client.MetricFamily {
	value, ok := f.metricFamilies.Load(name)
	if !ok {
		return nil
	}
	metricFamily, ok := value.(*io_prometheus_client.MetricFamily)
	if !ok {
		return nil
	}
	return metricFamily
}

func (f *FakePrometheus) ClearMetrics() {
	keys := []string{}
	f.metricFamilies.Range(func(k, _ interface{}) bool {
		key, ok := k.(string)
		if ok {
			keys = append(keys, key)
		}
		return true
	})
	for _, key := range keys {
		f.metricFamilies.Delete(key)
	}
	atomic.StoreInt32(&f.hasMetrics, 0)
}
