package f1testing

import "context"

// ScenarioFn initialises a scenario and returns the iteration function (RunFn) to be invoked for every iteration
// of the tests.
//
// ctx is cancelled when the run is interrupted (SIGINT/SIGTERM), times out (--max-duration), or reaches
// max iterations. Pass it to context-aware operations or check ctx.Done() to abort long-running setup.
type ScenarioFn func(ctx context.Context, t *T) RunFn

// RunFn performs a single iteration of the scenario. 't' may be used for asserting
// results or failing the scenario.
//
// ctx is cancelled when the run is stopped. Pass it to context-aware operations or check ctx.Done()
// to exit early when the user interrupts.
type RunFn func(ctx context.Context, t *T)
