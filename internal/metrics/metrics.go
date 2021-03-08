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

type Metrics struct {
	Setup            *prometheus.SummaryVec
	Iteration        *prometheus.SummaryVec
	Teardown         *prometheus.SummaryVec
	ProgressRegistry *prometheus.Registry
	Progress         *prometheus.SummaryVec
}

var (
	m    *Metrics
	once sync.Once
)

func Instance() *Metrics {
	once.Do(func() {
		percentileObjectives := map[float64]float64{0.5: 0.05, 0.75: 0.05, 0.9: 0.01, 0.95: 0.001, 0.99: 0.001, 0.9999: 0.00001, 1.0: 0.00001}
		m = &Metrics{
			Setup: prometheus.NewSummaryVec(prometheus.SummaryOpts{
				Namespace:  "form3",
				Subsystem:  "loadtest",
				Name:       "setup",
				Help:       "Duration of setup functions.",
				Objectives: percentileObjectives,
			}, []string{"test", "result"}),
			Iteration: prometheus.NewSummaryVec(prometheus.SummaryOpts{
				Namespace:  "form3",
				Subsystem:  "loadtest",
				Name:       "iteration",
				Help:       "Duration of iteration functions.",
				Objectives: percentileObjectives,
			}, []string{"test", "stage", "result"}),
			Progress: prometheus.NewSummaryVec(prometheus.SummaryOpts{
				Namespace:  "form3",
				Subsystem:  "loadtest",
				Name:       "iteration",
				Help:       "Duration of iteration functions.",
				Objectives: percentileObjectives,
			}, []string{"test", "stage", "result"}),
			Teardown: prometheus.NewSummaryVec(prometheus.SummaryOpts{
				Namespace:  "form3",
				Subsystem:  "loadtest",
				Name:       "teardown",
				Help:       "Duration of teardown functions.",
				Objectives: percentileObjectives,
			}, []string{"test", "result"}),
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

func Result(failed bool) string {
	if failed {
		return "fail"
	}
	return "success"
}

func (metrics *Metrics) Reset() {
	metrics.Iteration.Reset()
	metrics.Setup.Reset()
	metrics.Teardown.Reset()
}

func (metrics *Metrics) Record(metric MetricType, name string, stage string, result string, nanoseconds int64) {
	switch metric {
	case SetupResult:
		metrics.Setup.WithLabelValues(name, result).Observe(float64(nanoseconds))
	case IterationResult:
		metrics.Iteration.WithLabelValues(name, stage, result).Observe(float64(nanoseconds))
		metrics.Progress.WithLabelValues(name, stage, result).Observe(float64(nanoseconds))
	case TeardownResult:
		metrics.Teardown.WithLabelValues(name, result).Observe(float64(nanoseconds))
	}
}
