package testing_test

import (
	"bytes"
	"errors"
	"log/slog"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/form3tech-oss/f1/v2/internal/log"
	f1testing "github.com/form3tech-oss/f1/v2/pkg/f1/testing"
)

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
	logs := buf.String()
	require.Contains(t, logs, "recovered panic in scenario")
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
	require.Regexp(t, regexp.MustCompile("stack_trace=\"goroutine"), logs)
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

func TestErrorSetsTheFailedState(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		args []any
	}{
		"error argument": {
			args: []any{errors.New("boom")},
		},
		"no arguments": {
			args: []any{},
		},
		"random arguments": {
			args: []any{"boom", 1, 2.0},
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			newT, teardown := newT()
			defer teardown()

			newT.Error(test.args...)
			require.True(t, newT.Failed())
		})
	}
}

func TestErrorfSetsTheFailedState(t *testing.T) {
	t.Parallel()

	newT, teardown := newT()
	defer teardown()

	newT.Errorf("boom")
	require.True(t, newT.Failed())
}

func TestFatalSetsTheFailedState(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		args []any
	}{
		"error argument": {
			args: []any{errors.New("boom")},
		},
		"no arguments": {
			args: []any{},
		},
		"random arguments": {
			args: []any{"boom", 1, 2.0},
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			newT, teardown := newT()
			defer teardown()

			done := make(chan struct{})
			go func() {
				defer catchPanics(done)
				newT.Fatal(test.args...)
			}()
			<-done

			require.True(t, newT.Failed())
		})
	}
}

func TestFatalfSetsTheFailedState(t *testing.T) {
	t.Parallel()

	newT, teardown := newT()
	defer teardown()

	done := make(chan struct{})
	go func() {
		defer catchPanics(done)
		newT.Fatalf("boom")
	}()
	<-done

	require.True(t, newT.Failed())
}

func TestNameReturnsScenarioName(t *testing.T) {
	t.Parallel()

	newT, teardown := newT()
	defer teardown()

	require.Equal(t, "test", newT.Name())
}

func catchPanics(done chan<- struct{}) {
	_ = recover()
	close(done)
}

func newT() (*f1testing.T, func()) {
	logger := log.NewDiscardLogger()
	logrus := log.NewSlogLogrusLogger(logger)

	return f1testing.NewTWithOptions(
		"test",
		f1testing.WithIteration("iteration 0"),
		f1testing.WithLogger(logger),
		f1testing.WithLogrusLogger(logrus),
	)
}
