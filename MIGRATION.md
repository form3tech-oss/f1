# Migration Guide: F1 v2 → v3

This guide documents all breaking changes in F1 v3 and how to migrate your code. The v3 release modernises the API, adds context support, and removes deprecated features.

---

## Table of Contents

1. [Overview](#1-overview)
2. [Module Path and go.mod](#2-module-path-and-gomod)
3. [Import Path Changes](#3-import-path-changes)
4. [Entry Points: Execute and Run](#4-entry-points-execute-and-run)
5. [F1 Construction and Options](#5-f1-construction-and-options)
6. [Scenario Registration](#6-scenario-registration)
7. [Scenario Function Signatures](#7-scenario-function-signatures)
8. [Testing Package Rename](#8-testing-package-rename)
9. [T Type Changes](#9-t-type-changes)
10. [Metrics Package Removal](#10-metrics-package-removal)
11. [CLI and Flag Changes](#11-cli-and-flag-changes)
12. [Removed Features](#12-removed-features)
13. [Complete Before/After Example](#13-complete-beforeafter-example)
14. [Migration Checklist](#14-migration-checklist)

---

## 1. Overview

| Area | v2 | v3 |
|------|-----|-----|
| Module | `github.com/form3tech-oss/f1/v2` | `github.com/form3tech-oss/f1/v3` |
| Testing package | `pkg/f1/testing` | `pkg/f1/f1testing` |
| Entry point (library) | `ExecuteWithArgs(args)` | `Run(ctx, args)` |
| F1 construction | `New().WithLogger(l)` | `New(WithLogger(l))` |
| Scenario registration | `Add(name, fn)` | `AddScenario(name, fn)` |
| Scenario options | `Description(d)`, `Parameter(p)` | `WithDescription(d)`, `WithParameter(p)` |
| Scenario/Run signatures | `func(t *T)` | `func(ctx context.Context, t *T)` |
| T.Error / T.Fatal | `Error(err error)`, `Fatal(err error)` | `Error(args ...any)`, `Fatal(args ...any)` |
| T.Logger | `*logrus.Logger` | `*slog.Logger` |
| Metrics | `metrics.GetMetrics()` | Removed; use `WithStaticMetrics` |

---

## 2. Module Path and go.mod

**Change**: Update the module path from `v2` to `v3`.

```diff
# go.mod
- module github.com/form3tech-oss/f1/v2
+ module github.com/form3tech-oss/f1/v3
```

Update your dependency:

```bash
go get github.com/form3tech-oss/f1/v3@latest
```

Or in `go.mod`:

```diff
- github.com/form3tech-oss/f1/v2 v2.x.x
+ github.com/form3tech-oss/f1/v3 v3.x.x
```

---

## 3. Import Path Changes

**Change**: All imports must use the `v3` path and the renamed `f1testing` package.

```diff
import (
-	"github.com/form3tech-oss/f1/v2/pkg/f1"
-	"github.com/form3tech-oss/f1/v2/pkg/f1/testing"
+	"github.com/form3tech-oss/f1/v3/pkg/f1"
+	"github.com/form3tech-oss/f1/v3/pkg/f1/f1testing"
)
```

**Package rename**: `pkg/f1/testing` → `pkg/f1/f1testing`. The name `testing` clashed with Go's standard `testing` package; `f1testing` avoids this and makes the origin explicit.

---

## 4. Entry Points: Execute and Run

### Execute (unchanged)

`Execute()` remains the same for typical CLI usage from `main()`:

```go
// v2 and v3 — no change
f1.New().AddScenario("myTest", myScenario).Execute()
```

### ExecuteWithArgs → Run

**Change**: `ExecuteWithArgs(args)` is removed. Use `Run(ctx, args)` instead. `Run` returns an `error` and never exits; it accepts `context.Context` for cancellation and timeouts.

| v2 | v3 |
|----|-----|
| `f.ExecuteWithArgs(args)` | `f.Run(context.Background(), args)` |
| `err := f.ExecuteWithArgs(args)` | `err := f.Run(ctx, args)` |

```diff
// v2
- err := f.ExecuteWithArgs([]string{"run", "constant", "-r", "1/s", "-d", "10s", "myTest"})

// v3
+ err := f.Run(context.Background(), []string{"run", "constant", "-r", "1/s", "-d", "10s", "myTest"})
```

**With cancellation** (e.g. in tests):

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
err := f.Run(ctx, []string{"run", "constant", "myTest"})
```

**Pass `nil` for args** to use `os.Args` (e.g. when called from `main`):

```go
f.Run(context.Background(), nil)  // equivalent to Execute() but returns error
```

---

## 5. F1 Construction and Options

**Change**: `New()` now accepts functional options. Fluent methods `WithLogger` and `WithStaticMetrics` are removed in favour of constructor options.

| v2 | v3 |
|----|-----|
| `New()` | `New()` — unchanged |
| `New().WithLogger(logger)` | `New(WithLogger(logger))` |
| `New().WithStaticMetrics(labels)` | `New(WithStaticMetrics(labels))` |

```diff
// v2
- f := f1.New().WithLogger(myLogger).WithStaticMetrics(map[string]string{"env": "prod"})

// v3
+ f := f1.New(
+ 	f1.WithLogger(myLogger),
+ 	f1.WithStaticMetrics(map[string]string{"env": "prod"}),
+ )
```

---

## 6. Scenario Registration

### Add → AddScenario

**Change**: `Add` is renamed to `AddScenario` for clarity.

```diff
// v2
- f.Add("myTest", myScenario)

// v3
+ f.AddScenario("myTest", myScenario)
```

### Scenario Options: Description and Parameter

**Change**: `Description(d)` and `Parameter(p)` are renamed to `WithDescription(d)` and `WithParameter(p)` for consistency with the functional options pattern.

```diff
// v2
- f.Add("myTest", myScenario,
- 	scenarios.Description("Load test for API X"),
- 	scenarios.Parameter(scenarios.ScenarioParameter{Name: "rate", Default: "1/s"}),
- )

// v3
+ f.AddScenario("myTest", myScenario,
+ 	scenarios.WithDescription("Load test for API X"),
+ 	scenarios.WithParameter(scenarios.ScenarioParameter{Name: "rate", Default: "1/s"}),
+ )
```

---

## 7. Scenario Function Signatures

**Change**: `ScenarioFn` and `RunFn` now receive `context.Context` as the first parameter. The context is cancelled when the run is interrupted (SIGINT/SIGTERM), times out (`--max-duration`), or reaches max iterations.

### ScenarioFn

| v2 | v3 |
|----|-----|
| `func(t *T) RunFn` | `func(ctx context.Context, t *T) RunFn` |

```diff
// v2
- func myScenario(t *f1testing.T) f1testing.RunFn {
- 	runFn := func(t *f1testing.T) { ... }
+ func myScenario(ctx context.Context, t *f1testing.T) f1testing.RunFn {
+ 	runFn := func(ctx context.Context, t *f1testing.T) { ... }
  	return runFn
  }
```

### RunFn

| v2 | v3 |
|----|-----|
| `func(t *T)` | `func(ctx context.Context, t *T)` |

**Using context** for cancellation or timeouts:

```go
func myScenario(ctx context.Context, t *f1testing.T) f1testing.RunFn {
	return func(ctx context.Context, t *f1testing.T) {
		req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
		resp, err := http.DefaultClient.Do(req)  // respects ctx cancellation
		// ...
	}
}
```

---

## 8. Testing Package Rename

**Change**: The package `pkg/f1/testing` is renamed to `pkg/f1/f1testing`. Update all references.

```diff
- import "github.com/form3tech-oss/f1/v2/pkg/f1/testing"
+ import "github.com/form3tech-oss/f1/v3/pkg/f1/f1testing"

- func myScenario(t *testing.T) testing.RunFn {
+ func myScenario(ctx context.Context, t *f1testing.T) f1testing.RunFn {
```

Type references:

| v2 | v3 |
|----|-----|
| `*testing.T` | `*f1testing.T` |
| `testing.ScenarioFn` | `f1testing.ScenarioFn` |
| `testing.RunFn` | `f1testing.RunFn` |

---

## 9. T Type Changes

### 9.1 Error and Fatal Signatures

**Change**: `Error` and `Fatal` now accept `args ...any` (matching `testing.T`), instead of `err error`. This enables sharing test helpers between `go test` and f1 scenarios.

| v2 | v3 |
|----|-----|
| `Error(err error)` | `Error(args ...any)` |
| `Fatal(err error)` | `Fatal(args ...any)` |

**Backward compatibility**: Existing `Error(err)` and `Fatal(err)` calls still work — a single `error` is passed as the first argument and formatted with `fmt.Sprintln`.

```go
// All of these work in v3
t.Error(err)
t.Error("failed:", err)
t.Fatal(err)
t.Fatalf("iteration %s failed: %v", t.Iteration, err)
```

### 9.2 Log Levels

**Change**: `Error`, `Errorf`, `Fatal`, and `Fatalf` now log at ERROR level. `Log` and `Logf` log at INFO level. In v2, Error/Fatal delegated to Log (INFO); v3 aligns with `testing.T` semantics.

### 9.3 Logger() Return Type

**Change**: `T.Logger()` now returns `*slog.Logger` instead of `*logrus.Logger`. Logrus is removed as a dependency.

```diff
// v2 — Logger() returned *logrus.Logger
- logger := t.Logger()
- logger.WithField("key", "value").Info("msg")

// v3 — Logger() returns *slog.Logger
+ logger := t.Logger()
+ logger.With("key", "value").Info("msg")
```

### 9.4 T.Time() Removed

**Change**: `T.Time(stageName string, f func())` is removed. Internal metrics are no longer exposed via the testing package. If you need timing, record it yourself:

```go
// v2
- t.Time("http_request", func() { doRequest() })

// v3 — record timing yourself if needed
+ start := time.Now()
+ doRequest()
+ duration := time.Since(start)
+ // use duration as needed (e.g. custom metrics, logging)
```

### 9.5 NewT() Removed

**Change**: `NewT(iter, scenarioName string)` is removed. Use `NewTWithOptions` only. The framework creates `T` instances internally; you typically only need `NewTWithOptions` for tests.

```diff
// v2
- t, teardown := testing.NewT("1", "myScenario")

// v3
+ t, teardown := f1testing.NewTWithOptions("myScenario", f1testing.WithIteration("1"))
```

### 9.6 WithLogrusLogger Removed

**Change**: `WithLogrusLogger(logrusLogger *logrus.Logger)` is removed. Use `WithLogger(*slog.Logger)` when constructing `T` via `NewTWithOptions`.

```diff
// v2
- t, teardown := testing.NewTWithOptions("myScenario",
- 	testing.WithLogrusLogger(logrusLogger),
- )

// v3
+ t, teardown := f1testing.NewTWithOptions("myScenario",
+ 	f1testing.WithLogger(slogLogger),
+ )
```

---

## 10. Metrics Package Removal

**Change**: `pkg/f1/metrics.GetMetrics()` is removed. Internal metrics are no longer exposed. Use `WithStaticMetrics` on the F1 instance for custom labels.

```diff
// v2 — direct access to internal metrics (deprecated)
- m := metrics.GetMetrics()
- m.RecordIterationStage(...)

// v3 — no replacement for GetMetrics
+ // Use f1.New(WithStaticMetrics(map[string]string{"env": "prod"})) for labels
+ // Internal metrics (iteration counts, latency, etc.) are not exposed
```

If you used `GetMetrics()` for custom labels, migrate to `WithStaticMetrics`:

```go
f1.New(WithStaticMetrics(map[string]string{
	"environment": "staging",
	"service":     "my-api",
}))
```

---

## 11. CLI and Flag Changes

### 11.1 Renamed Flags

| v2 | v3 |
|----|-----|
| `--cpuprofile` | `--cpu-profile` |
| `--memprofile` | `--mem-profile` |
| `--iterationFrequency` (staged, gaussian) | `--iteration-frequency` |

### 11.2 Flag Grouping

Run command flags are now grouped in help output (Output, Duration & limits, Concurrency, Failure handling, Shutdown, Trigger options). Behaviour is unchanged.

### 11.3 Removed Flags

| Flag | Action |
|------|--------|
| `--verbose-fail` | Removed (was deprecated) |

---

## 12. Removed Features

### 12.1 Chart Command

**Change**: The `chart` subcommand (`f1 chart ...`) is removed. The go-chart and asciigraph dependencies are removed. Use external tools (e.g. Grafana, Prometheus) for visualisation.

### 12.2 Fluentd Integration

**Change**: Fluentd environment variables (`FluentdHost`, `FluentdPort`) and related code are removed. Use structured logs (slog) and your log aggregation pipeline instead.

### 12.3 Logrus

**Change**: Logrus is removed. All logging uses `log/slog`. Migrate `WithLogrusLogger` to `WithLogger(*slog.Logger)`.

---

## 13. Complete Before/After Example

### v2

```go
package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/form3tech-oss/f1/v2/pkg/f1"
	"github.com/form3tech-oss/f1/v2/pkg/f1/scenarios"
	"github.com/form3tech-oss/f1/v2/pkg/f1/testing"
)

func main() {
	f := f1.New().
		WithLogger(slog.Default()).
		Add("myLoadTest", myScenario,
			scenarios.Description("API load test"),
			scenarios.Parameter(scenarios.ScenarioParameter{Name: "rate", Default: "1/s"}),
		)
	f.Execute()
}

func myScenario(t *testing.T) testing.RunFn {
	t.Cleanup(func() { fmt.Println("cleanup") })
	return func(t *testing.T) {
		if err := doWork(); err != nil {
			t.Error(err)
		}
	}
}
```

### v3

```go
package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/form3tech-oss/f1/v3/pkg/f1"
	"github.com/form3tech-oss/f1/v3/pkg/f1/f1testing"
	"github.com/form3tech-oss/f1/v3/pkg/f1/scenarios"
)

func main() {
	f := f1.New(
		f1.WithLogger(slog.Default()),
	).
		AddScenario("myLoadTest", myScenario,
			scenarios.WithDescription("API load test"),
			scenarios.WithParameter(scenarios.ScenarioParameter{Name: "rate", Default: "1/s"}),
		)
	f.Execute()
}

func myScenario(ctx context.Context, t *f1testing.T) f1testing.RunFn {
	t.Cleanup(func() { fmt.Println("cleanup") })
	return func(ctx context.Context, t *f1testing.T) {
		if err := doWork(); err != nil {
			t.Error(err)
		}
	}
}
```

---

## 14. Migration Checklist

Use this checklist when migrating from v2 to v3:

- [ ] Update `go.mod`: `github.com/form3tech-oss/f1/v2` → `github.com/form3tech-oss/f1/v3`
- [ ] Update imports: `pkg/f1/testing` → `pkg/f1/f1testing`
- [ ] Update scenario signature: add `ctx context.Context` as first param to `ScenarioFn` and `RunFn`
- [ ] Replace `Add(` → `AddScenario(`
- [ ] Replace `Description(` → `WithDescription(`, `Parameter(` → `WithParameter(`
- [ ] Replace `New().WithLogger(l)` → `New(WithLogger(l))`, `New().WithStaticMetrics(m)` → `New(WithStaticMetrics(m))`
- [ ] Replace `ExecuteWithArgs(args)` → `Run(context.Background(), args)` (or `Run(ctx, args)` with a context)
- [ ] Replace `NewT()` with `NewTWithOptions()` if used (e.g. in tests)
- [ ] Replace `WithLogrusLogger()` with `WithLogger(*slog.Logger)` if used
- [ ] Remove `metrics.GetMetrics()` usage; use `WithStaticMetrics` for labels
- [ ] Remove `T.Time()` usage; record timing manually if needed
- [ ] Update `T.Logger()` call sites: it now returns `*slog.Logger` (not `*logrus.Logger`)
- [ ] (Optional) `Error`/`Fatal` now use `args ...any`; existing `Error(err)`/`Fatal(err)` calls remain valid
- [ ] Update CLI invocations: `--cpuprofile` → `--cpu-profile`, `--memprofile` → `--mem-profile`, `--iterationFrequency` → `--iteration-frequency`
- [ ] Remove `--verbose-fail` if used
- [ ] Remove `f1 chart` usage; use external visualisation tools

---

## Quick Reference: Search and Replace

| Find | Replace |
|------|---------|
| `f1/v2` | `f1/v3` |
| `pkg/f1/testing` | `pkg/f1/f1testing` |
| `*testing.T` | `*f1testing.T` |
| `testing.ScenarioFn` | `f1testing.ScenarioFn` |
| `testing.RunFn` | `f1testing.RunFn` |
| `f.Add(` | `f.AddScenario(` |
| `scenarios.Description(` | `scenarios.WithDescription(` |
| `scenarios.Parameter(` | `scenarios.WithParameter(` |
| `ExecuteWithArgs(` | `Run(context.Background(), ` |
| `New().WithLogger(` | `New(WithLogger(` |
| `New().WithStaticMetrics(` | `New(WithStaticMetrics(` |

**Note**: Scenario and Run function signatures require manual edits to add `ctx context.Context` as the first parameter.
