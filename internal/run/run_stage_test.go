package run_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/google/uuid"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/form3tech-oss/f1/v2/internal/envsettings"
	"github.com/form3tech-oss/f1/v2/internal/fluentd"
	"github.com/form3tech-oss/f1/v2/internal/options"
	"github.com/form3tech-oss/f1/v2/internal/run"
	"github.com/form3tech-oss/f1/v2/internal/trace"
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
)

type RunTestStage struct {
	duration               time.Duration
	runCount               int32
	startTime              time.Time
	t                      *testing.T
	scenario               string
	runResult              *run.Result
	runError               error
	concurrency            int
	setupTeardownCount     *int32
	iterationTeardownCount *int32
	assert                 *assert.Assertions
	rate                   string
	maxIterations          int32
	maxFailures            int
	maxFailuresRate        int
	triggerType            TriggerType
	stages                 string
	frequency              string
	require                *require.Assertions
	distributionType       string
	configFile             string
	startRate              string
	endRate                string
	rampDuration           string
	durations              sync.Map
	f1                     *f1.F1

	tracer       trace.Tracer
	tracerWg     sync.WaitGroup
	tracerOutput bytes.Buffer
	tracerWriter *io.PipeWriter
	tracerReader *io.PipeReader

	settings envsettings.Settings

	metricData *MetricData
}

func NewRunTestStage(t *testing.T) (*RunTestStage, *RunTestStage, *RunTestStage) {
	t.Helper()
	setupTeardownCount := int32(0)
	iterationTeardownCount := int32(0)
	stage := &RunTestStage{
		t:                      t,
		concurrency:            100,
		assert:                 assert.New(t),
		require:                require.New(t),
		setupTeardownCount:     &setupTeardownCount,
		iterationTeardownCount: &iterationTeardownCount,
		f1:                     f1.New(),
		tracer:                 trace.NewConsoleTracer(io.Discard),
		settings:               envsettings.Get(),
		metricData:             NewMetricData(),
	}

	handler := FakePrometheusHandler(t, stage.metricData)
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)

	stage.settings.Prometheus.PushGateway = ts.URL
	stage.settings.Prometheus.LabelID = fakePrometheusID
	stage.settings.Prometheus.Namespace = fakePrometheusNamespace

	stage.tracerReader, stage.tracerWriter = io.Pipe()

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

