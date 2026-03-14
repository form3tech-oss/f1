package metrics

import (
	"sort"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	metricNamespace = "form3"
	metricSubsystem = "loadtest"
)

const (
	TestNameLabel = "test"
	StageLabel    = "stage"
	ResultLabel   = "result"
)

const IterationStage = "iteration"

type Metrics struct {
	setup                   *prometheus.SummaryVec
	iteration               *prometheus.SummaryVec
	registry                *prometheus.Registry
	iterationMetricsEnabled bool
	staticMetricLabelValues []string
}

func buildMetrics(staticMetrics map[string]string) *Metrics {
	percentileObjectives := map[float64]float64{
		0.5: 0.05, 0.75: 0.05, 0.9: 0.01, 0.95: 0.001, 0.99: 0.001, 0.9999: 0.00001, 1.0: 0.00001,
	}
	labelKeys := getStaticMetricLabelKeys(staticMetrics)
	return &Metrics{
		setup: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Namespace:  metricNamespace,
			Subsystem:  metricSubsystem,
			Name:       "setup",
			Help:       "Duration of setup functions.",
			Objectives: percentileObjectives,
		}, append([]string{TestNameLabel, ResultLabel}, labelKeys...)),
		iteration: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Namespace:  metricNamespace,
			Subsystem:  metricSubsystem,
			Name:       "iteration",
			Help:       "Duration of iteration functions.",
			Objectives: percentileObjectives,
		}, append([]string{TestNameLabel, StageLabel, ResultLabel}, labelKeys...)),
	}
}

func NewInstance(registry *prometheus.Registry,
	iterationMetricsEnabled bool,
	staticMetrics map[string]string,
) *Metrics {
	i := buildMetrics(staticMetrics)
	i.registry = registry

	i.registry.MustRegister(
		i.setup,
		i.iteration,
	)
	i.iterationMetricsEnabled = iterationMetricsEnabled
	i.staticMetricLabelValues = getStaticMetricLabelValues(staticMetrics)
	return i
}

// Gatherer returns the Prometheus registry for pushing metrics to a gateway.
func (m *Metrics) Gatherer() *prometheus.Registry {
	return m.registry
}

// IterationCollector returns the iteration metric for testing.
func (m *Metrics) IterationCollector() *prometheus.SummaryVec {
	return m.iteration
}

func (m *Metrics) Reset() {
	m.iteration.Reset()
	m.setup.Reset()
}

func (m *Metrics) RecordSetupResult(name string, result ResultType, nanoseconds int64) {
	labels := append([]string{name, result.String()}, m.staticMetricLabelValues...)
	m.setup.WithLabelValues(labels...).Observe(float64(nanoseconds))
}

func (m *Metrics) RecordIterationResult(name string, result ResultType, nanoseconds int64) {
	if !m.iterationMetricsEnabled {
		return
	}
	labels := append([]string{name, IterationStage, result.String()}, m.staticMetricLabelValues...)
	m.iteration.WithLabelValues(labels...).Observe(float64(nanoseconds))
}

func getStaticMetricLabelKeys(staticMetrics map[string]string) []string {
	return sortedKeys(staticMetrics)
}

func getStaticMetricLabelValues(staticMetrics map[string]string) []string {
	data := make([]string, 0, len(staticMetrics))
	for _, v := range sortedKeys(staticMetrics) {
		data = append(data, staticMetrics[v])
	}
	return data
}

func sortedKeys(staticMetrics map[string]string) []string {
	keys := make([]string, 0, len(staticMetrics))
	for k := range staticMetrics {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
