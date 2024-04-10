package f1_test

import (
	"os"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/form3tech-oss/f1/v2/pkg/f1"
	f1_testing "github.com/form3tech-oss/f1/v2/pkg/f1/testing"
)

type f1Stage struct {
	t       *testing.T
	assert  *assert.Assertions
	require *require.Assertions

	f1       *f1.F1
	errCh    chan error
	scenario string
	runCount atomic.Uint32
}

func newF1Stage(t *testing.T) (*f1Stage, *f1Stage, *f1Stage) {
	t.Helper()

	s := &f1Stage{
		t:       t,
		require: require.New(t),
		assert:  assert.New(t),
		f1:      f1.New(),
		errCh:   make(chan error),
	}

	return s, s, s
}

func (s *f1Stage) and() *f1Stage {
	return s
}

func (s *f1Stage) after_duration_signal_will_be_sent(duration time.Duration, signal syscall.Signal) *f1Stage {
	go func() {
		time.Sleep(duration)

		process, err := os.FindProcess(os.Getpid())
		if err != nil {
			s.errCh <- err
			return
		}

		s.errCh <- process.Signal(signal)
	}()

	return s
}

func (s *f1Stage) a_scenario_where_each_iteration_takes(duration time.Duration) *f1Stage {
	s.scenario = "scenario_where_each_iteration_takes_" + duration.String()
	s.f1.Add(s.scenario, func(*f1_testing.T) f1_testing.RunFn {
		return func(*f1_testing.T) {
			s.runCount.Add(1)
			time.Sleep(duration)
		}
	})

	return s
}

func (s *f1Stage) the_f1_scenario_is_executed_with_constant_rate_and_args(args ...string) *f1Stage {
	err := s.f1.ExecuteWithArgs(append([]string{
		"run", "constant", s.scenario,
	}, args...))
	s.require.NoError(err, "error executing scenarios")

	return s
}

func (s *f1Stage) expect_the_scenario_iterations_to_have_run_no_more_than(count uint32) *f1Stage {
	s.assert.Less(s.runCount.Load(), count)

	return s
}

func (s *f1Stage) expect_no_error_sending_signals() *f1Stage {
	err := <-s.errCh
	s.require.NoError(err)

	return s
}

func (s *f1Stage) expect_no_goroutines_to_run() *f1Stage {
	s.require.NoError(goleak.Find())

	return s
}
