package metrics

import (
	"errors"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type MetricType int

const (
	SetupResult MetricType = iota
	IterationResult
	TeardownResult
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

type Metrics struct {
	Setup            *prometheus.SummaryVec
	Iteration        *prometheus.SummaryVec
	Teardown         *prometheus.SummaryVec
	ProgressRegistry *prometheus.Registry
	Registry         *prometheus.Registry
	Progress         *prometheus.SummaryVec
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
		Progress: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Namespace:  metricNamespace,
			Subsystem:  metricSubsystem,
			Name:       "iteration",
			Help:       "Duration of iteration functions.",
			Objectives: percentileObjectives,
		}, []string{TestNameLabel, StageLabel, ResultLabel}),
		Teardown: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Namespace:  metricNamespace,
			Subsystem:  metricSubsystem,
			Name:       "teardown",
			Help:       "Duration of teardown functions.",
			Objectives: percentileObjectives,
		}, []string{TestNameLabel, ResultLabel}),
	}
}

func NewInstance(registry, progressRegistry *prometheus.Registry) *Metrics {
	i := buildMetrics()
	i.Registry = registry
	i.ProgressRegistry = progressRegistry

	i.Registry.MustRegister(
		i.Setup,
		i.Iteration,
		i.Teardown,
	)
	i.ProgressRegistry.MustRegister(i.Progress)

	return i
}

func Instance() *Metrics {
	once.Do(func() {
		defaultRegistry, ok := prometheus.DefaultRegisterer.(*prometheus.Registry)
		if !ok {
			panic(errors.New("casting prometheus.DefaultRegisterer to Registry"))
		}

		m = NewInstance(defaultRegistry, prometheus.NewRegistry())
	})
	return m
}

func (metrics *Metrics) Reset() {
	metrics.Iteration.Reset()
	metrics.Setup.Reset()
	metrics.Teardown.Reset()
}

func (metrics *Metrics) Record(metric MetricType, name string, stage string, result ResultType, nanoseconds int64) {
	switch metric {
	case SetupResult:
		metrics.Setup.WithLabelValues(name, result.String()).Observe(float64(nanoseconds))
	case IterationResult:
		metrics.Iteration.WithLabelValues(name, stage, result.String()).Observe(float64(nanoseconds))
		metrics.Progress.WithLabelValues(name, stage, result.String()).Observe(float64(nanoseconds))
	case TeardownResult:
		metrics.Teardown.WithLabelValues(name, result.String()).Observe(float64(nanoseconds))
	}
}
