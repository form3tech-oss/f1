<a href="https://pkg.go.dev/github.com/form3tech-oss/f1/v2/pkg/f1"><img align="right" src="https://pkg.go.dev/badge/github.com/form3tech-oss/f1/v2/pkg/f1.svg" alt="Go Reference"></a>
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
// of the tests.
type ScenarioFn func(t *T) RunFn

// RunFn performs a single iteration of the scenario. 't' may be used for asserting
// results or failing the scenario.
type RunFn func(t *T)
```

Writing tests is simply a case of implementing the types and registering them with `f1`:

```golang
package main

import (
	"fmt"

	"github.com/form3tech-oss/f1/v2/pkg/f1"
	"github.com/form3tech-oss/f1/v2/pkg/f1/testing"
)

func main() {
	// Create a new f1 instance, add all the scenarios and execute the f1 tool.
	// Any scenario that is added here can be executed like: `go run main.go run constant mySuperFastLoadTest`
	f1.New().Add("mySuperFastLoadTest", setupMySuperFastLoadTest).Execute()
}

// Performs any setup steps and returns a function to run on every iteration of the scenario
func setupMySuperFastLoadTest(t *testing.T) testing.RunFn {
	fmt.Println("Setup the scenario")
	
	// Register clean up function which will be invoked at the end of the scenario execution to clean up the setup
	t.Cleanup(func() {
		fmt.Println("Clean up the setup of the scenario")
	})
	
	runFn := func(t *testing.T) {
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

### Environment variables

| Name | Format | Default | Description |
| --- | --- | --- | --- |
| `PROMETHEUS_PUSH_GATEWAY` | string - `host:port` or `ip:port` | `""` | Configures the address of a [Prometheus Push Gateway](https://prometheus.io/docs/instrumenting/pushing/) for exposing metrics. The prometheus job name configured will be `f1-{scenario_name}`. Disabled by default.|
| `PROMETHEUS_NAMESPACE` | string | `""` | Sets the metric label `namespace` to the specified value. Label is omitted if the value provided is empty.|
| `PROMETHEUS_LABEL_ID` | string | `""` | Sets the metric label `id` to the specified value. Label is omitted if the value provided is empty.|
| `LOG_FILE_PATH` | string | `""`| Specify the log file path used if `--verbose` is disabled. The logfile path will be an automatically generated temp file if not specified. |
| `LOG_LEVEL` | string | `"info"`| Specify the log level of the default logger, one of: `debug`, `warn`, `error`  |
| `LOG_FORMAT` | string | `""`| Specify the log format of the default logger, defaults to `text` formatter, allows `json`  |

## Contributions
If you'd like to help improve `f1`, please fork this repo and raise a PR!
