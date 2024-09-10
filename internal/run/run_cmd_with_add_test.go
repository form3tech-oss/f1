package run_test

import (
	"testing"
	"time"
)


func TestWithAddSimpleFlow(t *testing.T) {
	t.Parallel()

	given, when, then := NewRunWithAddTestStage(t)

	test := testParam{
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

	when.a_timer_is_started().and().
		the_run_command_is_executed()

	then.
		the_command_finished_with_failure_of(test.expectedFailure).and().
		the_command_should_have_run_for_approx(test.expectedRunTime).and().
		the_number_of_started_iterations_should_be(test.expectedCompletedTests).and().
		the_number_of_dropped_iterations_should_be(test.expectedDroppedIterations)
}

func TestWithAddParameterised(t *testing.T) {
	t.Parallel()

	for _, test := range []testParam{
		{
			name:                   "basic test",
			constantRate:           "10/100ms",
			testDuration:           100 * time.Millisecond,
			concurrency:            100,
			iterationDuration:      100 * time.Millisecond,
			distributionType:       "none",
			expectedRunTime:        100 * time.Millisecond,
			expectedCompletedTests: 10,
		},
		{
			name:                   "finishes at ends of duration",
			constantRate:           "10/2s",
			testDuration:           2 * time.Second,
			concurrency:            100,
			iterationDuration:      200 * time.Millisecond,
			distributionType:       "none",
			expectedRunTime:        2 * time.Second,
			expectedCompletedTests: 10,
		},
		{
			name:                   "times out",
			constantRate:           "1/s",
			testDuration:           1 * time.Second,
			concurrency:            100,
			iterationDuration:      2 * time.Second,
			distributionType:       "none",
			expectedRunTime:        2 * time.Second,
			expectedCompletedTests: 1,
		},
		{
			name:                   "next iteration can start if previous still running",
			constantRate:           "10/1s",
			testDuration:           2 * time.Second,
			concurrency:            200,
			iterationDuration:      2 * time.Second,
			distributionType:       "none",
			expectedRunTime:        3 * time.Second,
			expectedCompletedTests: 20,
		},
		{
			name:                      "next iteration won't start if previous still running and limited by concurrency",
			constantRate:              "10/1s",
			testDuration:              2 * time.Second,
			expectedRunTime:           2 * time.Second,
			expectedCompletedTests:    10,
			concurrency:               10,
			iterationDuration:         2 * time.Second,
			distributionType:          "none",
			expectedDroppedIterations: 10,
			expectedFailure:           true,
		},
		{
			name:                   "limited iterations",
			constantRate:           "10/100ms",
			testDuration:           1 * time.Second,
			concurrency:            100,
			iterationDuration:      100 * time.Millisecond,
			maxIterations:          7,
			distributionType:       "none",
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
			distributionType:       "none",
			expectedCompletedTests: 17,
		},
		{
			name:                   "regular distribution of a constant rate",
			constantRate:           "10/s",
			testDuration:           2 * time.Second,
			concurrency:            100,
			iterationDuration:      1 * time.Millisecond,
			distributionType:       "regular",
			expectedRunTime:        2 * time.Second,
			expectedCompletedTests: 20,
		},
		{
			name:                   "random distribution of a constant rate",
			constantRate:           "10/s",
			testDuration:           2 * time.Second,
			concurrency:            100,
			iterationDuration:      1 * time.Millisecond,
			distributionType:       "random",
			expectedRunTime:        2 * time.Second,
			expectedCompletedTests: 20,
		},
		{
			name:                   "run only half of requests on half of the time using regular distribution",
			constantRate:           "10/2s",
			testDuration:           1 * time.Second,
			concurrency:            100,
			iterationDuration:      1 * time.Millisecond,
			distributionType:       "regular",
			expectedRunTime:        1 * time.Second,
			expectedCompletedTests: 5,
		},
		{
			name:                   "simple staged test",
			triggerType:            Staged,
			stages:                 "0ms:0, 50ms: 100, 100ms: 100, 50ms:0",
			iterationFrequency:     "100ms",
			testDuration:           200 * time.Millisecond,
			concurrency:            100,
			iterationDuration:      1 * time.Millisecond,
			distributionType:       "none",
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
			distributionType:       "none",
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
			distributionType:       "regular",
			expectedCompletedTests: 100,
		},
		{
			name:                   "random distribution of a staged trigger",
			triggerType:            Staged,
			stages:                 "0ms:0, 50ms: 100, 100ms: 100, 50ms:0",
			iterationFrequency:     "100ms",
			testDuration:           200 * time.Millisecond,
			concurrency:            100,
			iterationDuration:      1 * time.Millisecond,
			distributionType:       "random",
			expectedCompletedTests: 100,
		},
		{
			name:                   "run half of requests on half of the time on staged test using regular distribution",
			triggerType:            Staged,
			stages:                 "0ms:0, 4s: 100",
			iterationFrequency:     "1s",
			testDuration:           3500 * time.Millisecond,
			concurrency:            100,
			iterationDuration:      1 * time.Millisecond,
			distributionType:       "regular",
			expectedCompletedTests: 112,
		},
		{
			name:                   "users test slow iterations",
			triggerType:            Users,
			testDuration:           1900 * time.Millisecond,
			expectedRunTime:        1900 * time.Millisecond,
			expectedCompletedTests: 10,
			concurrency:            10,
			iterationDuration:      2 * time.Second,
		},
		{
			name:                   "users test normal iterations",
			triggerType:            Users,
			testDuration:           1900 * time.Millisecond,
			expectedRunTime:        1900 * time.Millisecond,
			expectedCompletedTests: 20,
			concurrency:            10,
			iterationDuration:      1 * time.Second,
		},
		{
			name:                   "users test fast iterations",
			triggerType:            Users,
			testDuration:           1900 * time.Millisecond,
			expectedRunTime:        1900 * time.Millisecond,
			expectedCompletedTests: 40,
			concurrency:            10,
			iterationDuration:      500 * time.Millisecond,
		},
		{
			name:                   "simple ramp test",
			triggerType:            Ramp,
			startRate:              "0/100ms",
			endRate:                "10/100ms",
			rampDuration:           "1s",
			testDuration:           1 * time.Second,
			concurrency:            50,
			maxIterations:          1000,
			iterationDuration:      1 * time.Millisecond,
			distributionType:       "none",
			expectedRunTime:        1 * time.Second,
			expectedCompletedTests: 45,
		},
		{
			name:                   "ramp test using max-duration instead of ramp-duration",
			triggerType:            Ramp,
			startRate:              "0/100ms",
			endRate:                "10/100ms",
			testDuration:           1 * time.Second,
			concurrency:            50,
			maxIterations:          1000,
			iterationDuration:      1 * time.Millisecond,
			distributionType:       "none",
			expectedRunTime:        1 * time.Second,
			expectedCompletedTests: 45,
		},
		{
			name:                   "ramp test using max-duration larger than the ramp-duration",
			triggerType:            Ramp,
			startRate:              "0/100ms",
			endRate:                "10/100ms",
			rampDuration:           "500ms",
			testDuration:           1 * time.Second,
			concurrency:            50,
			maxIterations:          1000,
			iterationDuration:      1 * time.Millisecond,
			distributionType:       "none",
			expectedRunTime:        1 * time.Second,
			expectedCompletedTests: 20,
		},
		{
			name:                   "ramp test with 1s default unit rate",
			triggerType:            Ramp,
			startRate:              "10",
			endRate:                "20",
			rampDuration:           "1s",
			testDuration:           1 * time.Second,
			concurrency:            50,
			maxIterations:          1000,
			iterationDuration:      1 * time.Millisecond,
			distributionType:       "none",
			expectedRunTime:        1 * time.Second,
			expectedCompletedTests: 10,
		},
		{
			name:                   "regular distribution of a ramp test",
			triggerType:            Ramp,
			startRate:              "0/100ms",
			endRate:                "10/100ms",
			rampDuration:           "1s",
			testDuration:           1 * time.Second,
			concurrency:            50,
			maxIterations:          1000,
			iterationDuration:      1 * time.Millisecond,
			distributionType:       "regular",
			expectedRunTime:        1 * time.Second,
			expectedCompletedTests: 45,
		},
		{
			name:                   "random distribution of a ramp test",
			triggerType:            Ramp,
			startRate:              "0/100ms",
			endRate:                "10/100ms",
			rampDuration:           "1s",
			testDuration:           1 * time.Second,
			concurrency:            50,
			maxIterations:          1000,
			iterationDuration:      1 * time.Millisecond,
			distributionType:       "random",
			expectedRunTime:        1 * time.Second,
			expectedCompletedTests: 45,
		},
		{
			name:                   "simple config file test",
			triggerType:            File,
			configFile:             "../testdata/config-file.yaml",
			testDuration:           5 * time.Second,
			concurrency:            50,
			maxIterations:          1000,
			iterationDuration:      150 * time.Millisecond,
			expectedRunTime:        1800 * time.Millisecond,
			expectedCompletedTests: 105,
		},
		{
			name:                   "staged users without max iterations",
			triggerType:            File,
			configFile:             "../testdata/config-file-issue-268.yaml",
			testDuration:           5 * time.Second,
			iterationDuration:      0,
			expectedRunTime:        3000 * time.Millisecond,
			expectedCompletedTests: Any,
		},
		{
			name:                   "config file test using limited max-duration",
			triggerType:            File,
			configFile:             "../testdata/config-file.yaml",
			testDuration:           650 * time.Millisecond,
			concurrency:            50,
			maxIterations:          1000,
			iterationDuration:      100 * time.Millisecond,
			expectedRunTime:        700 * time.Millisecond,
			expectedCompletedTests: 60,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			given, when, then := NewRunWithAddTestStage(t)

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

			when.a_timer_is_started().and().
				the_run_command_is_executed()

			then.
				the_command_finished_with_failure_of(test.expectedFailure).and().
				the_command_should_have_run_for_approx(test.expectedRunTime).and().
				the_number_of_started_iterations_should_be(test.expectedCompletedTests).and().
				the_number_of_dropped_iterations_should_be(test.expectedDroppedIterations).and().
				setup_teardown_is_called().and().
				iteration_teardown_is_called_n_times(test.expectedCompletedTests)
		})
	}
}

