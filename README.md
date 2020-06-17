# f1
A load testing tool that allows you to run scenarios written in Golang for a variety of different profiles.

## Usage
The `pkg/f1` package can be imported into a `main.go` file to create an `f1` powered command line interface to run your load tests:

```golang
package main

import (
    "github.com/form3tech-oss/f1/pkg/f1"
)

function main() {
    f1.Execute()
}
``` 

This will give you a basic `f1` command line runner with the various running modes (see below) that are bundled with it. You can get more information about the various options available by simply running `go run main.go --help`

However, in order to use `f1` to run a load test scenario you've written you'll need to register it as follows:

```golang
package main

import (
    "github.com/form3tech-oss/f1/pkg/f1"
    "github.com/form3tech-oss/f1/pkg/f1/testing"
    "fmt"
)

function main() {
    testing.Add("mySuperFastLoadTest", setupMySuperFastLoadTest)
    f1.Execute()
}

function setupMySuperFastLoadTest(t *testing.T) (testing.RunFn, testing.TeardownFn) {
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

## Test lifecyle
`f1` assumes a lifecyle of tests which is implemented in the `pkg/loadtester` package which basically consists of:

1. Every test has a "setup" phase, which is executed once at the beginning of the test run.
2. Every test has one or more "iteration" runs, which are run as many times as is necessary based on your running mode.
3. Every test has a "teardown" phase, which is run once at the end of each test.

## Running modes
`f1` supports various running modes (available as subcommands to the `run` command). At the time of writing, these are:

* `constant` - applies load at a constant rate.
* `staged` - applies load at various stages (e.g. 1/s for 10s, then 2/s for 10s).
* `users` - applies load from a pool of virtual users.
* `gaussian` - applies load based on a Gaussian distribution.
