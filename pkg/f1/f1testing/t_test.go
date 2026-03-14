package f1testing_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/form3tech-oss/f1/v3/internal/log"
	"github.com/form3tech-oss/f1/v3/pkg/f1/f1testing"
)

// commonTInterface is the subset of methods that f1testing.T and testing.T share with identical
// signatures. This compile-time check ensures f1testing.T stays compatible with testing.T for
// these methods, enabling users to share test helpers between f1 scenarios and standard go tests.
// The interface is test-only and not exposed in the package.
var (
	_ commonTInterface = (*f1testing.T)(nil)
	_ commonTInterface = (*testing.T)(nil)
)

//nolint:interfacebloat,inamedparam // Single interface for compile-time verification; Cleanup matches testing.T signature
type commonTInterface interface {
	Cleanup(func())
	Error(args ...any)
	Errorf(format string, args ...any)
	Fail()
	FailNow()
	Failed() bool
	Fatal(args ...any)
	Fatalf(format string, args ...any)
	Log(args ...any)
	Logf(format string, args ...any)
	Name() string
}

func parseJSONLogLine(t *testing.T, line string) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(line), &m))
	return m
}

func assertLogFormat(t *testing.T, line string, wantLevel, wantMsg string, wantIteration float64, wantVUID float64) {
	t.Helper()
	m := parseJSONLogLine(t, line)
	require.Contains(t, m, "time", "log must have time field")
	require.Contains(t, m, "vuid", "log must have vuid field")
	require.Equal(t, wantLevel, m["level"], "level must match")
	require.Equal(t, wantMsg, m["msg"], "msg must match")
	iter, ok := m["iteration"].(float64)
	require.True(t, ok, "iteration must be float64 (JSON number)")
	require.InDelta(t, wantIteration, iter, 0, "iteration must match")
	vuid, ok := m["vuid"].(float64)
	require.True(t, ok, "vuid must be float64 (JSON number)")
	require.InDelta(t, wantVUID, vuid, 0, "vuid must match")
}

func TestNewTIsNotFailed(t *testing.T) {
	t.Parallel()

	newT, teardown := newT()
	defer teardown()

	require.False(t, newT.Failed())
}

func TestReportsPanicReasonWhenCleanupFails(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	newT, teardown := f1testing.NewTWithOptions("test", f1testing.WithLogger(logger))
	newT.Cleanup(func() {
		panic("boom")
	})
	teardown()
	require.Contains(t, buf.String(), "recovered panic in scenario")
}

func TestReportsErrorMessageWhenCleanupFails(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	newT, teardown := f1testing.NewTWithOptions("test", f1testing.WithLogger(logger))
	newT.Cleanup(func() {
		panic(errors.New("boom"))
	})
	teardown()
	logs := buf.String()
	require.Contains(t, logs, "recovered panic in scenario")
	require.Regexp(t, "stack_trace=\"goroutine", logs)
}

func TestCleanupCalledInReverseOrder(t *testing.T) {
	t.Parallel()

	var actual []int
	newT, teardown := newT()

	newT.Cleanup(func() {
		actual = append(actual, 1)
	})

	newT.Cleanup(func() {
		actual = append(actual, 2)
	})

	teardown()

	expected := []int{2, 1}
	require.Equal(t, expected, actual)
}

func TestFailNowSetsTheFailedState(t *testing.T) {
	t.Parallel()

	newT, teardown := newT()
	defer teardown()

	done := make(chan struct{})
	go func() {
		defer catchPanics(done)
		newT.FailNow()
	}()
	<-done

	require.True(t, newT.Failed())
}

func TestFailSetsTheFailedState(t *testing.T) {
	t.Parallel()

	newT, teardown := newT()
	defer teardown()

	done := make(chan struct{})
	go func() {
		defer catchPanics(done)
		newT.Fail()
	}()
	<-done

	require.True(t, newT.Failed())
}

func TestError(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		args          []any
		wantMsg       string
		wantIteration float64
		wantVUID      float64
	}{
		"error argument": {
			args:          []any{errors.New("boom")},
			wantMsg:       "boom",
			wantIteration: 0,
			wantVUID:      0,
		},
		"no arguments": {
			args:          []any{},
			wantMsg:       "",
			wantIteration: 0,
			wantVUID:      0,
		},
		"multiple arguments": {
			args:          []any{"expected", 42, "got", 0},
			wantMsg:       "expected 42 got 0",
			wantIteration: 0,
			wantVUID:      0,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&buf, nil))
			newT, teardown := f1testing.NewTWithOptions("test",
				f1testing.WithIteration(uint64(tc.wantIteration)),
				f1testing.WithVUID(int(tc.wantVUID)),
				f1testing.WithLogger(logger),
			)
			defer teardown()

			newT.Error(tc.args...)
			require.True(t, newT.Failed())
			assertLogFormat(t, strings.TrimSpace(buf.String()), "ERROR", tc.wantMsg, tc.wantIteration, tc.wantVUID)
		})
	}
}

