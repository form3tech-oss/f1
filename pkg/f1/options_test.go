package f1_test

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/form3tech-oss/f1/v3/internal/log"
	"github.com/form3tech-oss/f1/v3/pkg/f1"
	"github.com/form3tech-oss/f1/v3/pkg/f1/f1testing"
)

func newPushGatewayServer(t *testing.T) (*httptest.Server, *atomic.Int32) {
	t.Helper()

	var count atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		count.Add(1)
	}))
	t.Cleanup(ts.Close)

	return ts, &count
}

func newF1WithScenario(name string, opts ...f1.Option) *f1.F1 {
	inst := f1.New(opts...)
	inst.AddScenario(name, func(_ context.Context, _ *f1testing.T) f1testing.RunFn {
		return func(_ context.Context, _ *f1testing.T) {}
	})

	return inst
}

func runConstant(t *testing.T, inst *f1.F1, scenario string) {
	t.Helper()

	err := inst.Run(context.Background(), []string{
		"run", "constant", scenario,
		"--rate", "1/1s",
		"--max-duration", "1s",
		"--max-iterations", "1",
	})
	require.NoError(t, err)
}

func TestEnvVarsUsedByDefault(t *testing.T) {
	ts, count := newPushGatewayServer(t)
	t.Setenv("PROMETHEUS_PUSH_GATEWAY", ts.URL)

	inst := newF1WithScenario("env_default")
	runConstant(t, inst, "env_default")

	require.Positive(t, count.Load(),
		"PROMETHEUS_PUSH_GATEWAY env var should trigger metrics push")
}

func TestWithSettingsReplacesPushGateway(t *testing.T) {
	t.Parallel()

	ts, count := newPushGatewayServer(t)
	inst := newF1WithScenario("settings_push",
		f1.WithSettings(f1.Settings{
			Prometheus: f1.PrometheusSettings{PushGateway: ts.URL},
		}),
	)
	runConstant(t, inst, "settings_push")

	require.Positive(t, count.Load(),
		"WithSettings should configure push gateway without env vars")
}

func TestWithSettingsEmptyIgnoresEnvVars(t *testing.T) {
	ts, count := newPushGatewayServer(t)
	t.Setenv("PROMETHEUS_PUSH_GATEWAY", ts.URL)

	inst := newF1WithScenario("no_env", f1.WithSettings(f1.Settings{}))
	runConstant(t, inst, "no_env")

	require.Equal(t, int32(0), count.Load(),
		"WithSettings(Settings{}) should ignore env vars")
}

func TestWithPrometheusPushGatewayOverridesEnv(t *testing.T) {
	ts, count := newPushGatewayServer(t)
	t.Setenv("PROMETHEUS_PUSH_GATEWAY", "http://env-should-not-be-used.invalid")

	inst := newF1WithScenario("override", f1.WithPrometheusPushGateway(ts.URL))
	runConstant(t, inst, "override")

	require.Positive(t, count.Load(),
		"WithPrometheusPushGateway should override env var")
}

func TestFineGrainedOverridesAfterWithSettings(t *testing.T) {
	t.Parallel()

	ts, count := newPushGatewayServer(t)
	inst := newF1WithScenario("fine_grained",
		f1.WithSettings(f1.Settings{}),
		f1.WithPrometheusPushGateway(ts.URL),
	)
	runConstant(t, inst, "fine_grained")

	require.Positive(t, count.Load(),
		"fine-grained options should apply after WithSettings")
}

func TestWithLogLevelAndFormat(t *testing.T) {
	t.Parallel()

	inst := newF1WithScenario("log_opts",
		f1.WithSettings(f1.Settings{}),
		f1.WithLogLevel(slog.LevelDebug),
		f1.WithLogFormat(f1.LogFormatJSON),
	)
	runConstant(t, inst, "log_opts")
}

func TestWithLoggerTakesPrecedenceOverLogOptions(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := log.NewTestLogger(&buf)

	inst := newF1WithScenario("logger_precedence",
		f1.WithLogger(logger),
		f1.WithLogLevel(slog.LevelError),
		f1.WithLogFormat(f1.LogFormatJSON),
	)
	runConstant(t, inst, "logger_precedence")

	output := buf.String()
	require.NotEmpty(t, output, "WithLogger's logger should capture output")
	require.NotContains(t, output, `"level"`,
		"explicit logger format (text) should be used, not JSON from WithLogFormat")
}

func TestDefaultSettingsLoadsEnv(t *testing.T) {
	t.Setenv("PROMETHEUS_PUSH_GATEWAY", "http://test:9091")
	t.Setenv("PROMETHEUS_NAMESPACE", "test-ns")
	t.Setenv("PROMETHEUS_LABEL_ID", "test-id")
	t.Setenv("LOG_FILE_PATH", "/tmp/test.log")
	t.Setenv("F1_LOG_LEVEL", "debug")
	t.Setenv("F1_LOG_FORMAT", "json")

	s := f1.DefaultSettings()
	require.Equal(t, "http://test:9091", s.Prometheus.PushGateway)
	require.Equal(t, "test-ns", s.Prometheus.Namespace)
	require.Equal(t, "test-id", s.Prometheus.LabelID)
	require.Equal(t, "/tmp/test.log", s.Logging.FilePath)
	require.Equal(t, slog.LevelDebug, s.Logging.Level)
	require.Equal(t, f1.LogFormatJSON, s.Logging.Format)
}

func TestDefaultSettingsReturnsDefaults(t *testing.T) {
	t.Setenv("PROMETHEUS_PUSH_GATEWAY", "")
	t.Setenv("PROMETHEUS_NAMESPACE", "")
	t.Setenv("PROMETHEUS_LABEL_ID", "")
	t.Setenv("LOG_FILE_PATH", "")
	t.Setenv("F1_LOG_LEVEL", "")
	t.Setenv("F1_LOG_FORMAT", "")

	s := f1.DefaultSettings()
	require.Empty(t, s.Prometheus.PushGateway)
	require.Equal(t, slog.LevelInfo, s.Logging.Level)
	require.Equal(t, f1.LogFormatText, s.Logging.Format)
}

func TestParseLogLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"trace", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"INFO", slog.LevelInfo},
		{"", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"error", slog.LevelError},
		{"fatal", slog.LevelError},
		{"panic", slog.LevelError},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			got, err := f1.ParseLogLevel(tt.input)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestParseLogLevelInvalid(t *testing.T) {
	t.Parallel()

	_, err := f1.ParseLogLevel("invalid")
	require.Error(t, err)
	require.ErrorContains(t, err, "unknown log level")
}

func TestParseLogFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  f1.LogFormat
	}{
		{"text", f1.LogFormatText},
		{"TEXT", f1.LogFormatText},
		{"", f1.LogFormatText},
		{"json", f1.LogFormatJSON},
		{"JSON", f1.LogFormatJSON},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			got, err := f1.ParseLogFormat(tt.input)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestParseLogFormatInvalid(t *testing.T) {
	t.Parallel()

	_, err := f1.ParseLogFormat("yaml")
	require.Error(t, err)
	require.ErrorContains(t, err, "unknown log format")
}

func TestLogFormatString(t *testing.T) {
	t.Parallel()

	require.Equal(t, "text", f1.LogFormatText.String())
	require.Equal(t, "json", f1.LogFormatJSON.String())
}
