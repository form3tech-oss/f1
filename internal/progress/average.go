package progress

import (
	"sync/atomic"
	"time"
)

type IterationDurationsSnapshot struct {
	Average time.Duration
	Count   uint64
	Min     time.Duration
	Max     time.Duration
}

func (s IterationDurationsSnapshot) String() string {
	return "avg: " + s.Average.String() + ", " +
		"min: " + s.Min.String() + ", " +
		"max: " + s.Max.String()
}

// IterationDurations stores a execution times in nanoseconds
//
//	Each field is an atomic type for high-concurrency lock-free operation.
//	This leads to various inconsistencies when reading values, however
//	for the use case of progress reporting we prefer performance over perfect correctness.
type IterationDurations struct {
	// int64 can hold ~290 years of total execution durations in nanoseconds,
	// which should be enough for almost all realistic use cases.
	sum   atomic.Int64
	count atomic.Uint64

	min atomic.Int64
	max atomic.Int64
}

func (i *IterationDurations) Add(nanoseconds int64) {
	i.sum.Add(nanoseconds)
	i.count.Add(1)

	if nanoseconds > i.max.Load() {
		i.max.Store(nanoseconds)
	}

	min := i.min.Load()
	if min == 0 || nanoseconds < min {
		i.min.Store(nanoseconds)
	}
}

func (i *IterationDurations) Snapshot() IterationDurationsSnapshot {
	average, count := i.average()

	return IterationDurationsSnapshot{
		Average: time.Duration(average),
		Count:   count,
		Min:     time.Duration(i.min.Load()),
		Max:     time.Duration(i.max.Load()),
	}
}

func (i *IterationDurations) average() (int64, uint64) {
	count := i.count.Load()
	if count == 0 {
		return 0, 0
	}
	sum := i.sum.Load()

	return sum / int64(count), count
}

func (i *IterationDurations) Update(other *IterationDurations) {
	i.sum.Add(other.sum.Load())
	i.count.Add(other.count.Load())

	min := other.min.Load()
	if i.min.Load() == 0 || (i.min.Load() > min && min > 0) {
		i.min.Store(min)
	}

	max := other.max.Load()
	if i.max.Load() < max {
		i.max.Store(max)
	}
}

func (i *IterationDurations) Reset() {
	i.sum.Store(0)
	i.count.Store(0)
	i.max.Store(0)
	i.min.Store(0)
}

type DurationStats struct {
	running  IterationDurations
	lifetime IterationDurations
}

func (d *DurationStats) Record(duration time.Duration) {
	d.running.Add(duration.Abs().Nanoseconds())
}

func (d *DurationStats) CollectLifetime() (IterationDurationsSnapshot, IterationDurationsSnapshot) {
	running := d.running.Snapshot()
	d.lifetime.Update(&d.running)
	d.running.Reset()

	return running, d.lifetime.Snapshot()
}
