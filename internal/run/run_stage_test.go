package run_test

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/form3tech-oss/f1/v2/internal/console"
	"github.com/form3tech-oss/f1/v2/internal/envsettings"
	"github.com/form3tech-oss/f1/v2/internal/metrics"
	"github.com/form3tech-oss/f1/v2/internal/options"
	"github.com/form3tech-oss/f1/v2/internal/run"
	"github.com/form3tech-oss/f1/v2/internal/trigger/api"
	"github.com/form3tech-oss/f1/v2/internal/trigger/constant"
	"github.com/form3tech-oss/f1/v2/internal/trigger/file"
	"github.com/form3tech-oss/f1/v2/internal/trigger/ramp"
	"github.com/form3tech-oss/f1/v2/internal/trigger/staged"
	"github.com/form3tech-oss/f1/v2/internal/trigger/users"
	"github.com/form3tech-oss/f1/v2/pkg/f1"
	f1_testing "github.com/form3tech-oss/f1/v2/pkg/f1/testing"
)

const (
	fakePrometheusNamespace = "test-namespace"
	fakePrometheusID        = "test-run-name"
	iterationMetricFamily   = "form3_loadtest_iteration"
)

type RunTestStage struct {
	startTime              time.Time
	runError               error
	metrics                *metrics.Metrics
	printer                *console.Printer
	runInstance            *run.Run
	runResult              *run.Result
	t                      *testing.T
	require                *require.Assertions
	metricData             *MetricData
	scenarioCleanup        func()
	assert                 *assert.Assertions
	iterationCleanup       func()
	f1                     *f1.F1
	durations              sync.Map
	frequency              string
	rate                   string
	stages                 string
	distributionType       string
	configFile             string
	startRate              string
	endRate                string
	rampDuration           string
	scenario               string
	settings               envsettings.Settings
	maxFailures            uint64
	maxIterations          uint64
	maxFailuresRate        int
	duration               time.Duration
	concurrency            int
	triggerType            TriggerType
	iterationTeardownCount atomic.Uint32
	setupTeardownCount     atomic.Uint32
	runCount               atomic.Uint32
}

func NewRunTestStage(t *testing.T) (*RunTestStage, *RunTestStage, *RunTestStage) {
	t.Helper()
	stage := &RunTestStage{
		t:           t,
		concurrency: 100,
		assert:      assert.New(t),
		require:     require.New(t),
		f1:          f1.New(),
		settings:    envsettings.Get(),
		metricData:  NewMetricData(),
		printer:     console.NewPrinter(io.Discard, io.Discard),
		metrics:     metrics.NewInstance(prometheus.NewRegistry(), true),
	}

	handler := FakePrometheusHandler(t, stage.metricData)
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)

	stage.settings.Prometheus.PushGateway = ts.URL
	stage.settings.Prometheus.LabelID = fakePrometheusID
	stage.settings.Prometheus.Namespace = fakePrometheusNamespace

	stage.scenarioCleanup = func() { stage.setupTeardownCount.Add(1) }
	stage.iterationCleanup = func() { stage.iterationTeardownCount.Add(1) }

	return stage, stage, stage
}

func (s *RunTestStage) a_rate_of(rate string) *RunTestStage {
	s.rate = rate
	return s
}

func (s *RunTestStage) and() *RunTestStage {
	return s
}

func (s *RunTestStage) a_duration_of(i time.Duration) *RunTestStage {
	s.duration = i
	return s
}

func (s *RunTestStage) a_concurrency_of(concurrency int) *RunTestStage {
	s.concurrency = concurrency
	return s
}

func (s *RunTestStage) a_max_failures_of(maxFailures uint64) *RunTestStage {
	s.maxFailures = maxFailures
	return s
}

func (s *RunTestStage) a_max_failures_rate_of(maxFailuresRate int) *RunTestStage {
	s.maxFailuresRate = maxFailuresRate
	return s
}

func (s *RunTestStage) a_config_file_location_of(commandsFile string) *RunTestStage {
	s.configFile = commandsFile
	return s
}