func TestWithAddNoneDistribution(t *testing.T) {
	t.Parallel()

	given, when, then := NewRunWithAddTestStage(t)

	given.
		a_trigger_type_of(Constant).and().
		a_rate_of("10/s").and().
		a_distribution_type("none").and().
		a_duration_of(500 * time.Millisecond).and().
		a_concurrency_of(50).and().
		an_iteration_limit_of(1000).and().
		a_scenario_where_each_iteration_takes(1 * time.Millisecond)

	when.a_timer_is_started().and().
		the_run_command_is_executed()

	then.there_should_be_x_requests_sent_over_y_intervals_of_z_ms(10, 1, 1000)
}

func TestWithAddRegularDistribution(t *testing.T) {
	t.Parallel()

	given, when, then := NewRunWithAddTestStage(t)

	given.
		a_trigger_type_of(Constant).and().
		a_rate_of("10/s").and().
		a_distribution_type("regular").and().
		a_duration_of(500 * time.Millisecond).and().
		a_concurrency_of(50).and().
		an_iteration_limit_of(1000).and().
		a_scenario_where_each_iteration_takes(1 * time.Millisecond)

	when.a_timer_is_started().and().
		the_run_command_is_executed()

	then.there_should_be_x_requests_sent_over_y_intervals_of_z_ms(1, 5, 100)
}

