package staged_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/form3tech-oss/f1/v2/internal/options"
	"github.com/form3tech-oss/f1/v2/internal/trigger/staged"
	"github.com/form3tech-oss/f1/v2/internal/ui"
	"github.com/form3tech-oss/f1/v2/internal/workers"
)

func TestParseStages_With_Valid_String(t *testing.T) {
	t.Parallel()

	value := "0s:1,10s:1,20s:20,1m:50,1h:200"
	expected := []staged.Stage{
		{
			StartTarget: 0,
			EndTarget:   1,
			Duration:    0 * time.Second,
		},
		{
			StartTarget: 0,
			EndTarget:   1,
			Duration:    10 * time.Second,
		},
		{
			StartTarget: 0,
			EndTarget:   20,
			Duration:    20 * time.Second,
		},
		{
			StartTarget: 0,
			EndTarget:   50,
			Duration:    1 * time.Minute,
		},
		{
			StartTarget: 0,
			EndTarget:   200,
			Duration:    1 * time.Hour,
		},
	}
	actual, err := staged.ParseStages(value)

	require.NoError(t, err)
	assert.ElementsMatch(t, actual, expected)
}

func TestParseStages_Error_Too_Many_Elements(t *testing.T) {
	t.Parallel()

	value := "0s:1:2"
	stages, err := staged.ParseStages(value)
	require.Error(t, err)
	assert.Nil(t, stages)
}

func TestParseStages_Error_Bad_Duration(t *testing.T) {
	t.Parallel()

	value := "0BB:1"
	stages, err := staged.ParseStages(value)
	require.Error(t, err)
	assert.Nil(t, stages)
}

func TestParseStages_Error_Bad_Target(t *testing.T) {
	t.Parallel()

	value := "1s:BB"
	stages, err := staged.ParseStages(value)
	require.Error(t, err)
	assert.Nil(t, stages)
}

func TestParseStages_Error_Negative_Target(t *testing.T) {
	t.Parallel()

	value := "1s:5,10s:-5"
	stages, err := staged.ParseStages(value)
	require.ErrorContains(t, err, "target -5 in stage 1 can't be negative")
	assert.Nil(t, stages)
}

func TestParseStages_Error_Negative_Duration(t *testing.T) {
	t.Parallel()

	value := "-10s:5"
	stages, err := staged.ParseStages(value)
	require.ErrorContains(t, err, "duration -10s in stage 0 can't be negative")
	assert.Nil(t, stages)
}

func TestRate_Error_No_Total_Duration(t *testing.T) {
	t.Parallel()

	builder := staged.Rate()
	require.NoError(t, builder.Flags.Parse([]string{"--stages", "0s:10,0s:20"}))

	_, err := builder.New(builder.Flags)

	require.ErrorContains(t, err, "stages `0s:10,0s:20` must have a total duration above 0")
}

func TestRate_StartTime(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, startTime string
		expectedStart   time.Time
	}{
		{
			name:          "RFC3339 UTC",
			startTime:     "2024-01-01T10:00:00Z",
			expectedStart: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		},
		{
			name:          "RFC3339 offset",
			startTime:     "2024-01-01T10:00:00-05:00",
			expectedStart: time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
		},
		{
			// the original layout matched "+07:00" literally, so these values mean UTC
			name:          "legacy +07:00 suffix",
			startTime:     "2024-01-01T10:00:00+07:00",
			expectedStart: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			builder := staged.Rate()
			require.NoError(t, builder.Flags.Parse([]string{
				"--startTime", test.startTime,
				"--stages", "10s:10",
				"--distribution", "none",
			}))

			trigger, err := builder.New(builder.Flags)
			require.NoError(t, err)

			// the single stage ramps from 0 to 10 over 10s from the start time
			assert.Equal(t, 0, trigger.DryRun(test.expectedStart.Add(-time.Second)), "before start")
			assert.Equal(t, 5, trigger.DryRun(test.expectedStart.Add(5*time.Second)), "mid stage")
		})
	}
}

func TestRate_UnparseableStartTime_WarnsAndStartsNow(t *testing.T) {
	t.Parallel()

	builder := staged.Rate()
	require.NoError(t, builder.Flags.Parse([]string{
		"--startTime", "2024-01-01 10:00",
		"--stages", "10s:10",
		"--distribution", "none",
	}))

	trigger, err := builder.New(builder.Flags)
	require.NoError(t, err, "v2 ignores start times it can't parse")

	// without a start time, the stages start at the first rate calculation
	now := time.Now()
	assert.Equal(t, 0, trigger.DryRun(now), "at start")
	assert.Equal(t, 5, trigger.DryRun(now.Add(5*time.Second)), "mid stage")

	var logs bytes.Buffer
	output := ui.NewOutput(slog.New(slog.NewTextHandler(&logs, nil)), ui.NewDiscardPrinter(), false, false)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	trigger.Trigger(ctx, output, workers.New(0, nil), options.RunOptions{})

	assert.Contains(t, logs.String(), "ignoring --startTime, stages start now: parsing startTime")
}
