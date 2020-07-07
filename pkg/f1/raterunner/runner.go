package raterunner

import (
	"errors"
	"time"
)

// RunFunction function to be executed at specified rate
type RunFunction func(rate time.Duration, t time.Time)

type Rate struct {
	Start time.Duration // when to start this rate
	Rate  time.Duration // run in amount of Duration
}

// New creates a new runner which varies in time according to given rates
func New(fn RunFunction, rates []Rate) (*RateRunner, error) {
	if len(rates) == 0 {
		return nil, errors.New("empty rates")
	}

	rateRunner := &RateRunner{
		terminateRunner: make(chan bool, 1),
		restartRates:    make(chan bool, 1),
		runFunction:     fn,
		rates:           rates,
		nextRateIndex:   0,
	}

	return rateRunner, nil
}

type RateRunner struct {
	terminateRunner chan bool
	restartRates    chan bool
	// function that is going to be run at specific timed intervals, according to current rate set at a specific moment in time
	runFunction RunFunction
	rates       []Rate
	// index for the next rate in rates array
	nextRateIndex int
	// runs runFunction at current rate interval
	fnTicker *time.Ticker
	// rateTimer controls when to Start next rate interval
	rateTimer *time.Timer
}

// Terminate finishes the runner
func (rr *RateRunner) Terminate() {
	rr.terminateRunner <- true
}

// RestartRate resets function calling back to first defined rate
func (rr *RateRunner) RestartRate() {
	rr.restartRates <- true
}

// run starts running the function following the rates given to the constructor
func (rr *RateRunner) Run() {
	go func() {
		rr.rateTimer = time.NewTimer(rr.rates[0].Start)
		rr.fnTicker = time.NewTicker(time.Hour)
		for {
			select {
			case <-rr.restartRates:
				rr.nextRateIndex = 0
				rr.scheduleNextRate(rr.nextRateIndex)
			case <-rr.rateTimer.C:
				rate := rr.rates[rr.nextRateIndex]
				rr.nextRateIndex++
				rr.scheduleNextRate(rr.nextRateIndex)
				rr.runAtRate(rate)
			case t := <-rr.fnTicker.C:
				rr.runFunction(rr.rates[rr.nextRateIndex-1].Rate, t)
			case <-rr.terminateRunner:
				rr.rateTimer.Stop()
				rr.fnTicker.Stop()
				return
			}
		}
	}()
}

func (rr *RateRunner) scheduleNextRate(rateIndex int) {
	if rateIndex < len(rr.rates) {
		nextRate := rr.rates[rateIndex]
		// close rateTimer if it hasn't run yet to prevent double runs
		rr.rateTimer.Stop()
		rr.rateTimer = time.NewTimer(nextRate.Start)
	}
}

func (rr *RateRunner) runAtRate(rate Rate) {
	rr.fnTicker.Stop()
	rr.fnTicker = time.NewTicker(rate.Rate)
}
