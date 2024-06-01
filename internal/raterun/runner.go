package raterun

import (
	"context"
	"errors"
	"time"
)

// RunFunction function to be executed every period
type RunFunction func(period time.Duration)

type Schedule struct {
	StartDelay time.Duration
	Frequency  time.Duration
}

func New(fn RunFunction, schedules []Schedule) (*Runner, error) {
	if len(schedules) == 0 {
		return nil, errors.New("empty schedules")
	}

	rateRunner := &Runner{
		restart:     make(chan struct{}, 1),
		runFunction: fn,
		schedules:   newSchedules(schedules),
	}

	return rateRunner, nil
}

type Runner struct {
	restart     chan struct{}
	runFunction RunFunction

	schedules *schedules
}

func (r *Runner) Restart() {
	r.restart <- struct{}{}
}

func (r *Runner) Run(ctx context.Context) {
	go func() {
		for {
			select {
			case <-r.restart:
				r.schedules.startFirst()
			case <-r.schedules.timeUntilNextSchedule():
				r.schedules.startNext()
			case <-r.schedules.currentScheduleTicker():
				r.runFunction(r.schedules.currentFrequency())
			case <-ctx.Done():
				r.schedules.stop()
				return
			}
		}
	}()
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
