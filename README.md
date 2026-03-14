<a href="https://pkg.go.dev/github.com/form3tech-oss/f1/v3/pkg/f1"><img align="right" src="https://pkg.go.dev/badge/github.com/form3tech-oss/f1/v3/pkg/f1.svg" alt="Go Reference"></a>
# f1
`f1` is a flexible load testing framework using the `go` language for test scenarios. This allows test scenarios to be developed as code, utilising full development principles such as test driven development. Test scenarios with multiple stages and multiple modes are ideally suited to this environment.

At Form3, many of our test scenarios using this framework combine REST API calls with asynchronous notifications from message queues. To achieve this, we need to have a worker pool listening to messages on the queue and distribute them to the appropriate instance of an active test run. We use this with thousands of concurrent test iterations in tests covering millions of iterations and running for multiple days.

## Usage
### Writing load tests
Test scenarios consist of two stages: 
* Setup - represented by `ScenarioFn` which is called once at the start of a test; this may be useful for generating resources needed for all tests, or subscribing to message queues.
* Run - represented by `RunFn` which is called for every iteration of the test, often in parallel with multiple goroutines.

Cleanup functions can be provided for both stages: `Setup` and `Run` which will be executed in LIFO order.
These `ScenarioFn` and `RunFn` functions are defined as types in `f1`:

```golang
// ScenarioFn initialises a scenario and returns the iteration function (RunFn) to be invoked for every iteration
// of the tests. ctx is cancelled when the run is interrupted or times out.
type ScenarioFn func(ctx context.Context, t *T) RunFn

// RunFn performs a single iteration of the scenario. 't' may be used for asserting
// results or failing the scenario. ctx is cancelled when the run is stopped; check ctx.Done() for cancellation.
type RunFn func(ctx context.Context, t *T)
```

Writing tests is simply a case of implementing the types and registering them with `f1`:

```golang
package main

import (
	"context"
	"fmt"

	"github.com/form3tech-oss/f1/v3/pkg/f1"
	"github.com/form3tech-oss/f1/v3/pkg/f1/f1testing"
)

func main() {
	// Create a new f1 instance, add all the scenarios and execute the f1 tool.
	// Any scenario that is added here can be executed like: `go run main.go run constant mySuperFastLoadTest`
	f1.New().AddScenario("mySuperFastLoadTest", setupMySuperFastLoadTest).Execute()
}

// Performs any setup steps and returns a function to run on every iteration of the scenario
func setupMySuperFastLoadTest(ctx context.Context, t *f1testing.T) f1testing.RunFn {
	fmt.Println("Setup the scenario")

	// Register clean up function which will be invoked at the end of the scenario execution to clean up the setup
	t.Cleanup(func() {
		fmt.Println("Clean up the setup of the scenario")
	})

	runFn := func(ctx context.Context, t *f1testing.T) {
		fmt.Println("Run the test")

		// Register clean up function for each test which will be invoked in LIFO order after each iteration
		t.Cleanup(func() {
			fmt.Println("Clean up the test execution")
		})
	}

	return runFn
}
```