func (s *RunTestStage) a_start_rate_of(startRate string) *RunTestStage {
	s.startRate = startRate
	return s
}

func (s *RunTestStage) a_end_rate_of(endRate string) *RunTestStage {
	s.endRate = endRate
	return s
}

func (s *RunTestStage) a_ramp_duration_of(rampDuration string) *RunTestStage {
	s.rampDuration = rampDuration
	return s
}

func (s *RunTestStage) setupRun() {
	r, err := run.NewRun(
		options.RunOptions{
			Scenario:        s.scenario,
			MaxDuration:     s.duration,
			Concurrency:     s.concurrency,
			MaxIterations:   s.maxIterations,
			MaxFailures:     s.maxFailures,
			MaxFailuresRate: s.maxFailuresRate,
		},
		s.build_trigger(), s.settings, s.metrics, s.printer)
	if err != nil {
		s.runError = fmt.Errorf("run create: %w", err)
	}
	s.runInstance = r
}

func (s *RunTestStage) the_run_command_is_executed() *RunTestStage {
	s.setupRun()

	var err error
	s.runResult, err = s.runInstance.Do(context.TODO(), s.f1.GetScenarios())
	if err != nil {
		s.runError = fmt.Errorf("run do: %w", err)
		return s
	}

	return s
}

func (s *RunTestStage) the_run_command_is_executed_and_cancelled_after(duration time.Duration) *RunTestStage {
	s.setupRun()

	var err error
	ctx, cancel := context.WithTimeout(context.TODO(), duration)
	defer cancel()

	s.runResult, err = s.runInstance.Do(ctx, s.f1.GetScenarios())
	if err != nil {
		s.runError = fmt.Errorf("run do: %w", err)
		return s
	}

	return s
}

func (s *RunTestStage) a_timer_is_started() *RunTestStage {
	s.startTime = time.Now()
	return s
}

func (s *RunTestStage) the_command_should_have_run_for_approx(expectedDuration time.Duration) *RunTestStage {
	if expectedDuration > 0 {
		diff := s.runResult.TestDuration - expectedDuration
		// Generally, we want timings to be within 100ms of our expected values, but where the expectation
		// is greater than 1 second, within 500ms is close enough.
		marginForError := (100 * time.Millisecond).Seconds()
		if expectedDuration.Seconds() > 1 {
			marginForError = (500 * time.Millisecond).Seconds()
		}
		msg := fmt.Sprintf(
			"difference between expected (%fs) an actual (%fs) durations was more than %fs",
			expectedDuration.Seconds(),
			s.runResult.TestDuration.Seconds(),
			marginForError)
		s.assert.LessOrEqual(math.Abs(diff.Seconds()), marginForError, msg)
	}
	return s
}

func (s *RunTestStage) the_number_of_started_iterations_should_be(expected uint32) *RunTestStage {
	s.assert.Equal(int(expected), int(s.runCount.Load()), "number of started iterations")
	return s
}

func (s *RunTestStage) the_command_should_fail() *RunTestStage {
	s.assert.NotNil(s.runResult, "run result is nil")
	s.assert.True(s.runResult.Failed(), "command did not fail")
	return s
}

func (s *RunTestStage) a_test_scenario_that_always_fails() *RunTestStage {
	s.scenario = "scenario_that_always_fails"
	s.f1.Add(s.scenario, func(scenarioT *f1_testing.T) f1_testing.RunFn {
		scenarioT.Cleanup(s.scenarioCleanup)

		return func(iterationT *f1_testing.T) {
			iterationT.Cleanup(s.iterationCleanup)

			iterationT.FailNow()
		}
	})
	return s
}

func (s *RunTestStage) a_test_scenario_that_always_panics() *RunTestStage {
	s.scenario = "scenario_that_always_panics"
	s.f1.Add(s.scenario, func(scenarioT *f1_testing.T) f1_testing.RunFn {
		scenarioT.Cleanup(s.scenarioCleanup)

		return func(iterationT *f1_testing.T) {
			iterationT.Cleanup(s.iterationCleanup)

			panic("test panic in scenario iteration")
		}
	})
	return s
}

