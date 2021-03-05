package testing

import (
	"runtime/debug"
	"time"

	"github.com/form3tech-oss/f1/pkg/f1/metrics"
	"github.com/google/uuid"
)

type ActiveScenario struct {
	Stages     []Stage
	TeardownFn TeardownFn
	Name       string
	id         string
	m          *metrics.Metrics
	t          *T
}

func NewActiveScenario(name string, fn MultiStageSetupFn) (*ActiveScenario, bool) {
	t := NewT("0", "setup", name)

	s := &ActiveScenario{
		Name: name,
		id:   uuid.New().String(),
		m:    metrics.Instance(),
		t:    t,
	}

	start := time.Now()
	done := make(chan struct{})
	go func() {
		defer checkResults(t, done)
		s.Stages, s.TeardownFn = fn(t)
	}()

	// wait for completion
	<-done
	s.m.Record(metrics.SetupResult, s.Name, "setup", metrics.Result(t.HasFailed()), time.Since(start).Nanoseconds())
	return s, !t.HasFailed()
}

// Run performs a single iteration of the test. It returns `true` if the test was successful, `false` otherwise.
func (s *ActiveScenario) Run(metric metrics.MetricType, stage, vu, iter string, f func(t *T)) bool {
	t := NewT(vu, iter, s.Name)
	defer t.teardown()

	start := time.Now()
	done := make(chan struct{})
	go func() {
		defer checkResults(t, done)
		f(t)
	}()

	// wait for completion
	<-done
	s.m.Record(metric, s.Name, stage, metrics.Result(t.HasFailed()), time.Since(start).Nanoseconds())
	return !t.HasFailed()
}

func checkResults(t *T, done chan<- struct{}) {
	r := recover()
	if r != nil {
		err, isError := r.(error)
		if isError {
			t.FailWithError(err)
			debug.PrintStack()
		} else {
			t.Errorf("panic in %s: %v", t.Iteration, err)
		}
	}
	close(done)
}

func (s *ActiveScenario) RecordDroppedIteration() {
	s.m.Record(metrics.IterationResult, s.Name, "single", "dropped", 0)
}

func (s *ActiveScenario) Teardown() {
	s.t.teardown()
}
