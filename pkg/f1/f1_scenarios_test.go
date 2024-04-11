package f1_test

import "testing"

func TestCombineScenarios(t *testing.T) {
	t.Parallel()

	given, when, then := newF1ScenarioStage(t)

	given.
		f1_is_configured_to_run_a_combined_scenario()

	when.
		the_f1_scenario_is_executed()

	then.
		each_scenarios_setup_and_iteration_functions_are_called()
}
