package workers

import (
	"context"
	"time"
)

// deadline is a context's deadline, or zero if it has none. The deadline timer cancels the
// context from a separate goroutine, so the context may not reflect a passed deadline yet.
type deadline time.Time

func newDeadline(ctx context.Context) deadline {
	d, _ := ctx.Deadline()
	return deadline(d)
}

func (d deadline) passed() bool {
	t := time.Time(d)
	return !t.IsZero() && !time.Now().Before(t)
}
