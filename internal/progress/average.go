package progress

import (
	"sync"
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
	return snapshotOf(i.sum.Load(), i.count.Load(), i.min.Load(), i.max.Load())
}

func snapshotOf(sum, count, minimum, maximum int64) IterationDurationsSnapshot {
	if count <= 0 {
		return IterationDurationsSnapshot{}
	}

	return IterationDurationsSnapshot{
		Average: time.Duration(sum / count),
		Count:   uint64(count),
		Min:     time.Duration(minimum),
		Max:     time.Duration(maximum),
	}
}

// Update moves the values of other into i, resets other, and returns the values moved.
// Adds to other may run concurrently, but Updates into i must not.
//
// Each value is read and zeroed in one step with Swap(0). Reading it with Load and then
// zeroing it with Store would lose any Add to other that ran between the two: the Store
// would wipe it before it reached i. With Swap, an Add that runs at the same time lands
// either in the swapped value, and so in i and the returned snapshot, or in other after
// the swap, to be moved by the next Update.
//
// The four values are swapped one after the other, so an Add that runs during an Update
// can be split between the values it returns and the next Update's. The totals in i catch
// up, but each of the two returned snapshots can rarely have a skewed average, or a min
// or max that is 0 or from an iteration it doesn't count.
func (i *IterationDurations) Update(other *IterationDurations) IterationDurationsSnapshot {
	sum := other.sum.Swap(0)
	count := other.count.Swap(0)
	currentMin := other.min.Swap(0)
	currentMax := other.max.Swap(0)

	i.sum.Add(sum)
	i.count.Add(count)

	if i.min.Load() == 0 || (i.min.Load() > currentMin && currentMin > 0) {
		i.min.Store(currentMin)
	}

	if i.max.Load() < currentMax {
		i.max.Store(currentMax)
	}

	return snapshotOf(sum, count, currentMin, currentMax)
}

// DurationStats tracks iteration durations for progress reporting.
type DurationStats struct {
	// collectMu serialises CollectLifetime, which updates lifetime with plain Load and Store.
	// Record doesn't take it.
	collectMu sync.Mutex
	// running holds the iterations recorded since the last CollectLifetime.
	// Workers add to it concurrently through Record.
	running IterationDurations
	// lifetime holds the totals for the whole run.
	lifetime IterationDurations
}

func (d *DurationStats) Record(nanoseconds int64) {
	d.running.Add(nanoseconds)
}

// CollectLifetime moves the running values into the lifetime totals. It returns the values
// moved, for the period since the last call, and the new totals.
//
// It runs on every progress tick, while workers may still be calling Record, and once more for
// the final summary. Update makes the move without losing those records, so a record is
// counted in at most one period. Records moved by the final call are in no period.
func (d *DurationStats) CollectLifetime() (IterationDurationsSnapshot, IterationDurationsSnapshot) {
	d.collectMu.Lock()
	defer d.collectMu.Unlock()

	period := d.lifetime.Update(&d.running)

	return period, d.lifetime.Snapshot()
}
