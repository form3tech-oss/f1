package metrics

import (
	"errors"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	metricNamespace = "form3"
	metricSubsystem = "loadtest"
)

const IterationMetricName = "form3_loadtest_iteration"

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
}

//nolint:gochecknoglobals // removing the global Instance is a breaking change
var (
	m    *Metrics
	once sync.Once
)

func buildMetrics() *Metrics {
	percentileObjectives := map[float64]float64{
		0.5: 0.05, 0.75: 0.05, 0.9: 0.01, 0.95: 0.001, 0.99: 0.001, 0.9999: 0.00001, 1.0: 0.00001,
	}

	return &Metrics{
		Setup: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Namespace:  metricNamespace,
			Subsystem:  metricSubsystem,
			Name:       "setup",
			Help:       "Duration of setup functions.",
			Objectives: percentileObjectives,
		}, []string{TestNameLabel, ResultLabel}),
		Iteration: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Namespace:  metricNamespace,
			Subsystem:  metricSubsystem,
			Name:       "iteration",
			Help:       "Duration of iteration functions.",
			Objectives: percentileObjectives,
		}, []string{TestNameLabel, StageLabel, ResultLabel}),
	}
}

func NewInstance(registry *prometheus.Registry, iterationMetricsEnabled bool) *Metrics {
	i := buildMetrics()
	i.Registry = registry

	i.Registry.MustRegister(
		i.Setup,
		i.Iteration,
	)
	i.IterationMetricsEnabled = iterationMetricsEnabled

	return i
}

func Init(iterationMetricsEnabled bool) {
	once.Do(func() {
		defaultRegistry, ok := prometheus.DefaultRegisterer.(*prometheus.Registry)
		if !ok {
			panic(errors.New("casting prometheus.DefaultRegisterer to Registry"))
		}
		m = NewInstance(defaultRegistry, iterationMetricsEnabled)
	})
	m.IterationMetricsEnabled = iterationMetricsEnabled
}

func Instance() *Metrics {
	return m
}

func (metrics *Metrics) Reset() {
	metrics.Iteration.Reset()
	metrics.Setup.Reset()
}

func (metrics *Metrics) RecordSetupResult(name string, result ResultType, nanoseconds int64) {
	metrics.Setup.WithLabelValues(name, result.String()).Observe(float64(nanoseconds))
}

func (metrics *Metrics) RecordIterationResult(name string, result ResultType, nanoseconds int64) {
	if !metrics.IterationMetricsEnabled {
		return
	}

	metrics.Iteration.WithLabelValues(name, IterationStage, result.String()).Observe(float64(nanoseconds))
}

func (metrics *Metrics) RecordIterationStage(name string, stage string, result ResultType, nanoseconds int64) {
	if !metrics.IterationMetricsEnabled {
		return
	}

	metrics.Iteration.WithLabelValues(name, stage, result.String()).Observe(float64(nanoseconds))
}