func (s *RunTestStage) a_max_failures_of(maxFailures int) *RunTestStage {
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

func (s *RunTestStage) i_execute_the_run_command() *RunTestStage {
	r, err := run.NewRun(
		options.RunOptions{
			Scenario:            s.scenario,
			MaxDuration:         s.duration,
			Concurrency:         s.concurrency,
			MaxIterations:       s.maxIterations,
			MaxFailures:         s.maxFailures,
			MaxFailuresRate:     s.maxFailuresRate,
			RegisterLogHookFunc: fluentd.LoggingHook(s.settings.Fluentd.Host, s.settings.Fluentd.Port),
		},
		s.build_trigger(), s.settings, s.tracer)
	if err != nil {
		s.runError = fmt.Errorf("run create: %w", err)
		return s
	}

	s.runResult, err = r.Do(s.f1.GetScenarios())
	if err != nil {
		s.runError = fmt.Errorf("run do: %w", err)
		return s
	}

	return s
}

func (s *RunTestStage) i_start_a_timer() *RunTestStage {
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

func (s *RunTestStage) the_number_of_started_iterations_should_be(expected int32) *RunTestStage {
	s.assert.Equal(expected, s.runCount, "number of started iterations")
	return s
}

func (s *RunTestStage) the_command_should_fail() *RunTestStage {
	s.assert.NotNil(s.runResult)
	s.assert.True(s.runResult.Failed())
	return s
}

func (s *RunTestStage) a_test_scenario_that_always_fails() *RunTestStage {
	s.scenario = uuid.New().String()
	s.f1.Add(s.scenario, func(t *f1_testing.T) f1_testing.RunFn {
		t.Cleanup(func() {
			atomic.AddInt32(s.setupTeardownCount, 1)
		})
		return func(t *f1_testing.T) {
			t.Cleanup(func() {
				atomic.AddInt32(s.iterationTeardownCount, 1)
			})
			t.FailNow()
		}
	})
	return s
}

func (s *RunTestStage) a_test_scenario_that_always_panics() *RunTestStage {
	s.scenario = uuid.New().String()
	s.f1.Add(s.scenario, func(t *f1_testing.T) f1_testing.RunFn {
		t.Cleanup(func() {
			atomic.AddInt32(s.setupTeardownCount, 1)
		})
		return func(t *f1_testing.T) {
			t.Cleanup(func() {
				atomic.AddInt32(s.iterationTeardownCount, 1)
			})
			panic("aaargh!")
		}
	})
	return s
}

func (s *RunTestStage) a_test_scenario_that_always_fails_an_assertion() *RunTestStage {
	s.scenario = uuid.New().String()
	s.f1.Add(s.scenario, func(t *f1_testing.T) f1_testing.RunFn {
		t.Cleanup(func() {
			atomic.AddInt32(s.setupTeardownCount, 1)
		})
		return func(t *f1_testing.T) {
			t.Cleanup(func() {
				atomic.AddInt32(s.iterationTeardownCount, 1)
			})
			assert.Equal(t, 3, 1+1)
		}
	})
	return s
}

func (s *RunTestStage) a_test_scenario_that_always_fails_setup() *RunTestStage {
	s.scenario = uuid.New().String()
	s.f1.Add(s.scenario, func(t *f1_testing.T) f1_testing.RunFn {
		t.Cleanup(func() {
			atomic.AddInt32(s.setupTeardownCount, 1)
		})
		t.FailNow()
		return nil
	})
	return s
}

func (s *RunTestStage) a_scenario_where_each_iteration_takes(duration time.Duration) *RunTestStage {
	s.scenario = uuid.New().String()
	s.f1.Add(s.scenario, func(t *f1_testing.T) f1_testing.RunFn {
		t.Cleanup(func() {
			atomic.AddInt32(s.setupTeardownCount, 1)
		})
		s.runCount = 0
		return func(t *f1_testing.T) {
			t.Cleanup(func() {
				atomic.AddInt32(s.iterationTeardownCount, 1)
			})
			atomic.AddInt32(&s.runCount, 1)
			s.durations.Store(time.Now(), time.Since(s.startTime))
			time.Sleep(duration)
		}
	})
	return s
}

func (s *RunTestStage) setup_teardown_is_called() *RunTestStage {
	s.assert.Equal(int32(1), atomic.LoadInt32(s.setupTeardownCount))
	return s
}

func (s *RunTestStage) iteration_teardown_is_called_n_times(n int32) *RunTestStage {
	s.assert.Equal(n, atomic.LoadInt32(s.iterationTeardownCount))
	return s
}

func (s *RunTestStage) a_test_scenario_that_fails_intermittently() *RunTestStage {
	s.scenario = uuid.New().String()
	s.f1.Add(s.scenario, func(t *f1_testing.T) f1_testing.RunFn {
		t.Cleanup(func() {
			atomic.AddInt32(s.setupTeardownCount, 1)
		})
		s.runCount = 0
		return func(t *f1_testing.T) {
			t.Cleanup(func() {
				atomic.AddInt32(s.iterationTeardownCount, 1)
			})
			count := atomic.AddInt32(&s.runCount, 1)
			t.Require().Equal(int32(0), count%2)
		}
	})
	return s
}

func (s *RunTestStage) the_results_should_show_n_failures(expectedFailures uint64) *RunTestStage {
	s.assert.Equal(expectedFailures, s.runResult.FailedIterationCount)
	return s
}

func (s *RunTestStage) the_results_should_show_n_successful_iterations(expected uint64) *RunTestStage {
	s.assert.Equal(expected, s.runResult.SuccessfulIterationCount)
	return s
}

func (s *RunTestStage) the_number_of_dropped_iterations_should_be(expected uint64) *RunTestStage {
	s.assert.Equal(int(expected), int(s.runResult.DroppedIterationCount))
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

	s.assert.Greater(len(distributionMap), 1)

	return s
}

func (s *RunTestStage) the_command_finished_with_failure_of(expected bool) *RunTestStage {
	s.assert.Equal(expected, s.runResult.Failed(), "command failed")
	return s
}

func (s *RunTestStage) an_iteration_limit_of(iterations int32) *RunTestStage {
	s.maxIterations = iterations
	return s
}

func (s *RunTestStage) the_test_run_is_started() *RunTestStage {
	go func() {
		r, err := run.NewRun(
			options.RunOptions{
				Scenario:            s.scenario,
				MaxDuration:         s.duration,
				Concurrency:         s.concurrency,
				MaxIterations:       s.maxIterations,
				RegisterLogHookFunc: fluentd.LoggingHook("", ""),
			},
			s.build_trigger(), s.settings, s.tracer)
		if err != nil {
			s.runError = fmt.Errorf("new run: %w", err)
			return
		}

		s.runResult, s.runError = r.Do(s.f1.GetScenarios())
	}()
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

		t, err = constant.Rate().New(flags, s.tracer)
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

		t, err = staged.Rate().New(flags, s.tracer)
		require.NoError(s.t, err)
	case Users:
		flags := users.Rate().Flags
		t, err = users.Rate().New(flags, s.tracer)
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

		t, err = ramp.Rate().New(flags, s.tracer)
		require.NoError(s.t, err)
	case File:
		flags := file.Rate().Flags

		err := flags.Parse([]string{s.configFile})
		require.NoError(s.t, err)

		t, err = file.Rate().New(flags, s.tracer)
		require.NoError(s.t, err)
	}
	return t
}

func (s *RunTestStage) the_test_run_is_interrupted() *RunTestStage {
	time.Sleep(50 * time.Millisecond)
	require.NoError(s.t, syscall.Kill(syscall.Getpid(), syscall.SIGTERM))
	return s
}

func (s *RunTestStage) setup_teardown_is_called_within_50ms() *RunTestStage {
	err := retry.Do(func() error {
		if atomic.LoadInt32(s.setupTeardownCount) <= 0 {
			return errors.New("no teardown yet")
		}
		return nil
	}, retry.Delay(10*time.Millisecond), retry.Attempts(5))
	s.require.NoError(err)
	s.assert.GreaterOrEqual(atomic.LoadInt32(s.setupTeardownCount), int32(1))
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
	s.assert.False(s.metricData.Empty())
	return s
}

func (s *RunTestStage) a_scenario_where_the_final_iteration_takes_100ms() *RunTestStage {
	s.scenario = uuid.New().String()
	s.f1.Add(s.scenario, func(t *f1_testing.T) f1_testing.RunFn {
		t.Cleanup(func() {
			atomic.AddInt32(s.setupTeardownCount, 1)
		})
		s.runCount = 0
		return func(t *f1_testing.T) {
			t.Cleanup(func() {
				atomic.AddInt32(s.iterationTeardownCount, 1)
			})
			current := atomic.AddInt32(&s.runCount, 1)
			if current == 400 {
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
	err := retry.Do(func() error {
		metricNames := s.metricData.GetMetricNames()
		for _, mn := range metricNames {
			if mn == metricName {
				return nil
			}
		}
		return fmt.Errorf("%v did not contain %s", metricNames, metricName)
	})
	s.require.NoError(err)
	return s
}

func (s *RunTestStage) the_iteration_metric_has_n_results(n int, result string) *RunTestStage {
	err := retry.Do(func() error {
		metricFamily := s.metricData.GetMetricFamily("form3_loadtest_iteration")
		s.require.NotNil(metricFamily)
		resultMetric := getMetricByResult(metricFamily, result)
		s.require.NotNil(resultMetric)
		if uint64(n) == resultMetric.GetSummary().GetSampleCount() {
			return nil
		}
		return fmt.Errorf("expected %d to equal %d", uint64(n), resultMetric.GetSummary().GetSampleCount())
	})
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

func (s *RunTestStage) the_trace_output_should_be_present() *RunTestStage {
	s.require.NoError(s.tracerWriter.Close())
	s.tracerWg.Wait()

	s.require.NotEmpty(s.tracerOutput)
	return s
}

func (s *RunTestStage) tracing_is_enabled() *RunTestStage {
	s.tracerWg.Add(1)

	go func() {
		defer s.tracerWg.Done()
		_, err := io.Copy(&s.tracerOutput, s.tracerReader)
		s.require.NoError(err)
	}()

	s.tracer = trace.NewConsoleTracer(s.tracerWriter)
	return s
}

func (s *RunTestStage) a_concurrent_constant_trigger_is_configured() *RunTestStage {
	return s.
		a_trigger_type_of(Constant).and().
		a_rate_of("10/s").and().
		a_duration_of(500 * time.Millisecond).and().
		a_concurrency_of(50).and().
		an_iteration_limit_of(1000).and().
		a_scenario_where_each_iteration_takes(1 * time.Millisecond)
}

func (s *RunTestStage) a_fluentd_config_with_host_and_port(host, port string) *RunTestStage {
	s.t.Setenv(envsettings.EnvFluentdHost, host)
	s.t.Setenv(envsettings.EnvFluentdPort, port)
	s.settings = envsettings.Get()

	return s
}

func (s *RunTestStage) run_fails_with_error_containing(message string) *RunTestStage {
	s.require.ErrorContains(s.runError, message)

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