func TestWithAddRandomDistribution(t *testing.T) {
	t.Parallel()

	given, when, then := NewRunWithAddTestStage(t)

	given.
		a_trigger_type_of(Constant).and().
		a_rate_of("10/s").and().
		a_distribution_type("random").and().
		a_duration_of(1 * time.Second).and().
		a_concurrency_of(50).and().
		an_iteration_limit_of(1000).and().
		a_scenario_where_each_iteration_takes(1 * time.Millisecond)

	when.a_timer_is_started().and().
		the_run_command_is_executed()

	then.the_requests_are_not_sent_all_at_once()
}

func TestWithAddRunScenarioThatFailsSetup(t *testing.T) {
	t.Parallel()

	given, when, then := NewRunWithAddTestStage(t)

	given.
		a_test_scenario_that_always_fails_setup().and().
		a_rate_of("1/s").and().
		a_duration_of(1 * time.Second)

	when.the_run_command_is_executed()

	then.the_command_should_fail().and().
		metrics_are_pushed_to_prometheus()
}

func TestWithAddRunScenarioThatFails(t *testing.T) {
	t.Parallel()

	given, when, then := NewRunWithAddTestStage(t)

	given.
		a_test_scenario_that_always_fails().and().
		a_rate_of("1").and().
		a_duration_of(1 * time.Second)

	when.the_run_command_is_executed()

	then.the_command_should_fail().and().
		setup_teardown_is_called().and().
		iteration_teardown_is_called_n_times(1).and().
		metrics_are_pushed_to_prometheus()
}

