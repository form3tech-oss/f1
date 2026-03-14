/*
Package f1testing provides the scenario execution context, analogous to Go's testing package.
It provides a T type which is injected into setup and iteration run functions, with common
functionality such as assertions and cleanup.

Both ScenarioFn and RunFn receive a context.Context as their first parameter. The context
is cancelled when the run is interrupted (SIGINT/SIGTERM), times out (--max-duration), or
reaches max iterations. Pass it to context-aware operations or check ctx.Done() to abort
long-running work when the run stops.
*/
package f1testing
