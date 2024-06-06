package raterun

import (
	"context"
	"errors"
	"time"
)

// RunFunction is a function type that represents the function to be executed by the Runner.
//
// It will be called with the frequency at which the function is executed.
type RunFunction func(frequency time.Duration)

// Schedule configures when and how frequent the Runner will execute the function
type Schedule struct {
	// StartDelay is when the function will start executing at a certain frequency
	StartDelay time.Duration
	// Frequency configures how often the function will be executed during this Schedule
	Frequency time.Duration
}

// New creates a new runner that will execute fn as defined by the provided schedules
//
// Each Schedule in schedules defines how often fn should be executed at any given point in time.
func New(fn RunFunction, schedules []Schedule) (*Runner, error) {
	if len(schedules) == 0 {
		return nil, errors.New("empty schedules")
	}

	rateRunner := &Runner{
		restart:     make(chan struct{}, 1),
		runFunction: fn,
		schedules:   newSchedules(schedules),
		stopped:     make(chan struct{}),
	}

	return rateRunner, nil
}

type Runner struct {
	restart     chan struct{}
	runFunction RunFunction

	schedules *schedules
	cancel    context.CancelFunc
	stopped   chan struct{}
}

// Restart will stop the current schedule and start from the first one defined.
func (r *Runner) Restart() {
	r.restart <- struct{}{}
}

// Start starts the execution of the runner.
//
// Start is non-blockig and runs in a go routine. The provided context can be used to manage the
// lifecycle. Stop() will also terminate the runner.
func (r *Runner) Start(ctx context.Context) {
	defer close(r.stopped)

	schedulesCtx, schedulesCtxCancel := context.WithCancel(ctx)
	r.cancel = schedulesCtxCancel

	go func() {
		for {
			select {
			case <-r.restart:
				r.schedules.startFirst()
			case <-r.schedules.timeUntilNextSchedule():
				r.schedules.startNext()
			case <-r.schedules.currentScheduleTicker():
				r.runFunction(r.schedules.currentFrequency())
			case <-schedulesCtx.Done():
				r.schedules.stop()
				return
			}
		}
	}()
}

// Stop stopps the runner and will block until the runner is stopped
func (r *Runner) Stop() {
	r.cancel()
	<-r.stopped
}

type schedules struct {
	ticker               *time.Ticker
	nextScheduleTimer    *time.Timer
	list                 []Schedule
	currentScheduleIndex int
}

func newSchedules(list []Schedule) *schedules {
	return &schedules{
		list:                 list,
		currentScheduleIndex: -1,
		ticker:               time.NewTicker(time.Hour),
		nextScheduleTimer:    time.NewTimer(list[0].StartDelay),
	}
}

func (s *schedules) start(index int) {
	if index >= len(s.list) {
		return
	}

	s.ticker.Stop()
	s.currentScheduleIndex = index
	s.ticker = time.NewTicker(s.list[s.currentScheduleIndex].Frequency)

	nextIndex := s.currentScheduleIndex + 1
	s.nextScheduleTimer.Stop()
	if nextIndex >= len(s.list) {
		return
	}

	s.nextScheduleTimer = time.NewTimer(s.list[nextIndex].StartDelay)
}

func (s *schedules) startFirst() {
	s.start(0)
}

func (s *schedules) startNext() {
	s.start(s.currentScheduleIndex + 1)
}

func (s *schedules) currentFrequency() time.Duration {
	return s.list[s.currentScheduleIndex].Frequency
}

func (s *schedules) stop() {
	s.ticker.Stop()
	s.nextScheduleTimer.Stop()
}

func (s *schedules) timeUntilNextSchedule() <-chan time.Time {
	return s.nextScheduleTimer.C
}

func (s *schedules) currentScheduleTicker() <-chan time.Time {
	return s.ticker.C
}