func TestWithAddRunScenarioThatPanics(t *testing.T) {
	t.Parallel()

	given, when, then := NewRunWithAddTestStage(t)

	given.
		a_test_scenario_that_always_panics().and().
		a_rate_of("1").and().
		a_duration_of(1 * time.Second)

	when.the_run_command_is_executed()

	then.the_command_should_fail().and().
		setup_teardown_is_called().and().
		iteration_teardown_is_called_n_times(1).and().
		metrics_are_pushed_to_prometheus()
}

func TestWithAddRunScenarioThatFailsAnAssertion(t *testing.T) {
	t.Parallel()

	given, when, then := NewRunWithAddTestStage(t)

	given.
		a_test_scenario_that_always_fails_an_assertion().and().
		a_rate_of("1").and().
		a_duration_of(1 * time.Second)

	when.the_run_command_is_executed()

	then.the_command_should_fail().and().
		setup_teardown_is_called().and().
		iteration_teardown_is_called_n_times(1).and().
		metrics_are_pushed_to_prometheus()
}

func TestWithAddRunScenarioThatFailsOccasionally(t *testing.T) {
	t.Parallel()

	given, when, then := NewRunWithAddTestStage(t)
	given.
		a_test_scenario_that_fails_intermittently().and().
		a_rate_of("100/1s").and().
		// Run less than 1 second, since if we run exactly for 1 second the test might run into another iteration.
		// This would then lead to 200 requests being made, making the test fail
		a_duration_of(500 * time.Millisecond).and().
		a_distribution_type("none")

	when.the_run_command_is_executed()

	then.the_results_should_show_n_failures(50).and().
		the_results_should_show_n_successful_iterations(50).and().
		setup_teardown_is_called().and().
		iteration_teardown_is_called_n_times(100)
}

func TestWithAddInterruptedRun(t *testing.T) {
	t.Parallel()

	given, when, then := NewRunWithAddTestStage(t)

	given.
		a_timer_is_started().
		a_rate_of("5/10ms").and().
		a_duration_of(5 * time.Second).and().
		a_scenario_where_each_iteration_takes(0 * time.Second).and().
		a_distribution_type("none")

	when.
		the_run_command_is_executed_and_cancelled_after(500 * time.Millisecond)

	then.
		setup_teardown_is_called_within(600 * time.Millisecond).and().
		metrics_are_pushed_to_prometheus().and().
		there_is_a_metric_called("form3_loadtest_iteration").and().
		expect_the_stdout_output_to_include([]string{
			"Interrupted - waiting for active tests to complete",
		})
}