func (s *RunTestStage) a_test_scenario_that_always_fails_an_assertion() *RunTestStage {
	s.scenario = "scenario_that_always_fails_an_assertion"
	s.f1.Add(s.scenario, func(scenarioT *f1_testing.T) f1_testing.RunFn {
		scenarioT.Cleanup(s.scenarioCleanup)

		return func(iterationT *f1_testing.T) {
			iterationT.Cleanup(s.iterationCleanup)

			assert.True(iterationT, false)
		}
	})
	return s
}

func (s *RunTestStage) a_test_scenario_that_always_fails_setup() *RunTestStage {
	s.scenario = "scenario_that_always_fails_setup"
	s.f1.Add(s.scenario, func(scenarioT *f1_testing.T) f1_testing.RunFn {
		scenarioT.Cleanup(s.scenarioCleanup)

		scenarioT.FailNow()
		return nil
	})
	return s
}

func (s *RunTestStage) a_scenario_where_each_iteration_takes(duration time.Duration) *RunTestStage {
	s.scenario = "scenario_where_each_iteration_takes_" + duration.String()
	s.f1.Add(s.scenario, func(scenarioT *f1_testing.T) f1_testing.RunFn {
		scenarioT.Cleanup(s.scenarioCleanup)

		s.runCount.Store(0)

		return func(iterationT *f1_testing.T) {
			iterationT.Cleanup(s.iterationCleanup)

			s.runCount.Add(1)
			s.durations.Store(time.Now(), time.Since(s.startTime))
			time.Sleep(duration)
		}
	})
	return s
}

func (s *RunTestStage) setup_teardown_is_called() *RunTestStage {
	s.assert.Equal(1, int(s.setupTeardownCount.Load()), "setup teardown was not called")
	return s
}

func (s *RunTestStage) iteration_teardown_is_called_n_times(n uint32) *RunTestStage {
	s.assert.Equal(int(n), int(s.iterationTeardownCount.Load()), "iteration teardown was not called expected times")
	return s
}

func (s *RunTestStage) a_test_scenario_that_fails_intermittently() *RunTestStage {
	s.scenario = "scenario_that_fails_intermittently"
	s.f1.Add(s.scenario, func(scenarioT *f1_testing.T) f1_testing.RunFn {
		scenarioT.Cleanup(s.scenarioCleanup)

		s.runCount.Store(0)
		return func(t *f1_testing.T) {
			t.Cleanup(s.iterationCleanup)

			count := s.runCount.Add(1)
			t.Require().Equal(uint32(0), count%2)
		}
	})
	return s
}

func (s *RunTestStage) the_results_should_show_n_failures(expectedFailures uint64) *RunTestStage {
	s.assert.Equal(int(expectedFailures), int(s.runResult.Snapshot().FailedIterationDurations.Count), "failure count does not match expected")
	return s
}

func (s *RunTestStage) the_results_should_show_n_successful_iterations(expected uint64) *RunTestStage {
	s.assert.Equal(int(expected), int(s.runResult.Snapshot().SuccessfulIterationDurations.Count), "success count does not match expected")
	return s
}

func (s *RunTestStage) the_number_of_dropped_iterations_should_be(expected uint64) *RunTestStage {
	s.assert.Equal(int(expected), int(s.runResult.Snapshot().DroppedIterationCount), "dropped count does not match expected")
	return s
}

func (s *RunTestStage) distribution_duration_map_of_requests() map[time.Duration]int32 {
	distributionMap := make(map[time.Duration]int32)
	s.durations.Range(func(_, value interface{}) bool {
		requestDuration, ok := value.(time.Duration)
		s.require.True(ok)
		truncatedDuration := requestDuration.Truncate(100 * time.Millisecond)
		existingDuration := distributionMap[truncatedDuration] + 1
		distributionMap[truncatedDuration] = existingDuration
		return true
	})

	return distributionMap
}