### Running load tests
Once you have written a load test and compiled a binary test runner, you can use the various ["trigger modes"](https://github.com/form3tech-oss/f1/tree/master/internal/trigger) that `f1` supports. These are available as subcommands to the `run` command, so try running `f1 run --help` for more information). The trigger modes currently implemented are as follows:

* `constant` - applies load at a constant rate (e.g. one request per second, irrespective of request duration).
* `staged` - applies load at various stages (e.g. one request per second for 10s, then two per second for 10s).
* `users` - applies load from a pool of users (e.g. requests from two users being sent sequentially - they are as fast or as slow as the requests themselves).
* `gaussian` - applies load based on a [Gaussian distribution](https://en.wikipedia.org/wiki/Normal_distribution) (e.g. varies load throughout a given duration with a mean and standard deviation).
* `ramp` - applies load constantly increasing or decreasing an initial load during a given ramp duration (e.g. from 0/s requests to 100/s requests during 10s).
* `file` - applies load based on a yaml config file - the file can contain any of the previous load modes (e.g. ["config-file-example.yaml"](config-file-example.yaml)).

#### Output description

Currently, output from running f1 load tests looks like that:
```
[   1s]  ✔    20  ✘     0 (20/s)   avg: 72ns, min: 125ns, max: 27.590042ms
```

It provides the following information:
- `[   1s]` how long the test has been running for,
- `✔    20` number of successful iterations,
- `✘     0` number of failed iterations,
- `(20/s)` (attempted) rate,
- `avg: 72ns, min: 125ns, max: 27.590042ms` average, min and max iteration times.

### Configuration

f1 can be configured via environment variables, programmatic options, or both. By default, environment variables are read at construction time. Programmatic options override env vars for the fields they set.

#### Settings reference

| Setting | Environment variable | Programmatic option | Default |
| --- | --- | --- | --- |
| Prometheus push gateway | `PROMETHEUS_PUSH_GATEWAY` | `f1.WithPrometheusPushGateway(url)` | disabled |
| Prometheus namespace label | `PROMETHEUS_NAMESPACE` | `f1.WithPrometheusNamespace(ns)` | `""` |
| Prometheus ID label | `PROMETHEUS_LABEL_ID` | `f1.WithPrometheusLabelID(id)` | `""` |
| Log file path | `LOG_FILE_PATH` | `f1.WithLogFilePath(path)` | auto temp file |
| Log level | `F1_LOG_LEVEL` | `f1.WithLogLevel(slog.LevelDebug)` | `slog.LevelInfo` |
| Log format | `F1_LOG_FORMAT` | `f1.WithLogFormat(f1.LogFormatJSON)` | `f1.LogFormatText` |

Log level and format options use Go's standard `slog.Level` and f1's `LogFormat` type for compile-time safety. Use `f1.ParseLogLevel(string)` and `f1.ParseLogFormat(string)` to convert from strings (e.g. from config files).

#### Configuring without environment variables

Use `f1.WithSettings(f1.Settings{})` to start from zero values, ignoring all environment variables. Fine-grained options (`WithLogLevel`, `WithPrometheusPushGateway`, etc.) still apply after the baseline:

```golang
f1.New(
    f1.WithSettings(f1.Settings{}),
    f1.WithLogLevel(slog.LevelWarn),
    f1.WithLogFormat(f1.LogFormatJSON),
).AddScenario("myScenario", mySetup).Execute()
```

For full control, pass a complete `f1.Settings` struct:

```golang
f1.New(
    f1.WithSettings(f1.Settings{
        Prometheus: f1.PrometheusSettings{
            PushGateway: "http://pushgateway:9091",
            Namespace:   "my-namespace",
        },
        Logging: f1.LoggingSettings{
            Level:  slog.LevelDebug,
            Format: f1.LogFormatJSON,
        },
    }),
).AddScenario("myScenario", mySetup).Execute()
```

#### Precedence

Settings are resolved in this order (highest priority first):

1. **Programmatic options** — values passed to `f1.New()` (applied in order)
2. **Environment variables** — read at construction time (baseline when no `WithSettings` is used)
3. **Defaults** — `slog.LevelInfo`, `LogFormatText`, no Prometheus push

When `f1.WithLogger(logger)` is used, the caller owns the logger entirely. `WithLogLevel`, `WithLogFormat`, and the corresponding env vars have no effect:

```golang
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
f1.New(
    f1.WithLogger(logger),
).AddScenario("myScenario", mySetup).Execute()
```

#### Default env-backed behaviour

When no `WithSettings` is provided, environment variables are used as the baseline (backward-compatible with previous releases):

```golang
// Env vars like PROMETHEUS_PUSH_GATEWAY are read automatically
f1.New().AddScenario("myScenario", mySetup).Execute()

// Fine-grained options override individual env var values
f1.New(
    f1.WithPrometheusPushGateway("http://pushgateway:9091"),
    f1.WithLogLevel(slog.LevelDebug),
).AddScenario("myScenario", mySetup).Execute()
```

## Contributions
If you'd like to help improve `f1`, please fork this repo and raise a PR!
