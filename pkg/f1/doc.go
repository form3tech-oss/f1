/*
F1 is a flexible load testing framework that provides a CLI which can be used to inject load,
as well as a set of Go packages that can be used to write load test scenarios.

F1 can be used to write simple load test scenarios which, for example, make an HTTP request.
However, it can also be used to write more complex scenarios which might trigger a synchronous
action and then await asynchronous feedback (e.g. making an HTTP request, and then waiting for
a message to arrive via a message broker).


Writing load tests


Test scenarios consist of two stages:

1. The test setup which is called once at the start of a test. This may be useful for generating resources needed for all tests, or subscribing to message queues.

2. A single iteration's run function. This is called for every iteration of the test, often in parallel with other iterations.

Cleanup functions can also be provided for both stages, and are executed in LIFO order.

Types are provided for setup and iteration/run functions as below:

	// ScenarioFn initialises a scenario and returns the iteration function (RunFn) to be invoked for every iteration
	// of the tests.
	type ScenarioFn func(t *T) RunFn

	// RunFn performs a single iteration of the scenario. 't' may be used for asserting
	// results or failing the scenario.
	type RunFn func(t *T)

Writing tests is simply a case of implementing the types and registering them with F1:

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


Running load tests


Once you have written a load test and compiled a binary test runner, you can use the various "trigger modes" (See here for more details: https://github.com/form3tech-oss/f1/tree/master/internal/trigger) that F1 supports. These are available as subcommands to the "f1 run" command, so trying "f1 run --help" will provide more information. The trigger modes currently implemented are as follows:

- constant: applies load at a constant rate (e.g. one request per second, irrespective of request duration).

- staged: applies load at various stages (e.g. one request per second for 10s, then two per second for 10s).

- users: applies load from a pool of users (e.g. requests from two users being sent sequentially - they are as fast or as slow as the requests themselves).

- gaussian: applies load based on a Gaussian distribution (e.g. varies load throughout a given duration with a mean and standard deviation).

- ramp: applies load constantly increasing or decreasing an initial load during a given ramp duration (e.g. from 0/s requests to 100/s requests during 10s).

- file: applies load based on a yaml config file - the file can contain any of the previous load modes.

To make use of the F1 CLI, follow the usage example above and then run:

	go run main.go --help

You can of course also compile an "F1" binary, as follows:

	go build -o f1 main.go
	./f1 --help
	./f1 run constant -r 1/s -d 10s mySuperFastLoadTest

*/
package f1
