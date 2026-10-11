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
	count atomic.Int64

	min atomic.Int64
	max atomic.Int64
}

func (i *IterationDurations) Add(nanoseconds int64) {
	i.sum.Add(nanoseconds)
	i.count.Add(1)

	if nanoseconds > i.max.Load() {
		i.max.Store(nanoseconds)
	}

	currentMin := i.min.Load()
	if currentMin == 0 || nanoseconds < currentMin {
		i.min.Store(nanoseconds)
	}
}

func (i *IterationDurations) Snapshot() IterationDurationsSnapshot {
	return durationTotals{
		sum:   i.sum.Load(),
		count: i.count.Load(),
		min:   i.min.Load(),
		max:   i.max.Load(),
	}.snapshot()
}

type durationTotals struct {
	sum, count, min, max int64
}

func (t durationTotals) snapshot() IterationDurationsSnapshot {
	if t.count <= 0 {
		return IterationDurationsSnapshot{
			Min: time.Duration(t.min),
			Max: time.Duration(t.max),
		}
	}

	return IterationDurationsSnapshot{
		Average: time.Duration(t.sum / t.count),
		Count:   uint64(t.count),
		Min:     time.Duration(t.min),
		Max:     time.Duration(t.max),
	}
}

// drain takes the current values and resets them to zero. Each counter is swapped
// atomically, so values added concurrently are kept for the next drain rather than lost.
// The counters are not swapped together, so a value added during a drain can be split
// across two periods; totals over the whole run stay exact.
func (i *IterationDurations) drain() durationTotals {
	return durationTotals{
		sum:   i.sum.Swap(0),
		count: i.count.Swap(0),
		min:   i.min.Swap(0),
		max:   i.max.Swap(0),
	}
}

func (i *IterationDurations) merge(totals durationTotals) {
	i.sum.Add(totals.sum)
	i.count.Add(totals.count)

	if i.min.Load() == 0 || (i.min.Load() > totals.min && totals.min > 0) {
		i.min.Store(totals.min)
	}

	if i.max.Load() < totals.max {
		i.max.Store(totals.max)
	}
}

type DurationStats struct {
	running  IterationDurations
	lifetime IterationDurations
}

func (d *DurationStats) Record(nanoseconds int64) {
	d.running.Add(nanoseconds)
}

func (d *DurationStats) CollectLifetime() (IterationDurationsSnapshot, IterationDurationsSnapshot) {
	totals := d.running.drain()
	d.lifetime.merge(totals)

	return totals.snapshot(), d.lifetime.Snapshot()
}
