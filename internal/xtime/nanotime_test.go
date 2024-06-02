package xtime_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/form3tech-oss/f1/v2/internal/xtime"
)

func TestNanoTime(t *testing.T) {
	t.Parallel()

	for range 100 {
		t1 := xtime.NanoTime()
		t2 := xtime.NanoTime()
		assert.LessOrEqual(t, t1, t2, "monotonic clock should always increase")
	}
}