func TestWithAddInterruptedRun_TimesOut(t *testing.T) {
	t.Parallel()

	given, when, then := NewRunWithAddTestStage(t)

	given.
		a_timer_is_started().
		a_rate_of("1/1s").and().
		a_duration_of(5 * time.Second).and().
		a_scenario_where_each_iteration_takes(3 * time.Second).and().
		wait_for_completion_timeout_of(1 * time.Second).and().
		a_distribution_type("none")

	when.
		the_run_command_is_executed_and_cancelled_after(500 * time.Millisecond)

	then.
		setup_teardown_is_called_within(600*time.Millisecond + 1*time.Second).and().
		metrics_are_pushed_to_prometheus().and().
		expect_the_stdout_output_to_include([]string{
			"Interrupted - waiting for active tests to complete",
			"Active tests not completed after 1s. Stopping...",
		})

	// needed for goroutine checker
	// sleep is the time left in the interation duration after the timeout
	time.Sleep(3 * time.Second)
}

func TestWithAddMaxDurationReached_TimesOut(t *testing.T) {
	t.Parallel()

	given, when, then := NewRunWithAddTestStage(t)

	given.
		a_timer_is_started().
		a_rate_of("1/1s").and().
		a_duration_of(500 * time.Millisecond).and().
		a_scenario_where_each_iteration_takes(3 * time.Second).and().
		wait_for_completion_timeout_of(1 * time.Second).and().
		a_distribution_type("none")

	when.
		the_run_command_is_executed()

	then.
		setup_teardown_is_called_within(600*time.Millisecond + 1*time.Second).and().
		metrics_are_pushed_to_prometheus().and().
		expect_the_stdout_output_to_include([]string{
			"Max Duration Elapsed - waiting for active tests to complete",
			"Active tests not completed after 1s. Stopping...",
		})

	// needed for goroutine checker
	// sleep is the time left in the interation duration after the timeout
	time.Sleep(3 * time.Second)
}

func TestWithAddFinalRunMetrics(t *testing.T) {
	t.Parallel()

	given, when, then := NewRunWithAddTestStage(t)
	given.
		a_rate_of("100/100ms").and().
		a_duration_of(450 * time.Millisecond).and().
		a_scenario_where_iteration_n_takes_100ms(400)

	when.the_run_command_is_executed()

	then.
		metrics_are_pushed_to_prometheus().and().
		the_100th_percentile_is_slow().and().
		all_other_percentiles_are_fast()
}

func TestWithAddSetupMetricsAreRecorded(t *testing.T) {
	t.Parallel()

	given, when, then := NewRunWithAddTestStage(t)

	given.
		a_rate_of("1/s").and().
		a_scenario_where_each_iteration_takes(1 * time.Millisecond)

	when.the_run_command_is_executed()

	then.
		metrics_are_pushed_to_prometheus().and().
		there_is_a_metric_called("form3_loadtest_setup")
}

func TestWithAddGroupedLabels(t *testing.T) {
	t.Parallel()

	given, when, then := NewRunWithAddTestStage(t)

	given.
		a_rate_of("10/s").and().
		a_scenario_where_each_iteration_takes(1 * time.Millisecond)

	when.the_run_command_is_executed()

	then.
		metrics_are_pushed_to_prometheus().and().
		all_exported_metrics_contain_label("namespace", fakePrometheusNamespace).and().
		all_exported_metrics_contain_label("id", fakePrometheusID)
}

func TestWithAddFailureCounts(t *testing.T) {
	t.Parallel()

	given, when, then := NewRunWithAddTestStage(t)

	given.
		a_rate_of("10/s").and().
		a_duration_of(500 * time.Millisecond).and().
		a_test_scenario_that_fails_intermittently().and().
		a_distribution_type("none")

	when.the_run_command_is_executed()

	then.
		metrics_are_pushed_to_prometheus().and().
		there_is_a_metric_called("form3_loadtest_iteration").and().
		the_iteration_metric_has_n_results(5, "success").and().
		the_iteration_metric_has_n_results(5, "fail")
}

