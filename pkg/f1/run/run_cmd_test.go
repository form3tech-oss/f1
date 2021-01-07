package run

import (
	"testing"
	"time"
)

// TestSimpleFlow is equivalent to a single run from TestParameterised. It's useful for debugging individual test runs
func TestSimpleFlow(t *testing.T) {
	//t.Skip("Duplicate of Parameterised test. Useful for manual testing when adding new tests or debugging, so leaving in place")
	given, when, then := NewRunTestStage(t)

	test := TestParam{
		name:                   "basic test",
		constantRate:           "10/100ms",
		testDuration:           100 * time.Millisecond,
		concurrency:            100,
		iterationDuration:      100 * time.Millisecond,
		expectedRunTime:        100 * time.Millisecond,
		expectedCompletedTests: 10,
	}
	given.
		a_trigger_type_of(test.triggerType).and().
		a_rate_of(test.constantRate).and().
		a_stage_of(test.stages).and().
		an_iteration_frequency_of(test.iterationFrequency).and().
		a_distribution_type(test.distributionType).and().
		a_duration_of(test.testDuration).and().
		a_concurrency_of(test.concurrency).and().
		an_iteration_limit_of(test.maxIterations).and().
		a_scenario_where_each_iteration_takes(test.iterationDuration).and().
		a_config_file_location_of(test.configFile).and().
		a_start_rate_of(test.startRate).and().
		a_end_rate_of(test.endRate).and().
		a_ramp_duration_of(test.rampDuration)

	when.i_start_a_timer().and().
		i_execute_the_run_command()

	then.
		the_command_finished_with_failure_of(test.expectedFailure).and().
		the_command_should_have_run_for_approx(test.expectedRunTime).and().
		the_number_of_started_iterations_should_be(test.expectedCompletedTests).and().
		the_number_of_dropped_iterations_should_be(test.expectedDroppedIterations)

}

type TriggerType int

const (
	Constant TriggerType = iota
	Staged
	Users
	Ramp
	File
)

type TestParam struct {
	name                      string
	triggerType               TriggerType
	constantRate              string
	testDuration              time.Duration
	expectedRunTime           time.Duration
	expectedCompletedTests    int32
	concurrency               int
	iterationDuration         time.Duration
	expectedDroppedIterations uint64
	expectedFailure           bool
	maxIterations             int32
	stages                    string
	iterationFrequency        string
	distributionType          string
	configFile                string
	startRate                 string
	endRate                   string
	rampDuration              string
}