func (s *RunTestStage) there_should_be_x_requests_sent_over_y_intervals_of_z_ms(requests, intervals, ms int) *RunTestStage {
	expectedDistribution := map[time.Duration]int32{}
	for i := range intervals {
		key := time.Duration(i) * time.Duration(ms) * time.Millisecond
		expectedDistribution[key] = int32(requests)
	}

	distributionMap := s.distribution_duration_map_of_requests()

	s.assert.Equal(expectedDistribution, distributionMap)

	return s
}

func (s *RunTestStage) the_requests_are_not_sent_all_at_once() *RunTestStage {
	distributionMap := s.distribution_duration_map_of_requests()

	s.assert.Greater(len(distributionMap), 1, "unexpected distribution: %v", distributionMap)

	return s
}

func (s *RunTestStage) the_command_finished_with_failure_of(expected bool) *RunTestStage {
	s.assert.Equal(expected, s.runResult.Failed(), "command failed")
	return s
}

func (s *RunTestStage) an_iteration_limit_of(iterations uint64) *RunTestStage {
	s.maxIterations = iterations
	return s
}

func (s *RunTestStage) build_trigger() *api.Trigger {
	var t *api.Trigger
	var err error
	switch s.triggerType {
	case Constant:
		flags := constant.Rate().Flags

		err = flags.Set("rate", s.rate)
		require.NoError(s.t, err)

		if s.distributionType != "" {
			err = flags.Set("distribution", s.distributionType)
			require.NoError(s.t, err)
		}

		t, err = constant.Rate().New(flags)
		require.NoError(s.t, err)
	case Staged:
		flags := staged.Rate().Flags

		err = flags.Set("stages", s.stages)
		require.NoError(s.t, err)

		err = flags.Set("iterationFrequency", s.frequency)
		require.NoError(s.t, err)

		if s.distributionType != "" {
			err = flags.Set("distribution", s.distributionType)
			require.NoError(s.t, err)
		}

		t, err = staged.Rate().New(flags)
		require.NoError(s.t, err)
	case Users:
		flags := users.Rate().Flags
		t, err = users.Rate().New(flags)
		require.NoError(s.t, err)
	case Ramp:
		flags := ramp.Rate().Flags

		err = flags.Set("start-rate", s.startRate)
		require.NoError(s.t, err)

		err = flags.Set("end-rate", s.endRate)
		require.NoError(s.t, err)

		if s.rampDuration != "" {
			err = flags.Set("ramp-duration", s.rampDuration)
			require.NoError(s.t, err)
		} else {
			flags.DurationP("max-duration", "d", time.Second, "--max-duration 1s (stop after 1 second)")
			err = flags.Set("max-duration", s.duration.String())
			require.NoError(s.t, err)
		}

		if s.distributionType != "" {
			err = flags.Set("distribution", s.distributionType)
			require.NoError(s.t, err)
		}

		t, err = ramp.Rate().New(flags)
		require.NoError(s.t, err)
	case File:
		flags := file.Rate(s.printer).Flags

		err := flags.Parse([]string{s.configFile})
		require.NoError(s.t, err)

		t, err = file.Rate(s.printer).New(flags)
		require.NoError(s.t, err)
	}
	return t
}

func (s *RunTestStage) setup_teardown_is_called_within(duration time.Duration) *RunTestStage {
	s.setup_teardown_is_called()

	s.assert.WithinDuration(s.startTime, time.Now(), duration)

	return s
}

func (s *RunTestStage) a_trigger_type_of(triggerType TriggerType) *RunTestStage {
	s.triggerType = triggerType
	return s
}

func (s *RunTestStage) a_stage_of(stages string) *RunTestStage {
	s.stages = stages
	return s
}

func (s *RunTestStage) an_iteration_frequency_of(frequency string) *RunTestStage {
	s.frequency = frequency
	return s
}

func (s *RunTestStage) a_distribution_type(distributionType string) *RunTestStage {
	s.distributionType = distributionType
	return s
}

func (s *RunTestStage) metrics_are_pushed_to_prometheus() *RunTestStage {
	s.assert.False(s.metricData.Empty(), "metric data is empty")
	return s
}

