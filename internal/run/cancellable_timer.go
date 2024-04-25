package run

import (
	"time"

	"github.com/form3tech-oss/f1/v2/internal/trace"
)

type CancellableTimer struct {
	cancel chan struct{}
	timer  *time.Timer
	C      chan bool
	tracer trace.Tracer
}

func NewCancellableTimer(d time.Duration, tracer trace.Tracer) *CancellableTimer {
	timer := &CancellableTimer{
		C:      make(chan bool),
		cancel: make(chan struct{}),
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
func (c *CancellableTimer) Cancel() {
	select {
	case c.cancel <- struct{}{}:
		c.tracer.Event("Closing Channel 'cancel'")
	default:
		return
	}
}
