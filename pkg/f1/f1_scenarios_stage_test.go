package f1_test

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/form3tech-oss/f1/v2/pkg/f1"
	f1_testing "github.com/form3tech-oss/f1/v2/pkg/f1/testing"
)

type f1ScenariosStage struct {
	t         *testing.T
	runner    *f1.F1
	scenarios []*scenario
}

type scenario struct {
	setups     atomic.Uint32
	iterations atomic.Uint32
}

func (s *scenario) scenariofn(*f1_testing.T) f1_testing.RunFn {
	s.setups.Add(1)

	return func(*f1_testing.T) {
		s.iterations.Add(1)
	}
}

func newScenario(setups, iterations uint32) *scenario {
	s := &scenario{}
	s.setups.Store(setups)
	s.iterations.Store(iterations)

	return s
}

func newF1ScenarioStage(t *testing.T) (*f1ScenariosStage, *f1ScenariosStage, *f1ScenariosStage) {
	t.Helper()

	s := &f1ScenariosStage{
		t: t,
	}

	for range 10 {
		n := newScenario(0, 0)
		s.scenarios = append(s.scenarios, n)
	}

	return s, s, s
}

func (s *f1ScenariosStage) f1_is_configured_to_run_a_combined_scenario() {
	scenarios := make([]f1_testing.ScenarioFn, len(s.scenarios))
	for i, scn := range s.scenarios {
		fn := scn.scenariofn
		scenarios[i] = fn
	}

	s.runner = f1.New().Add("combined", f1.CombineScenarios(scenarios...))
}

func (s *f1ScenariosStage) the_f1_scenario_is_executed() {
	err := s.runner.ExecuteWithArgs([]string{
		"run", "constant", "combined",
		"--rate", "5/1s",
		"--max-duration", "1s",
	})
	require.NoError(s.t, err, "error executing scenarios")
}

func (s *f1ScenariosStage) each_scenarios_setup_and_iteration_functions_are_called() {
	for _, scn := range s.scenarios {
		assert.Equal(s.t, uint32(1), scn.setups.Load())
		assert.GreaterOrEqual(s.t, scn.iterations.Load(), uint32(5))
	}
}
