package progress

import (
	"sync/atomic"
	"time"

	"github.com/form3tech-oss/f1/v2/internal/metrics"
)

type Stats struct {
	successfulIterationDurations DurationStats
	failedIterationDurations     DurationStats

	droppedIterationCount atomic.Uint64
}

func (s *Stats) Record(result metrics.ResultType, duration time.Duration) {
	switch result {
	case metrics.SucessResult:
		s.successfulIterationDurations.Record(duration)
	case metrics.FailedResult:
		s.failedIterationDurations.Record(duration)
	case metrics.DroppedResult:
		s.droppedIterationCount.Add(1)
	case metrics.UnknownResult:
	}
}

func (s *Stats) Snapshot(period time.Duration) Snapshot {
	recentSufessfull, lifetimeSuccessful := s.successfulIterationDurations.CollectLifetime()
	_, lifetimeFailed := s.failedIterationDurations.CollectLifetime()

	return Snapshot{
		Period:                                period,
		DroppedIterationCount:                 s.droppedIterationCount.Load(),
		SuccessfulIterationDurationsForPeriod: recentSufessfull,
		SuccessfulIterationDurations:          lifetimeSuccessful,
		FailedIterationDurations:              lifetimeFailed,
	}
}

func (s *Stats) Total() Snapshot {
	_, lifetimeSuccessful := s.successfulIterationDurations.CollectLifetime()
	_, lifetimeFailed := s.failedIterationDurations.CollectLifetime()

	return Snapshot{
		DroppedIterationCount:        s.droppedIterationCount.Load(),
		SuccessfulIterationDurations: lifetimeSuccessful,
		FailedIterationDurations:     lifetimeFailed,
	}
}

type Snapshot struct {
	DroppedIterationCount                 uint64
	SuccessfulIterationDurationsForPeriod IterationDurationsSnapshot
	SuccessfulIterationDurations          IterationDurationsSnapshot
	FailedIterationDurations              IterationDurationsSnapshot
	Period                                time.Duration
}

func (s *Snapshot) Iterations() uint64 {
	return s.FailedIterationDurations.Count + s.SuccessfulIterationDurations.Count + s.DroppedIterationCount
}

func (s *Snapshot) IterationsStarted() uint64 {
	return s.SuccessfulIterationDurations.Count + s.FailedIterationDurations.Count
}

func (s *Snapshot) FailedIterationsRate() uint64 {
	return s.FailedIterationDurations.Count * 100 / s.Iterations()
}