func TestWithAddParameterisedMaxFailures(t *testing.T) {
	t.Parallel()

	for _, test := range []testParam{
		{
			name:            "pass with 5 max failures",
			maxFailures:     5,
			expectedFailure: false,
		},
		{
			name:            "pass with superior max failures",
			maxFailures:     6,
			expectedFailure: false,
		},
		{
			name:            "fails with inferior max failures",
			maxFailures:     3,
			expectedFailure: true,
		},
		{
			name:            "pass with 50% max failures rate",
			maxFailuresRate: 50,
			expectedFailure: false,
		},
		{
			name:            "pass with superior max failures rate",
			maxFailuresRate: 60,
			expectedFailure: false,
		},
		{
			name:            "fails with inferior max failures rate",
			maxFailuresRate: 30,
			expectedFailure: true,
		},
		{
			name:            "pass with inferior max failures and max failures rate",
			maxFailures:     3,
			maxFailuresRate: 30,
			expectedFailure: true,
		},
		{
			name:            "fails with inferior max failures rate and superior max failures",
			maxFailures:     6,
			maxFailuresRate: 30,
			expectedFailure: true,
		},
		{
			name:            "fails with inferior max failures and superior max failures rate",
			maxFailures:     3,
			maxFailuresRate: 60,
			expectedFailure: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			given, when, then := NewRunWithAddTestStage(t)

			given.
				a_rate_of("10/100ms").and().
				a_max_failures_of(test.maxFailures).and().
				a_max_failures_rate_of(test.maxFailuresRate).and().
				a_duration_of(100 * time.Millisecond).and().
				a_test_scenario_that_fails_intermittently().and().
				a_distribution_type("none")

			when.the_run_command_is_executed()

			then.
				the_iteration_metric_has_n_results(5, "success").and().
				the_iteration_metric_has_n_results(5, "fail").and().
				the_command_finished_with_failure_of(test.expectedFailure)
		})
	}
}

