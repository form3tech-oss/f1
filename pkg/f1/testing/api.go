package testing

// ScenarioFn initialises a scenario and returns the iteration function (RunFn) to be invoked for every iteration
// of the tests.
type ScenarioFn func(t *T) RunFn

// RunFn performs a single iteration of the scenario. 't' may be used for asserting
// results or failing the scenario.
type RunFn func(t *T)
