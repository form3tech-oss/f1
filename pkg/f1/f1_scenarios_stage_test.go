package f1_test

import (
	"sync/atomic"
	"testing"

	"github.com/form3tech-oss/f1/v2/pkg/f1"
	f1_testing "github.com/form3tech-oss/f1/v2/pkg/f1/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type f1ScenariosStage struct {
	t         *testing.T
	scenarios []*scenario
	runner    *f1.F1
}

type scenario struct {
	setups     *int32
	iterations *int32
}

func (s scenario) scenariofn(_ *f1_testing.T) f1_testing.RunFn {
	atomic.AddInt32(s.setups, 1)

	return func(_ *f1_testing.T) {
		atomic.AddInt32(s.iterations, 1)
	}
}

func newScenario(setups, iterations int32) scenario {
	return scenario{
		setups:     &setups,
		iterations: &iterations,
	}
}

func newF1ScenarioStage(t *testing.T) (given, when, then *f1ScenariosStage) {
	s := &f1ScenariosStage{
		t: t,
	}

	for i := 0; i < 10; i++ {
		n := newScenario(0, 0)
		s.scenarios = append(s.scenarios, &n)
	}

	return s, s, s
}

func (s *f1ScenariosStage) f1_is_configured_to_run_a_combined_scenario() {
	var scenarios []f1_testing.ScenarioFn
	for _, scn := range s.scenarios {
		fn := scn.scenariofn
		scenarios = append(scenarios, fn)
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
		assert.Equal(s.t, *scn.setups, int32(1))
		assert.GreaterOrEqual(s.t, *scn.iterations, int32(5))
	}
}