func (s *RunTestStage) a_scenario_where_iteration_n_takes_100ms(n uint32) *RunTestStage {
	s.scenario = fmt.Sprintf("scenario_where_iteration_%d_takes_100ms", n)
	s.f1.Add(s.scenario, func(scenarioT *f1_testing.T) f1_testing.RunFn {
		scenarioT.Cleanup(s.scenarioCleanup)

		s.runCount.Store(0)

		return func(iterationT *f1_testing.T) {
			iterationT.Cleanup(s.iterationCleanup)

			current := s.runCount.Add(1)
			if current == n {
				time.Sleep(100 * time.Millisecond)
			}
		}
	})
	return s
}

func (s *RunTestStage) the_100th_percentile_is_slow() *RunTestStage {
	s.assert.GreaterOrEqual(s.metricData.GetIterationDuration(s.scenario, 1.0), float64(100*time.Millisecond))
	return s
}

func (s *RunTestStage) all_other_percentiles_are_fast() *RunTestStage {
	s.assert.Greater(s.metricData.GetIterationDuration(s.scenario, 0.9), 0.0)
	s.assert.LessOrEqual(s.metricData.GetIterationDuration(s.scenario, 0.9), float64(25*time.Millisecond))
	s.assert.Greater(s.metricData.GetIterationDuration(s.scenario, 0.95), 0.0)
	s.assert.LessOrEqual(s.metricData.GetIterationDuration(s.scenario, 0.95), float64(25*time.Millisecond))
	return s
}

func (s *RunTestStage) there_is_a_metric_called(metricName string) *RunTestStage {
	err := retry(func() error {
		metricNames := s.metricData.GetMetricNames()
		for _, mn := range metricNames {
			if mn == metricName {
				return nil
			}
		}
		return fmt.Errorf("%v did not contain %s", metricNames, metricName)
	}, 10, 50*time.Millisecond)
	s.require.NoError(err)
	return s
}

func (s *RunTestStage) the_iteration_metric_has_n_results(n int, result string) *RunTestStage {
	err := retry(func() error {
		metricFamily := s.metricData.GetMetricFamily(iterationMetricFamily)
		s.require.NotNil(metricFamily, "metric family %s not found", iterationMetricFamily)
		resultMetric := getMetricByResult(metricFamily, result)
		s.require.NotNil(resultMetric, "result metric %s is empty", result)
		if uint64(n) == resultMetric.GetSummary().GetSampleCount() {
			return nil
		}
		return fmt.Errorf("expected %d to equal %d", uint64(n), resultMetric.GetSummary().GetSampleCount())
	}, 10, 50*time.Millisecond)
	s.require.NoError(err)
	return s
}

func (s *RunTestStage) all_exported_metrics_contain_label(labelName string, labelValue string) *RunTestStage {
	metricNames := s.metricData.GetMetricNames()

	for _, name := range metricNames {
		metricFamily := s.metricData.GetMetricFamily(name)
		s.require.NotNil(metricFamily)

		for _, metric := range metricFamily.GetMetric() {
			match := false
			for _, label := range metric.GetLabel() {
				nameMatch := label.GetName() == labelName
				valueMatch := label.GetValue() == labelValue
				match = match || (nameMatch && valueMatch)
			}

			if !match {
				openMetrics := strings.Builder{}
				_, _ = expfmt.MetricFamilyToOpenMetrics(&openMetrics, metricFamily)
				s.require.FailNowf("Label is missing", "Metric %q do not have label %q with value %q:\n%s",
					metricFamily.GetName(), labelName, labelValue, openMetrics.String())
			}
		}
	}
	return s
}

func getMetricByResult(metricFamily *io_prometheus_client.MetricFamily, result string) *io_prometheus_client.Metric {
	for _, metric := range metricFamily.GetMetric() {
		for _, label := range metric.GetLabel() {
			if label.GetName() == "result" && label.GetValue() == result {
				return metric
			}
		}
	}
	return nil
}

func retry(retryable func() error, retries int, delay time.Duration) error {
	var err error
	for range retries {
		err = retryable()
		if err == nil {
			return nil
		}

		time.Sleep(delay)
	}
	return err
}
