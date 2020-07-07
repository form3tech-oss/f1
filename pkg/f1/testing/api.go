package testing

// SetupFn performs any setup required to run a scenario.
// It returns a RunFn to be invoked for every iteration of the tests
// and a TeardownFn to clear down any resources after all iterations complete
type SetupFn func(t *T) (RunFn, TeardownFn)

// RunFn performs a single iteration of the test. It my be used for asserting
// results or failing the test.
type RunFn func(t *T)

// TeardownFn clears down any resources from a test run after all iterations complete.
type TeardownFn RunFn

type MultiStageSetupFn func(t *T) ([]Stage, TeardownFn)

type Stage struct {
	Name  string
	RunFn RunFn
}