func TestErrorf(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	newT, teardown := f1testing.NewTWithOptions("test",
		f1testing.WithIteration(0),
		f1testing.WithVUID(0),
		f1testing.WithLogger(logger),
	)
	defer teardown()

	newT.Errorf("got %d errors", 3)
	require.True(t, newT.Failed())
	assertLogFormat(t, strings.TrimSpace(buf.String()), "ERROR", "got 3 errors", 0, 0)
}

func TestFatal(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		args          []any
		wantMsg       string
		wantIteration float64
		wantVUID      float64
	}{
		"error argument": {
			args:          []any{errors.New("boom")},
			wantMsg:       "boom",
			wantIteration: 0,
			wantVUID:      0,
		},
		"no arguments": {
			args:          []any{},
			wantMsg:       "",
			wantIteration: 0,
			wantVUID:      0,
		},
		"multiple arguments": {
			args:          []any{"boom", 1, 2.0},
			wantMsg:       "boom 1 2",
			wantIteration: 0,
			wantVUID:      0,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&buf, nil))
			newT, teardown := f1testing.NewTWithOptions("test",
				f1testing.WithIteration(uint64(tc.wantIteration)),
				f1testing.WithVUID(int(tc.wantVUID)),
				f1testing.WithLogger(logger),
			)
			defer teardown()

			done := make(chan struct{})
			go func() {
				defer catchPanics(done)
				newT.Fatal(tc.args...)
			}()
			<-done

			require.True(t, newT.Failed())
			assertLogFormat(t, strings.TrimSpace(buf.String()), "ERROR", tc.wantMsg, tc.wantIteration, tc.wantVUID)
		})
	}
}

func TestFatalf(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	newT, teardown := f1testing.NewTWithOptions("test",
		f1testing.WithIteration(0),
		f1testing.WithVUID(0),
		f1testing.WithLogger(logger),
	)
	defer teardown()

	done := make(chan struct{})
	go func() {
		defer catchPanics(done)
		newT.Fatalf("fatal: %s", "boom")
	}()
	<-done

	require.True(t, newT.Failed())
	assertLogFormat(t, strings.TrimSpace(buf.String()), "ERROR", "fatal: boom", 0, 0)
}

func TestNameReturnsScenarioName(t *testing.T) {
	t.Parallel()

	newT, teardown := newT()
	defer teardown()

	require.Equal(t, "test", newT.Name())
}

func TestWithVUIDSetsVirtualUserID(t *testing.T) {
	t.Parallel()

	newT, teardown := f1testing.NewTWithOptions("test", f1testing.WithVUID(42))
	defer teardown()

	require.Equal(t, 42, newT.VUID)
}

func TestLog(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		call          func(*f1testing.T)
		wantMsg       string
		wantIteration float64
		wantVUID      float64
	}{
		"single argument": {
			call:          func(t *f1testing.T) { t.Log("info message") },
			wantMsg:       "info message",
			wantIteration: 0,
			wantVUID:      0,
		},
		"multiple arguments": {
			call:          func(t *f1testing.T) { t.Log("step", 1, "of", 3) },
			wantMsg:       "step 1 of 3",
			wantIteration: 0,
			wantVUID:      0,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&buf, nil))
			newT, teardown := f1testing.NewTWithOptions("test",
				f1testing.WithIteration(uint64(tc.wantIteration)),
				f1testing.WithVUID(int(tc.wantVUID)),
				f1testing.WithLogger(logger),
			)
			defer teardown()

			tc.call(newT)
			assertLogFormat(t, strings.TrimSpace(buf.String()), "INFO", tc.wantMsg, tc.wantIteration, tc.wantVUID)
		})
	}
}

func TestLogf(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	newT, teardown := f1testing.NewTWithOptions("test",
		f1testing.WithIteration(0),
		f1testing.WithVUID(0),
		f1testing.WithLogger(logger),
	)
	defer teardown()

	newT.Logf("progress: %d%%", 50)
	assertLogFormat(t, strings.TrimSpace(buf.String()), "INFO", "progress: 50%", 0, 0)
}

func catchPanics(done chan<- struct{}) {
	_ = recover()
	close(done)
}

func newT() (*f1testing.T, func()) {
	logger := log.NewDiscardLogger()

	return f1testing.NewTWithOptions(
		"test",
		f1testing.WithIteration(0),
		f1testing.WithLogger(logger),
	)
}