func TestParameterised(t *testing.T) {
	for _, test := range []TestParam{
		{
			name:                   "basic test",
			constantRate:           "10/100ms",
			testDuration:           100 * time.Millisecond,
			concurrency:            100,
			iterationDuration:      100 * time.Millisecond,
			expectedRunTime:        100 * time.Millisecond,
			expectedCompletedTests: 10,
		},
		{
			name:                   "finishes at ends of duration",
			constantRate:           "10/2s",
			testDuration:           2 * time.Second,
			concurrency:            100,
			iterationDuration:      200 * time.Millisecond,
			expectedRunTime:        2 * time.Second,
			expectedCompletedTests: 10,
		},
		{
			name:                   "times out",
			constantRate:           "1/s",
			testDuration:           1 * time.Second,
			concurrency:            100,
			iterationDuration:      2 * time.Second,
			expectedRunTime:        2 * time.Second,
			expectedCompletedTests: 1,
			distributionType:       "none",
		},
		{
			name:                   "next iteration can start if previous still running",
			constantRate:           "10/1s",
			testDuration:           2 * time.Second,
			concurrency:            200,
			iterationDuration:      2 * time.Second,
			expectedRunTime:        3 * time.Second,
			expectedCompletedTests: 20,
			distributionType:       "none",
		},
		{
			name:                      "next iteration won't start if previous still running and limited by concurrency",
			constantRate:              "10/1s",
			testDuration:              2 * time.Second,
			expectedRunTime:           2 * time.Second,
			expectedCompletedTests:    10,
			concurrency:               10,
			iterationDuration:         2 * time.Second,
			expectedDroppedIterations: 10,
			expectedFailure:           true,
			distributionType:          "none",
		},
		{
			name:                   "limited iterations",
			constantRate:           "10/100ms",
			testDuration:           1 * time.Second,
			concurrency:            100,
			iterationDuration:      100 * time.Millisecond,
			maxIterations:          7,
			expectedRunTime:        100 * time.Millisecond,
			expectedCompletedTests: 7,
		},
		{
			name:                   "limited iterations running for multiple loops",
			constantRate:           "10/50ms",
			testDuration:           5 * time.Second,
			concurrency:            100,
			iterationDuration:      10 * time.Millisecond,
			maxIterations:          17,
			expectedCompletedTests: 17,
		},
		{
			name:                   "regular distribution of a constant rate",
			constantRate:           "10/s",
			testDuration:           2 * time.Second,
			concurrency:            100,
			iterationDuration:      1 * time.Millisecond,
			expectedRunTime:        2 * time.Second,
			expectedCompletedTests: 20,
			distributionType:       "regular",
		},
		{
			name:                   "random distribution of a constant rate",
			constantRate:           "10/s",
			testDuration:           2 * time.Second,
			concurrency:            100,
			iterationDuration:      1 * time.Millisecond,
			expectedRunTime:        2 * time.Second,
			expectedCompletedTests: 20,
			distributionType:       "random",
		},
		{
			name:                   "run only half of requests on half of the time using regular distribution",
			constantRate:           "10/2s",
			testDuration:           1 * time.Second,
			concurrency:            100,
			iterationDuration:      1 * time.Millisecond,
			expectedRunTime:        1 * time.Second,
			expectedCompletedTests: 5,
			distributionType:       "regular",
		},
		{
			name:                   "simple staged test",
			triggerType:            Staged,
			stages:                 "0ms:0, 50ms: 100, 100ms: 100, 50ms:0",
			iterationFrequency:     "100ms",
			testDuration:           200 * time.Millisecond,
			concurrency:            100,
			iterationDuration:      1 * time.Millisecond,
			expectedCompletedTests: 100,
		},
		{
			name:                   "staged test",
			triggerType:            Staged,
			stages:                 "0ms:0, 100ms: 10, 200ms:0",
			iterationFrequency:     "50ms",
			testDuration:           200 * time.Millisecond,
			concurrency:            100,
			iterationDuration:      1 * time.Millisecond,
			expectedCompletedTests: 23,
		},
		{
			name:                   "regular distribution of a staged trigger",
			triggerType:            Staged,
			stages:                 "0ms:0, 50ms: 100, 100ms: 100, 50ms:0",
			iterationFrequency:     "100ms",
			testDuration:           200 * time.Millisecond,
			concurrency:            100,
			iterationDuration:      1 * time.Millisecond,
			expectedCompletedTests: 100,
			distributionType:       "regular",
		},
		{
			name:                   "random distribution of a staged trigger",
			triggerType:            Staged,
			stages:                 "0ms:0, 50ms: 100, 100ms: 100, 50ms:0",
			iterationFrequency:     "100ms",
			testDuration:           200 * time.Millisecond,
			concurrency:            100,
			iterationDuration:      1 * time.Millisecond,
			expectedCompletedTests: 100,
			distributionType:       "random",
		},
		{
			name:                   "run half of requests on half of the time on staged test using regular distribution",
			triggerType:            Staged,
			stages:                 "0ms:0, 4s: 100",
			iterationFrequency:     "1s",
			testDuration:           3500 * time.Millisecond,
			concurrency:            100,
			iterationDuration:      1 * time.Millisecond,
			expectedCompletedTests: 112,
			distributionType:       "regular",
		},
		{
			name:                   "users test slow iterations",
			triggerType:            Users,
			testDuration:           2 * time.Second,
			expectedRunTime:        2 * time.Second,
			expectedCompletedTests: 10,
			concurrency:            10,
			iterationDuration:      2 * time.Second,
		},
		{
			name:                   "users test normal iterations",
			triggerType:            Users,
			testDuration:           2 * time.Second,
			expectedRunTime:        2 * time.Second,
			expectedCompletedTests: 20,
			concurrency:            10,
			iterationDuration:      1 * time.Second,
		},
		{
			name:                   "users test fast iterations",
			triggerType:            Users,
			testDuration:           2 * time.Second,
			expectedRunTime:        2 * time.Second,
			expectedCompletedTests: 40,
			concurrency:            10,
			iterationDuration:      500 * time.Millisecond,
		},
		{
			name:                   "config file test",
			triggerType:            File,
			configFile:             "testdata/config-file.yaml",
			testDuration:           5 * time.Second,
			concurrency:            50,
			maxIterations:          1000,
			iterationDuration:      100 * time.Millisecond,
			expectedRunTime:        2200 * time.Millisecond,
			expectedCompletedTests: 110,
		},
		{
			name:                   "simple ramp up test",
			triggerType:            Ramp,
			startRate:              "0/100ms",
			endRate:                "10/100ms",
			rampDuration:           "1s",
			testDuration:           1 * time.Second,
			concurrency:            50,
			maxIterations:          1000,
			iterationDuration:      1 * time.Millisecond,
			expectedRunTime:        1 * time.Second,
			expectedCompletedTests: 45,
			distributionType:       "none",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			given, when, then := NewRunTestStage(t)

			given.
				a_trigger_type_of(test.triggerType).and().
				a_rate_of(test.constantRate).and().
				a_stage_of(test.stages).and().
				an_iteration_frequency_of(test.iterationFrequency).and().
				a_distribution_type(test.distributionType).and().
				a_duration_of(test.testDuration).and().
				a_concurrency_of(test.concurrency).and().
				an_iteration_limit_of(test.maxIterations).and().
				a_scenario_where_each_iteration_takes(test.iterationDuration).and().
				a_config_file_location_of(test.configFile).and().
				a_start_rate_of(test.startRate).and().
				a_end_rate_of(test.endRate).and().
				a_ramp_duration_of(test.rampDuration)

			when.i_start_a_timer().and().
				i_execute_the_run_command()

			then.
				the_command_finished_with_failure_of(test.expectedFailure).and().
				the_command_should_have_run_for_approx(test.expectedRunTime).and().
				the_number_of_started_iterations_should_be(test.expectedCompletedTests).and().
				the_number_of_dropped_iterations_should_be(test.expectedDroppedIterations).and().
				teardown_is_called_once()
		})
	}
}

