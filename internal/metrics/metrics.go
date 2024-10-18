package metrics

import (
	"errors"
	"sort"
	"sync"

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
	Setup                   *prometheus.SummaryVec
	Iteration               *prometheus.SummaryVec
	Registry                *prometheus.Registry
	IterationMetricsEnabled bool
	staticMetricLabelValues []string
}

//nolint:gochecknoglobals // removing the global Instance is a breaking change
var (
	m    *Metrics
	once sync.Once
)

func buildMetrics(staticMetrics map[string]string) *Metrics {
	percentileObjectives := map[float64]float64{
		0.5: 0.05, 0.75: 0.05, 0.9: 0.01, 0.95: 0.001, 0.99: 0.001, 0.9999: 0.00001, 1.0: 0.00001,
	}
	labelKeys := getStaticMetricLabelKeys(staticMetrics)
	return &Metrics{
		Setup: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Namespace:  metricNamespace,
			Subsystem:  metricSubsystem,
			Name:       "setup",
			Help:       "Duration of setup functions.",
			Objectives: percentileObjectives,
		}, append([]string{TestNameLabel, ResultLabel}, labelKeys...)),
		Iteration: prometheus.NewSummaryVec(prometheus.SummaryOpts{
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
	i.Registry = registry

	i.Registry.MustRegister(
		i.Setup,
		i.Iteration,
	)
	i.IterationMetricsEnabled = iterationMetricsEnabled
	i.staticMetricLabelValues = getStaticMetricLabelValues(staticMetrics)
	return i
}

func Init(iterationMetricsEnabled bool) {
	InitWithStaticMetrics(iterationMetricsEnabled, nil)
}

func InitWithStaticMetrics(iterationMetricsEnabled bool, staticMetrics map[string]string) {
	once.Do(func() {
		defaultRegistry, ok := prometheus.DefaultRegisterer.(*prometheus.Registry)
		if !ok {
			panic(errors.New("casting prometheus.DefaultRegisterer to Registry"))
		}
		m = NewInstance(defaultRegistry, iterationMetricsEnabled, staticMetrics)
	})
}

func Instance() *Metrics {
	return m
}

func (metrics *Metrics) Reset() {
	metrics.Iteration.Reset()
	metrics.Setup.Reset()
}

func (metrics *Metrics) RecordSetupResult(name string, result ResultType, nanoseconds int64) {
	labels := append([]string{name, result.String()}, metrics.staticMetricLabelValues...)
	metrics.Setup.WithLabelValues(labels...).Observe(float64(nanoseconds))
}

func (metrics *Metrics) RecordIterationResult(name string, result ResultType, nanoseconds int64) {
	if !metrics.IterationMetricsEnabled {
		return
	}
	labels := append([]string{name, IterationStage, result.String()}, metrics.staticMetricLabelValues...)
	metrics.Iteration.WithLabelValues(labels...).Observe(float64(nanoseconds))
}

func (metrics *Metrics) RecordIterationStage(name string, stage string, result ResultType, nanoseconds int64) {
	if !metrics.IterationMetricsEnabled {
		return
	}
	labels := append([]string{name, stage, result.String()}, metrics.staticMetricLabelValues...)
	metrics.Iteration.WithLabelValues(labels...).Observe(float64(nanoseconds))
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
