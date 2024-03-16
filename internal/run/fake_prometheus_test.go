package run_test

import (
	"net/http"
	"strings"
	"sync"
	"testing"

	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

const defaultIterationFamilyName = "form3_loadtest_iteration"

func parseGroupLabels(requiestURI string) []*io_prometheus_client.LabelPair {
	// labels added via push.Grouping are passed through URI
	// example: /metrics/job/f1-f94b1fd3-1a08-4829-896e-792397ccdbfd/namespace/test-namespace/abc/cde

	parts := strings.Split(requiestURI, "/")[4:]
	labels := make([]*io_prometheus_client.LabelPair, 0, len(parts)/2)

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

func FakePrometheusHandler(t *testing.T, metricData *MetricData) http.HandlerFunc {
	t.Helper()

	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		if request == nil || request.Body == nil {
			t.Errorf("http empty request received")
			responseWriter.WriteHeader(http.StatusInternalServerError)
			return
		}

		defer request.Body.Close()

		metricFamily := &io_prometheus_client.MetricFamily{}
		err := expfmt.NewDecoder(request.Body, expfmt.ResponseFormat(request.Header)).Decode(metricFamily)
		if err != nil {
			t.Errorf("error decoding request body: %s", err)
			responseWriter.WriteHeader(http.StatusInternalServerError)
			return
		}

		if metricFamily.GetMetric() != nil {
			groupedLabels := parseGroupLabels(request.RequestURI)
			for _, m := range metricFamily.GetMetric() {
				m.Label = append(m.GetLabel(), groupedLabels...)
			}
		}

		mf := metricData.GetMetricFamily(metricFamily.GetName())
		if mf == nil {
			metricData.SetMetricFamily(metricFamily.GetName(), metricFamily)
		} else {
			mf.Metric = append(mf.Metric, metricFamily.GetMetric()...)
		}

		responseWriter.WriteHeader(http.StatusAccepted)
	})
}

type MetricData struct {
	data   map[string]*io_prometheus_client.MetricFamily
	dataMu sync.RWMutex
}

func NewMetricData() *MetricData {
	return &MetricData{
		data: make(map[string]*io_prometheus_client.MetricFamily),
	}
}

func (m *MetricData) GetMetricFamily(name string) *io_prometheus_client.MetricFamily {
	m.dataMu.RLock()
	defer m.dataMu.RUnlock()

	return m.data[name]
}

func (m *MetricData) SetMetricFamily(name string, data *io_prometheus_client.MetricFamily) {
	m.dataMu.Lock()
	defer m.dataMu.Unlock()

	m.data[name] = data
}

func (m *MetricData) Empty() bool {
	m.dataMu.RLock()
	defer m.dataMu.RUnlock()

	return len(m.data) == 0
}

func (m *MetricData) GetMetricNames() []string {
	m.dataMu.RLock()
	defer m.dataMu.RUnlock()

	names := make([]string, 0, len(m.data))

	for name := range m.data {
		names = append(names, name)
	}

	return names
}

func (m *MetricData) GetIterationDuration(scenario string, q float64) float64 {
	m.dataMu.RLock()
	defer m.dataMu.RUnlock()

	metrics := m.GetMetricFamily(defaultIterationFamilyName)
	if metrics == nil {
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
