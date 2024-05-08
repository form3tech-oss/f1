package staged

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCalculatorWithDefaultStartTime(t *testing.T) {
	t.Parallel()

	calculator := newRateCalculator([]stage{
		{
			startTarget: 1,
			endTarget:   1,
			duration:    1 * time.Minute,
		},
		{
			startTarget: 10,
			endTarget:   10,
			duration:    10 * time.Minute,
		},
	}, nil)
	rate := calculator.Rate(time.Now())

	assert.Equal(t, 0, rate)
}

func TestCalculatorWithSetStartTime(t *testing.T) {
	t.Parallel()

	startTime := time.Now().Add(-2 * time.Minute)
	calculator := newRateCalculator([]stage{
		{
			startTarget: 1,
			endTarget:   10,
			duration:    1 * time.Minute,
		},
		{
			startTarget: 10,
			endTarget:   10,
			duration:    10 * time.Minute,
		},
	}, &startTime)
	rate := calculator.Rate(time.Now())

	assert.Equal(t, 10, rate)
}

func TestCalculatorWithSetStartTimeOutOfRange(t *testing.T) {
	t.Parallel()

	startTime := time.Now().Add(-20 * time.Minute)
	calculator := newRateCalculator([]stage{
		{
			startTarget: 1,
			endTarget:   1,
			duration:    1 * time.Minute,
		},
		{
			startTarget: 10,
			endTarget:   10,
			duration:    10 * time.Minute,
		},
	}, &startTime)
	rate := calculator.Rate(time.Now())

	assert.Equal(t, 0, rate)
}
