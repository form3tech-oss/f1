package views

import (
	"log/slog"
	"time"

	"github.com/form3tech-oss/f1/v2/internal/log"
)

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

func (d TimeoutData) Log(logger *slog.Logger) {
	logger.Info("Max Duration Elapsed - waiting for active tests to complete", log.DurationAttr(d.Duration))
}

func (v *Views) Timeout(data TimeoutData) *ViewContext[TimeoutData] {
	return &ViewContext[TimeoutData]{
		view: v.timeout,
		data: data,
	}
}

func (d MaxIterationsReachedData) Log(logger *slog.Logger) {
	logger.Info("Max Iterations Reached - waiting for active tests to complete", log.DurationAttr(d.Duration))
}

func (v *Views) MaxIterationsReached(data MaxIterationsReachedData) *ViewContext[MaxIterationsReachedData] {
	return &ViewContext[MaxIterationsReachedData]{
		view: v.maxIterationsReached,
		data: data,
	}
}

func (d InterruptData) Log(logger *slog.Logger) {
	logger.Info("Interrupted - waiting for active tests to complete", log.DurationAttr(d.Duration))
}

func (v *Views) Interrupt(data InterruptData) *ViewContext[InterruptData] {
	return &ViewContext[InterruptData]{
		view: v.interrupt,
		data: data,
	}
}
