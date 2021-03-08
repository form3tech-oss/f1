# f1
`f1` is a flexible load testing framework using the `go` language for test scenarios. This allows test scenarios to be developed as code, utilising full development principles such as test driven development. Test scenarios with multiple stages and multiple modes are ideally suited to this environment.

At Form3, many of our test scenarios using this framework combine REST API calls with asynchronous notifications from message queues. To achieve this, we need to have a worker pool listening to messages on the queue and distribute them to the appropriate instance of an active test run. We use this with thousands of concurrent test iterations in tests covering millions of iterations and running for multiple days.

## Usage
### Building an executable binary
The `pkg/f1` package can be imported into a `main.go` file to create an `f1`-powered command line interface to run your load tests:

```golang
package main

import (
    "github.com/form3tech-oss/f1/pkg/f1"
)

func main() {
    f1.Execute()
}
``` 

This will give you a basic `f1` command line runner with the various running modes (see below) that are bundled with it. You can get more information about the various options available by simply running `go run main.go --help`

### Writing load tests
Test scenarios consist of three stages: `Setup`, `Run` and `Teardown`. Setup is called once at the start of a test; this may be useful for generating resources needed for all tests, or subscribing to message queues. Run is called for every iteration of the test, often in parallel with multiple goroutines. Teardown is called once after all iterations complete. These Setup, Run and Teardown functions are defined as types in `f1`:

```golang
// SetupFn performs any setup required to run a scenario.
// It returns a RunFn to be invoked for every iteration of the tests
// and a TeardownFn to clear down any resources after all iterations complete
type SetupFn func(t *T) (RunFn, TeardownFn)

// RunFn performs a single iteration of the test. It my be used for asserting
// results or failing the test.
type RunFn func(t *T)

// TeardownFn clears down any resources from a test run after all iterations complete.
type TeardownFn RunFn
```

Writing tests is simply a case of implementing the types and registering them with `f1`:

```golang
package main

import (
    "github.com/form3tech-oss/f1/pkg/f1"
    "github.com/form3tech-oss/f1/pkg/f1/testing"
    "fmt"
)

func main() {
    testing.Add("mySuperFastLoadTest", setupMySuperFastLoadTest)
    f1.Execute()
}

func setupMySuperFastLoadTest(t *testing.T) (testing.RunFn, testing.TeardownFn) {
    runFn := func(t *testing.T) {
        fmt.Println("Wow, super fast!")
    }

    teardownFn := func(t *testing.T) {
        fmt.Println("Wow, that was fast!")
    }

    return runFn, teardownFn
}
```

`testing.Add()` registers a new scenario that can be run with `go run main.go run constant mySuperFastLoadTest` (where `constant` is the running mode). The `setupMySuperFastLoadTest` function performs any setup steps and returns a function to run on every "iteration" of the test and a function to run at the end of every test.

### Running load tests
Once you have written a load test and compiled a binary test runner, you can use the various ["trigger modes"](https://github.com/form3tech-oss/f1/tree/master/pkg/f1/trigger) that `f1` supports. These are available as subcommands to the `run` command, so try running `f1 run --help` for more information). The trigger modes currently implemented are as follows:

* `constant` - applies load at a constant rate (e.g. one request per second, irrespective of request duration).
* `staged` - applies load at various stages (e.g. one request per second for 10s, then two per second for 10s).
* `users` - applies load from a pool of users (e.g. requests from two users being sent sequentially - they are as fast or as slow as the requests themselves).
* `gaussian` - applies load based on a [Gaussian distribution](https://en.wikipedia.org/wiki/Normal_distribution) (e.g. varies load throughout a given duration with a mean and standard deviation).
* `ramp` - applies load constantly increasing or decreasing an initial load during a given ramp duration (e.g. from 0/s requests to 100/s requests during 10s).
* `file` - applies load based on a yaml config file - the file can contain any of the previous load modes (e.g. ["config-file-example.yaml"](config-file-example.yaml)).

## Design decisions
### Why did we decide to write our own load testing tool?
At Form3, we invest a lot of engineering time into load and performance testing of our platform. We initially used [`k6`](https://github.com/loadimpact/k6) to develop and run these tests, but this was problematic for us for a couple of reasons:

1. The tests that `k6` executes are written in Javascript - in order to test our platform, we often need to do things not easily done in Javascript (e.g. connect to SQS queues). The tests themselves can get quite complicated, and Javascript is not well suited to testing these sorts of tests.
2. `k6` only really supports a single model for applying load - users. This model assumes you have a finite pool of users, repeatedly making requests in sequence. This doesn't really work for us, since the payments industry has a pool of millions of users, each of whom could make a payment at any moment - when they do, they don't wait around for the previous customer to finish!

### Enter `f1`
We started working on `f1`, because we already had a suite of load test scenarios that we had started writing in Go. `k6` interfaced with these by making web requests to a server that actually ran the tests - a bit of a hack.

We wanted to be able to write the tests in Go in a native load testing framework, which also supported our use case of applying load more aggressively (without waiting for requests to finish).

`f1` is the result. It supports writing load test scenarios natively in Go, which means you can make your tests as complicated as you like and test them well. It also has a variety of "trigger modes" (see above), which allow load to be applied in the same way as `k6`, but also in other, more aggressive modes. Writing new trigger modes is easy, so we welcome contributations to expand the list.

## Contributions
If you'd like to help improve `f1`, please fork this repo and raise a PR!
