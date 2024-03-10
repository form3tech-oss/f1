package run

import (
	"sync/atomic"
	"time"

	"github.com/form3tech-oss/f1/v2/internal/trace"
)

type CancellableTimer struct {
	cancel    chan bool
	timer     *time.Timer
	C         chan bool
	reset     chan time.Duration
	cancelled int32
	tracer    trace.Tracer
}

func NewCancellableTimer(d time.Duration, tracer trace.Tracer) *CancellableTimer {
	timer := &CancellableTimer{
		cancel: make(chan bool),
		C:      make(chan bool),
		reset:  make(chan time.Duration),
		timer:  time.NewTimer(d),
		tracer: tracer,
	}

	go timer.wait()

	return timer
}

// internal wait goroutine wrapping time.After
func (c *CancellableTimer) wait() {
	for {
		select {
		case d := <-c.reset:
			c.timer.Reset(d)
		case <-c.timer.C:
			c.C <- true
			return
		case <-c.cancel:
			c.tracer.ReceivedFromChannel("cancel")
			c.C <- false
			c.tracer.SentToChannel("C")
			return
		}
	}
}

// Cancel makes all the waiters receive false
func (c *CancellableTimer) Cancel() bool {
	c.tracer.Event("Closing Channel 'cancel'")
	if atomic.CompareAndSwapInt32(&c.cancelled, 0, 1) {
		close(c.cancel)
		return true
	}
	return false
}

func (c *CancellableTimer) Reset(duration time.Duration) {
	c.reset <- duration
}
