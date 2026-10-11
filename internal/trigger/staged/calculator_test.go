package staged_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/form3tech-oss/f1/v2/internal/trigger/staged"
)

func TestCalculatorWithDefaultStartTime(t *testing.T) {
	t.Parallel()

	calculator := staged.NewRateCalculator([]staged.Stage{
		{
			StartTarget: 1,
			EndTarget:   1,
			Duration:    1 * time.Minute,
		},
		{
			StartTarget: 10,
			EndTarget:   10,
			Duration:    10 * time.Minute,
		},
	}, nil)
	rate := calculator.Rate(time.Now())

	assert.Equal(t, 0, rate)
}

func TestCalculatorWithSetStartTime(t *testing.T) {
	t.Parallel()

	startTime := time.Now().Add(-2 * time.Minute)
	calculator := staged.NewRateCalculator([]staged.Stage{
		{
			StartTarget: 1,
			EndTarget:   10,
			Duration:    1 * time.Minute,
		},
		{
			StartTarget: 10,
			EndTarget:   10,
			Duration:    10 * time.Minute,
		},
	}, &startTime)
	rate := calculator.Rate(time.Now())

	assert.Equal(t, 10, rate)
}

func TestCalculatorWithSetStartTimeOutOfRange(t *testing.T) {
	t.Parallel()

	startTime := time.Now().Add(-20 * time.Minute)
	calculator := staged.NewRateCalculator([]staged.Stage{
		{
			StartTarget: 1,
			EndTarget:   1,
			Duration:    1 * time.Minute,
		},
		{
			StartTarget: 10,
			EndTarget:   10,
			Duration:    10 * time.Minute,
		},
	}, &startTime)
	rate := calculator.Rate(time.Now())

	assert.Equal(t, 0, rate)
}

func TestCalculatorWithStartTimeInTheFuture(t *testing.T) {
	t.Parallel()

	now := time.Now()
	startTime := now.Add(1 * time.Hour)
	calculator := staged.NewRateCalculator([]staged.Stage{
		{
			EndTarget: 10,
			Duration:  10 * time.Second,
		},
	}, &startTime)

	assert.Equal(t, 0, calculator.Rate(now))
	assert.Equal(t, 5, calculator.Rate(startTime.Add(5*time.Second)))
}
