package templates

import "time"

//nolint:lll // templates read better with long lines
const (
	timeoutTemplate              = `{cyan}[{{durationSeconds .Duration | printf "%5s"}}]  Max Duration Elapsed - waiting for active tests to complete{-}`
	maxIterationsReachedTemplate = `{cyan}[{{durationSeconds .Duration | printf "%5s"}}]  Max Iterations Reached - waiting for active tests to complete{-}`
	interruptTemplate            = `{cyan}[{{durationSeconds .Duration | printf "%5s"}}]  Interrupted - waiting for active tests to complete{-}`
)

type exitData struct {
	Duration time.Duration
}

type (
	TimeoutData              exitData
	MaxIterationsReachedData exitData
	InterruptData            exitData
)

func (t *Templates) Timeout(data TimeoutData) string {
	return render(t.timeout, data)
}

func (t *Templates) MaxIterationsReached(data MaxIterationsReachedData) string {
	return render(t.maxIterationsReached, data)
}

func (t *Templates) Interrupt(data InterruptData) string {
	return render(t.interrupt, data)
}
