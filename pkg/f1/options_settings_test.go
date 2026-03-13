package f1_test

import (
	"bytes"
	"context"
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

func TestWithPrometheusPushGatewayOverridesEnv(t *testing.T) {
	ts, count := newPushGatewayServer(t)
	t.Setenv("PROMETHEUS_PUSH_GATEWAY", "http://env-should-not-be-used.invalid")

	inst := newF1WithScenario("override", f1.WithPrometheusPushGateway(ts.URL))
	runConstant(t, inst, "override")

	require.Positive(t, count.Load(),
		"programmatic WithPrometheusPushGateway should override env var")
}

func TestWithLogLevelAndFormat(t *testing.T) {
	t.Setenv("PROMETHEUS_PUSH_GATEWAY", "")

	inst := newF1WithScenario("log_opts",
		f1.WithLogLevel("debug"),
		f1.WithLogFormat("json"),
	)
	runConstant(t, inst, "log_opts")
}

func TestWithLoggerTakesPrecedenceOverLogOptions(t *testing.T) {
	var buf bytes.Buffer
	logger := log.NewTestLogger(&buf)

	inst := newF1WithScenario("logger_precedence",
		f1.WithLogger(logger),
		f1.WithLogLevel("error"),
		f1.WithLogFormat("json"),
	)
	runConstant(t, inst, "logger_precedence")

	output := buf.String()
	require.NotEmpty(t, output, "WithLogger's logger should capture output")
	require.NotContains(t, output, `"level"`,
		"explicit logger format (text) should be used, not JSON from WithLogFormat")
}

func TestWithoutEnvSettingsIgnoresEnvVars(t *testing.T) {
	ts, count := newPushGatewayServer(t)
	t.Setenv("PROMETHEUS_PUSH_GATEWAY", ts.URL)

	inst := newF1WithScenario("no_env", f1.WithoutEnvSettings())
	runConstant(t, inst, "no_env")

	require.Equal(t, int32(0), count.Load(),
		"WithoutEnvSettings should prevent env var PROMETHEUS_PUSH_GATEWAY from being used")
}
