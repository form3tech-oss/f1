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

func TestWithSettingsEmptyDisablesAllSettings(t *testing.T) {
	t.Parallel()

	ts, count := newPushGatewayServer(t)
	_ = ts

	inst := newF1WithScenario("empty_settings", f1.WithSettings(f1.Settings{}))
	runConstant(t, inst, "empty_settings")

	require.Equal(t, int32(0), count.Load(),
		"WithSettings(Settings{}) should start from zero values; no push gateway")
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

func TestWithSettingsOverridesFinegrained(t *testing.T) {
	t.Parallel()

	ts, count := newPushGatewayServer(t)

	inst := newF1WithScenario("settings_last",
		f1.WithPrometheusPushGateway(ts.URL),
		f1.WithSettings(f1.Settings{}),
	)
	runConstant(t, inst, "settings_last")

	require.Equal(t, int32(0), count.Load(),
		"WithSettings placed after fine-grained options should replace them")
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

func TestWithLoggerTakesPrecedenceOverWithSettings(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := log.NewTestLogger(&buf)

	inst := newF1WithScenario("logger_over_settings",
		f1.WithSettings(f1.Settings{
			Logging: f1.LoggingSettings{
				Level:  slog.LevelError,
				Format: f1.LogFormatJSON,
			},
		}),
		f1.WithLogger(logger),
	)
	runConstant(t, inst, "logger_over_settings")

	output := buf.String()
	require.NotEmpty(t, output, "WithLogger should capture output")
	require.NotContains(t, output, `"level"`,
		"WithLogger text format should override Settings JSON format")
}

func TestWithSettingsAllFields(t *testing.T) {
	t.Parallel()

	ts, count := newPushGatewayServer(t)
	inst := newF1WithScenario("all_fields",
		f1.WithSettings(f1.Settings{
			Prometheus: f1.PrometheusSettings{
				PushGateway: ts.URL,
				Namespace:   "test-ns",
				LabelID:     "test-id",
			},
			Logging: f1.LoggingSettings{
				Level:  slog.LevelDebug,
				Format: f1.LogFormatJSON,
			},
		}),
	)
	runConstant(t, inst, "all_fields")

	require.Positive(t, count.Load(),
		"WithSettings with PushGateway should trigger metrics push")
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
