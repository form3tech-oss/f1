package testing

import (
	"time"
)

type CancellableTimer struct {
	cancel chan bool
	timer  *time.Timer
	C      chan bool
	reset  chan time.Duration
}

func NewCancellableTimer(d time.Duration) *CancellableTimer {
	timer := &CancellableTimer{
		cancel: make(chan bool),
		C:      make(chan bool),
		reset:  make(chan time.Duration),
		timer:  time.NewTimer(d),
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
			c.C <- false
			return
		}
	}
}

// Cancel makes all the waiters receive false
func (c *CancellableTimer) Cancel() {
	close(c.cancel)
}

func (c *CancellableTimer) Reset(duration time.Duration) {
	c.reset <- duration
}