func TestNoneDistribution(t *testing.T) {
	given, when, then := NewRunTestStage(t)

	given.
		a_trigger_type_of(Constant).and().
		a_rate_of("10/s").and().
		a_distribution_type("none").and().
		a_duration_of(500 * time.Millisecond).and().
		a_concurrency_of(50).and().
		an_iteration_limit_of(1000).and().
		a_scenario_where_each_iteration_takes(1 * time.Millisecond)

	when.i_start_a_timer().and().
		i_execute_the_run_command()

	then.there_should_be_x_requests_sent_over_y_intervals_of_z_ms(10, 1, 1000)
}

func TestRegularDistribution(t *testing.T) {
	given, when, then := NewRunTestStage(t)

	given.
		a_trigger_type_of(Constant).and().
		a_rate_of("10/s").and().
		a_distribution_type("regular").and().
		a_duration_of(500 * time.Millisecond).and().
		a_concurrency_of(50).and().
		an_iteration_limit_of(1000).and().
		a_scenario_where_each_iteration_takes(1 * time.Millisecond)

	when.i_start_a_timer().and().
		i_execute_the_run_command()

	then.there_should_be_x_requests_sent_over_y_intervals_of_z_ms(1, 5, 100)
}

func TestRandomDistribution(t *testing.T) {
	given, when, then := NewRunTestStage(t)

	given.
		a_trigger_type_of(Constant).and().
		a_rate_of("10/s").and().
		a_distribution_type("random").and().
		a_duration_of(500 * time.Millisecond).and().
		a_concurrency_of(50).and().
		an_iteration_limit_of(1000).and().
		a_scenario_where_each_iteration_takes(1 * time.Millisecond)

	when.i_start_a_timer().and().
		i_execute_the_run_command()

	then.the_requests_are_not_sent_all_at_once()
}

