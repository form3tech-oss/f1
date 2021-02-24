package testing

import (
	"runtime/debug"
	"sync"
	"time"

	"github.com/form3tech-oss/f1/pkg/f1/metrics"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type ActiveScenario struct {
	Stages              []Stage
	TeardownFn          TeardownFn
	env                 map[string]string
	Name                string
	id                  string
	autoTeardownTimer   *CancellableTimer
	autoTeardownTimerMu sync.RWMutex
	AutoTeardownAfter   time.Duration
	m                   *metrics.Metrics
}

func NewActiveScenarios(name string, env map[string]string, fn MultiStageSetupFn, autoTeardownIdleDuration time.Duration) (*ActiveScenario, bool) {

	s := &ActiveScenario{
		Name:              name,
		id:                uuid.New().String(),
		env:               env,
		AutoTeardownAfter: autoTeardownIdleDuration,
		m:                 metrics.Instance(),
	}
	successful := s.Run(metrics.SetupResult, "setup", "0", "setup", func(t *T) {
		s.Stages, s.TeardownFn = fn(t)

		if autoTeardownIdleDuration > 0 {
			s.SetAutoTeardown(NewCancellableTimer(autoTeardownIdleDuration))
			go func() {
				ok := <-s.AutoTeardown().C
				s.SetAutoTeardown(nil)
				if ok {
					s.autoTeardown()
				}
			}()
		}

		activeScenarios.Store(s.id, s)
		log.Infof("Added active scenario %s",
			s.id)
	})
	return s, successful
}

func (s *ActiveScenario) AutoTeardown() *CancellableTimer {
	s.autoTeardownTimerMu.RLock()
	defer s.autoTeardownTimerMu.RUnlock()

	return s.autoTeardownTimer
}

func (s *ActiveScenario) SetAutoTeardown(timer *CancellableTimer) {
	s.autoTeardownTimerMu.Lock()
	defer s.autoTeardownTimerMu.Unlock()

	s.autoTeardownTimer = timer
}

// Run performs a single iteration of the test. It returns `true` if the test was successful, `false` otherwise.
func (s *ActiveScenario) Run(metric metrics.MetricType, stage, vu, iter string, f func(t *T)) bool {
	t := NewT(s.env, vu, iter, s.Name)
	start := time.Now()
	done := make(chan struct{})
	go func() {
		defer s.checkResults(t, done)
		f(t)
	}()
	if s.AutoTeardown() != nil {
		s.AutoTeardown().Reset(s.AutoTeardownAfter)
	}
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

func (s *ActiveScenario) autoTeardown() {
	log.Warn("Teardown not called - triggering timed teardown")
	if s.TeardownFn != nil {
		successful := s.Run(metrics.TeardownResult, "teardown", "0", "teardown", s.TeardownFn)
		if !successful {
			log.Error("auto teardown failed")
		}
	}
	activeScenarios.Delete(s.id)
}

func (s *ActiveScenario) RecordDroppedIteration() {
	s.m.Record(metrics.IterationResult, s.Name, "single", "dropped", 0)
}
