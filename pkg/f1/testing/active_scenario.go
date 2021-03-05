package testing

import (
	"runtime/debug"
	"time"

	"github.com/form3tech-oss/f1/pkg/f1/metrics"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type ActiveScenario struct {
	Stages     []Stage
	TeardownFn TeardownFn
	Name       string
	id         string
	m          *metrics.Metrics
}

func NewActiveScenarios(name string, fn MultiStageSetupFn) (*ActiveScenario, bool) {
	s := &ActiveScenario{
		Name: name,
		id:   uuid.New().String(),
		m:    metrics.Instance(),
	}

	successful := s.Run(metrics.SetupResult, "setup", "0", "setup", func(t *T) {
		s.Stages, s.TeardownFn = fn(t)

		log.Infof("Added active scenario %s", s.id)
	})

	return s, successful
}

// Run performs a single iteration of the test. It returns `true` if the test was successful, `false` otherwise.
func (s *ActiveScenario) Run(metric metrics.MetricType, stage, vu, iter string, f func(t *T)) bool {
	t := NewT(vu, iter, s.Name)
	start := time.Now()
	done := make(chan struct{})
	go func() {
		defer s.checkResults(t, done)
		f(t)
	}()

	// wait for completion
	<-done
	s.m.Record(metric, s.Name, stage, metrics.Result(t.HasFailed()), time.Since(start).Nanoseconds())
	return !t.HasFailed()
}

func (s *ActiveScenario) checkResults(t *T, done chan<- struct{}) {
	r := recover()
	if r != nil {
		err, isError := r.(error)
		if isError {
			t.FailWithError(err)
			debug.PrintStack()
		} else {
			t.Errorf("panic in test iteration: %v", err)
		}
	}
	close(done)
}

func (s *ActiveScenario) RecordDroppedIteration() {
	s.m.Record(metrics.IterationResult, s.Name, "single", "dropped", 0)
}