func TestWithAddOutput_JSONLogging(t *testing.T) {
	t.Parallel()

	uiOnlyLogs := []logFieldMatchers{
		{
			"message":  anyValue,
			"level":    "info",
			"scenario": "scenario_where_each_iteration_takes_200ms",
		},
		{
			"message":  "Running scenario_where_each_iteration_takes_200ms for up to 10 iterations or up to 10s at a rate of Makes requests from a set of users specified by --concurrency",
			"level":    "info",
			"scenario": "scenario_where_each_iteration_takes_200ms",
		},
		{
			"message":         "progress",
			"level":           "info",
			"scenario":        "scenario_where_each_iteration_takes_200ms",
			"iteration_stats": anyValue,
		},
		{
			"message":         "progress",
			"level":           "info",
			"scenario":        "scenario_where_each_iteration_takes_200ms",
			"iteration_stats": anyValue,
		},
		{
			"message":  "Max Iterations Reached - waiting for active tests to complete",
			"level":    "info",
			"scenario": "scenario_where_each_iteration_takes_200ms",
			"duration": anyValue,
		},
		{
			"message":  "teardown completed",
			"level":    "info",
			"scenario": "scenario_where_each_iteration_takes_200ms",
		},
		{
			"message":         "Load Test Passed",
			"level":           "info",
			"scenario":        "scenario_where_each_iteration_takes_200ms",
			"iteration_stats": anyValue,
		},
	}

	scenarioOnlyLogs := []logFieldMatchers{
		{
			"message":  "setup",
			"level":    "info",
			"scenario": "scenario_where_each_iteration_takes_200ms",
		},
		{
			"message":  "logrus - setup",
			"level":    "info",
			"logger":   "logrus",
			"scenario": "scenario_where_each_iteration_takes_200ms",
		},

		{
			"message":  "first iteration",
			"level":    "info",
			"scenario": "scenario_where_each_iteration_takes_200ms",
		},
		{
			"message":  "logrus - first iteration",
			"level":    "info",
			"logger":   "logrus",
			"scenario": "scenario_where_each_iteration_takes_200ms",
		},
	}

	uiAndScenarioLogs := []logFieldMatchers{
		{
			"message":  "Running scenario_where_each_iteration_takes_200ms for up to 10 iterations or up to 10s at a rate of Makes requests from a set of users specified by --concurrency",
			"level":    "info",
			"scenario": "scenario_where_each_iteration_takes_200ms",
		},
		{
			"message":  "setup",
			"level":    "info",
			"scenario": "scenario_where_each_iteration_takes_200ms",
		},
		{
			"message":  "logrus - setup",
			"level":    "info",
			"logger":   "logrus",
			"scenario": "scenario_where_each_iteration_takes_200ms",
		},

		{
			"message":  "first iteration",
			"level":    "info",
			"scenario": "scenario_where_each_iteration_takes_200ms",
		},
		{
			"message":  "logrus - first iteration",
			"level":    "info",
			"logger":   "logrus",
			"scenario": "scenario_where_each_iteration_takes_200ms",
		},

		{
			"message":         "progress",
			"level":           "info",
			"scenario":        "scenario_where_each_iteration_takes_200ms",
			"iteration_stats": anyValue,
		},
		{
			"message":         "progress",
			"level":           "info",
			"scenario":        "scenario_where_each_iteration_takes_200ms",
			"iteration_stats": anyValue,
		},
		{
			"message":  "Max Iterations Reached - waiting for active tests to complete",
			"level":    "info",
			"scenario": "scenario_where_each_iteration_takes_200ms",
			"duration": anyValue,
		},
		{
			"message":  "teardown completed",
			"level":    "info",
			"scenario": "scenario_where_each_iteration_takes_200ms",
		},
		{
			"message":         "Load Test Passed",
			"level":           "info",
			"scenario":        "scenario_where_each_iteration_takes_200ms",
			"iteration_stats": anyValue,
		},
	}

	testCases := []struct {
		name                    string
		verbose                 bool
		interactive             bool
		logFilePath             string
		expectedStdoutContains  []string
		expectedStdoutLogLines  []logFieldMatchers
		expectedStderrLogLines  []logFieldMatchers
		expectedLogFileLogLines []logFieldMatchers
	}{
		{
			name:        "interactive - scenario logs to file & human readable to stdout",
			interactive: true,
			verbose:     false,
			expectedStdoutContains: []string{
				"Saving logs to",
				"F1 Load Tester",
				"Running scenario_where_each_iteration_takes_200ms scenario for up to 10 iterations or up to 10s at a rate of Makes requests from a set of users specified by --concurrency.",
				"[   1s]",
				"[   2s]  Max Iterations Reached - waiting for active tests to complete",
				"[Teardown] ✔",
				"Load Test Passed",
				"10 iterations started in",
				"Full logs:",
			},
			expectedLogFileLogLines: scenarioOnlyLogs,
		},
		{
			name:                   "interactive verbose - only structured logs to stdout",
			interactive:            true,
			verbose:                true,
			expectedStdoutLogLines: uiAndScenarioLogs,
		},
		{
			name:                    "non interactive - structured logs to stdout & scenario logs to file",
			interactive:             false,
			verbose:                 false,
			expectedStdoutLogLines:  uiOnlyLogs,
			expectedLogFileLogLines: scenarioOnlyLogs,
		},
		{
			name:                   "non interactive verbose - only structured logs to stdout",
			interactive:            false,
			verbose:                true,
			expectedStdoutLogLines: uiAndScenarioLogs,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			given, when, then := NewRunWithAddTestStage(t)

			given.
				verbose_flag_is(testCase.verbose).and().
				json_logging_is_enabled().and().
				terminal_is_interactive(testCase.interactive).and().
				a_trigger_type_of(Users).and().
				a_scenario_where_each_iteration_takes(200 * time.Millisecond).and().
				a_duration_of(10 * time.Second).and().
				a_concurrency_of(1).and().
				an_iteration_limit_of(10)

			when.
				the_run_command_is_executed()

			then.
				the_command_finished_successfully().and().
				expect_stderr_to_match_json_log(testCase.expectedStderrLogLines).and().
				expect_the_logfile_to_match_json_log(testCase.expectedLogFileLogLines)

			if len(testCase.expectedStdoutContains) != 0 {
				then.expect_the_stdout_output_to_include(testCase.expectedStdoutContains)
			} else {
				then.expect_stdout_to_match_json_log(testCase.expectedStdoutLogLines)
			}
		})
	}
}
