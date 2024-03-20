package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type MetricType int

const (
	SetupResult MetricType = iota
	IterationResult
	TeardownResult
)

type ResultType string

func (r ResultType) String() string {
	return string(r)
}

func ResultTypeFromString(result string) ResultType {
	switch result {
	case SucessResult.String():
		return SucessResult
	case FailedResult.String():
		return FailedResult
	case DroppedResult.String():
		return DroppedResult
	case UnknownResult.String():
		return UnknownResult
	default:
		return UnknownResult
	}
}

const (
	SucessResult  ResultType = "success"
	FailedResult  ResultType = "fail"
	DroppedResult ResultType = "dropped"
	UnknownResult ResultType = "unknown"
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
	Progress         *prometheus.SummaryVec
}

//nolint:gochecknoglobals // metrics are best suited as globals
var (
	m    *Metrics
	once sync.Once
)

func Instance() *Metrics {
	once.Do(func() {
		percentileObjectives := map[float64]float64{0.5: 0.05, 0.75: 0.05, 0.9: 0.01, 0.95: 0.001, 0.99: 0.001, 0.9999: 0.00001, 1.0: 0.00001}
		m = &Metrics{
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
		prometheus.MustRegister(
			m.Setup,
			m.Iteration,
			m.Teardown,
		)

		m.ProgressRegistry = prometheus.NewRegistry()
		m.ProgressRegistry.MustRegister(m.Progress)
	})
	return m
}

func Result(failed bool) ResultType {
	if failed {
		return FailedResult
	}
	return SucessResult
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
