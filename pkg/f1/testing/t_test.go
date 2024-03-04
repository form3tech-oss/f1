package testing_test

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	go_testing "testing"

	"github.com/stretchr/testify/require"

	"github.com/form3tech-oss/f1/v2/pkg/f1/testing"
)

func TestNewTIsNotFailed(t *go_testing.T) {
	newT, _ := testing.NewT("iteration 0", "test")
	require.False(t, newT.Failed())
}

func TestReportsPanicReasonWhenCleanupFails(t *go_testing.T) {
	var buf bytes.Buffer
	newT, teardown := testing.NewT("iteration 0", "test")

	defer func() {
		newT.Logger().SetOutput(os.Stderr)
	}()

	newT.Logger().SetOutput(&buf)
	newT.Cleanup(func() {
		panic("boom")
	})

	teardown()
	logs := buf.String()
	require.Contains(t, logs, "panic in 'test' scenario on iteration 0")
}

func TestReportsErrorMessageWhenCleanupFails(t *go_testing.T) {
	var buf bytes.Buffer
	newT, teardown := testing.NewT("iteration 0", "test")

	defer func() {
		newT.Logger().SetOutput(os.Stderr)
	}()

	newT.Logger().SetOutput(&buf)
	newT.Cleanup(func() {
		panic(fmt.Errorf("boom"))
	})

	teardown()
	logs := buf.String()
	require.Contains(t, logs, "panic in 'test' scenario on iteration 0")
	require.Regexp(t, regexp.MustCompile("stack_trace=\"goroutine"), logs)
}

func TestCleanupCalledInReverseOrder(t *go_testing.T) {
	var actual []int
	newT, teardown := testing.NewT("iteration 0", "test")

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

func TestFailNowSetsTheFailedState(t *go_testing.T) {
	newT, _ := testing.NewT("iteration 0", "test")

	done := make(chan struct{})
	go func() {
		defer catchPanics(done)
		newT.FailNow()
	}()
	<-done

	require.True(t, newT.Failed())
}

func TestFailSetsTheFailedState(t *go_testing.T) {
	newT, _ := testing.NewT("iteration 0", "test")

	done := make(chan struct{})
	go func() {
		defer catchPanics(done)
		newT.Fail()
	}()
	<-done

	require.True(t, newT.Failed())
}

func TestErrorSetsTheFailedState(t *go_testing.T) {
	newT, _ := testing.NewT("iteration 0", "test")

	newT.Error(fmt.Errorf("boom"))
	require.True(t, newT.Failed())
}

func TestErrorfSetsTheFailedState(t *go_testing.T) {
	newT, _ := testing.NewT("iteration 0", "test")

	newT.Errorf("boom")
	require.True(t, newT.Failed())
}

func TestFatalSetsTheFailedState(t *go_testing.T) {
	newT, _ := testing.NewT("iteration 0", "test")

	done := make(chan struct{})
	go func() {
		defer catchPanics(done)
		newT.Fatal(fmt.Errorf("boom"))
	}()
	<-done

	require.True(t, newT.Failed())
}

func TestFatalfSetsTheFailedState(t *go_testing.T) {
	newT, _ := testing.NewT("iteration 0", "test")

	done := make(chan struct{})
	go func() {
		defer catchPanics(done)
		newT.Fatalf("boom")
	}()
	<-done

	require.True(t, newT.Failed())
}

func TestNameReturnsScenarioName(t *go_testing.T) {
	newT, _ := testing.NewT("iteration 0", "test")
	require.Equal(t, "test", newT.Name())
}

func catchPanics(done chan<- struct{}) {
	recover()
	close(done)
}