func TestRunScenarioThatFailsSetup(t *testing.T) {
	given, when, then := NewRunTestStage(t)

	given.
		a_test_scenario_that_always_fails_setup().and().
		a_rate_of("1/s").and().
		a_duration_of(1 * time.Second)

	when.i_execute_the_run_command()

	then.the_command_should_fail().and().
		metrics_are_pushed_to_prometheus()
}

func TestRunScenarioThatFails(t *testing.T) {
	given, when, then := NewRunTestStage(t)

	given.
		a_test_scenario_that_always_fails().and().
		a_rate_of("1").and().
		a_duration_of(1 * time.Second)

	when.i_execute_the_run_command()

	then.the_command_should_fail().and().
		teardown_is_called().and().
		metrics_are_pushed_to_prometheus()
}

func TestRunScenarioThatPanics(t *testing.T) {
	given, when, then := NewRunTestStage(t)

	given.
		a_test_scenario_that_always_panics().and().
		a_rate_of("1").and().
		a_duration_of(1 * time.Second)

	when.i_execute_the_run_command()

	then.the_command_should_fail().and().
		teardown_is_called().and().
		metrics_are_pushed_to_prometheus()
}

func TestRunScenarioThatFailsAnAssertion(t *testing.T) {
	given, when, then := NewRunTestStage(t)

	given.
		a_test_scenario_that_always_fails_an_assertion().and().
		a_rate_of("1").and().
		a_duration_of(1 * time.Second)

	when.i_execute_the_run_command()

	then.the_command_should_fail().and().
		teardown_is_called().and().
		metrics_are_pushed_to_prometheus()
}

func TestRunScenarioThatFailsOccasionally(t *testing.T) {
	given, when, then := NewRunTestStage(t)
	given.
		a_test_scenario_that_fails_intermittently().and().
		a_rate_of("100/1s").and().
		// Run less than 1 second, since if we run exactly for 1 second the test might run into another iteration.
		// This would then lead to 200 requests being made, making the test fail
		a_duration_of(500 * time.Millisecond).and().
		a_distribution_type("none")

	when.i_execute_the_run_command()

	then.the_results_should_show_n_failures(50).and().
		the_results_should_show_n_successful_iterations(50).and().
		teardown_is_called()
}

func TestInterruptedRun(t *testing.T) {

	given, when, then := NewRunTestStage(t)
	given.
		a_rate_of("5/10ms").and().
		a_duration_of(5 * time.Second).and().
		a_scenario_where_each_iteration_takes(0 * time.Second).and().
		a_distribution_type("none")

	when.the_test_run_is_started().and().
		the_test_run_is_interrupted()

	then.
		teardown_is_called_within_50ms().and().
		metrics_are_pushed_to_prometheus()

}

func TestFinalRunMetrics(t *testing.T) {

	given, when, then := NewRunTestStage(t)
	given.
		a_rate_of("100/100ms").and().
		a_duration_of(450 * time.Millisecond).and().
		a_scenario_where_the_final_iteration_takes_100ms()

	when.i_execute_the_run_command()

	then.
		metrics_are_pushed_to_prometheus().and().
		the_100th_percentile_is_slow().and().
		all_other_percentiles_are_fast()

}

func TestSetupMetricsAreRecorded(t *testing.T) {
	given, when, then := NewRunTestStage(t)

	given.
		a_rate_of("1/s").and().
		a_scenario_where_each_iteration_takes(1 * time.Millisecond)

	when.i_execute_the_run_command()

	then.
		metrics_are_pushed_to_prometheus().and().
		there_is_a_metric_called("form3_loadtest_setup")
}

func TestFailureCounts(t *testing.T) {
	given, when, then := NewRunTestStage(t)

	given.
		a_rate_of("10/s").and().
		a_duration_of(500 * time.Millisecond).and().
		a_test_scenario_that_fails_intermittently().and().
		a_distribution_type("none")

	when.i_execute_the_run_command()

	then.
		metrics_are_pushed_to_prometheus().and().
		there_is_a_metric_called("form3_loadtest_iteration").and().
		the_iteration_metric_has_n_results(5, "success").and().
		the_iteration_metric_has_n_results(5, "fail")
}
