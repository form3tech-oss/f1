package api_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/form3tech-oss/f1/v2/internal/trigger/api"
)

func TestTrigger_Deadline(t *testing.T) {
	t.Parallel()

	runStart := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	for _, test := range []struct {
		name        string
		trigger     api.Trigger
		maxDuration time.Duration
		expected    time.Time
	}{
		{
			name:        "no schedule",
			maxDuration: time.Minute,
			expected:    runStart.Add(time.Minute),
		},
		{
			name:        "schedule shorter than the max duration",
			trigger:     api.Trigger{Duration: 10 * time.Second},
			maxDuration: time.Minute,
			expected:    runStart.Add(10 * time.Second),
		},
		{
			name:        "schedule longer than the max duration",
			trigger:     api.Trigger{Duration: time.Hour},
			maxDuration: time.Minute,
			expected:    runStart.Add(time.Minute),
		},
		{
			name:        "schedule started in the past",
			trigger:     api.Trigger{Start: runStart.Add(-time.Minute), Duration: 90 * time.Second},
			maxDuration: time.Hour,
			expected:    runStart.Add(30 * time.Second),
		},
		{
			name:        "max duration counts from a schedule start in the future",
			trigger:     api.Trigger{Start: runStart.Add(10 * time.Minute), Duration: 5 * time.Minute},
			maxDuration: time.Second,
			expected:    runStart.Add(10*time.Minute + time.Second),
		},
		{
			name:        "schedule starting in the future ends with its last stage",
			trigger:     api.Trigger{Start: runStart.Add(10 * time.Minute), Duration: 5 * time.Minute},
			maxDuration: time.Hour,
			expected:    runStart.Add(15 * time.Minute),
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			deadline := test.trigger.Deadline(runStart, test.maxDuration)

			assert.Equal(t, test.expected, deadline)
		})
	}
}
