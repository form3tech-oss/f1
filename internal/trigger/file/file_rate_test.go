package file_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/form3tech-oss/f1/v2/internal/trigger/file"
	"github.com/form3tech-oss/f1/v2/internal/ui"
)

func TestFileRate_StageStartSetsTheSchedule(t *testing.T) {
	t.Parallel()

	// Every stage ended long ago. The trigger is still built, so the chart command can
	// plot it; the run command rejects it.
	stageStart := time.Date(2020, 12, 10, 9, 0, 0, 0, time.UTC)
	configFile := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(configFile, []byte(`
scenario: template
limits:
  max-duration: 1h
  concurrency: 50
  max-iterations: 100
  ignore-dropped: true
schedule:
  stage-start: "`+stageStart.Format(time.RFC3339)+`"
stages:
- duration: 1m
  mode: constant
  rate: 1/s
  distribution: none
- duration: 1m
  mode: constant
  rate: 2/s
  distribution: none
`), 0o600))

	builder := file.Rate(ui.NewDiscardOutput())
	require.NoError(t, builder.Flags.Parse([]string{configFile}))

	trigger, err := builder.New(builder.Flags)

	require.NoError(t, err)
	assert.True(t, stageStart.Equal(trigger.Start), "start %s, expected %s", trigger.Start, stageStart)
	assert.Equal(t, 2*time.Minute, trigger.Duration)
	// The dry run follows the schedule, whenever the chart starts.
	assert.Equal(t, 0, trigger.DryRun(stageStart.Add(-time.Second)))
	assert.Equal(t, 1, trigger.DryRun(stageStart.Add(30*time.Second)))
	assert.Equal(t, 2, trigger.DryRun(stageStart.Add(90*time.Second)))
	assert.Equal(t, 0, trigger.DryRun(stageStart.Add(2*time.Minute)))
}

func TestFileRate_RampStageResumesMidSchedule(t *testing.T) {
	t.Parallel()

	stageStart := time.Date(2020, 12, 10, 9, 0, 0, 0, time.UTC)
	fileContent := `
scenario: template
limits:
  max-duration: 1h
  concurrency: 50
  max-iterations: 100
  ignore-dropped: true
schedule:
  stage-start: "` + stageStart.Format(time.RFC3339) + `"
stages:
- duration: 1m
  mode: constant
  rate: 1/s
  distribution: none
- duration: 100s
  mode: ramp
  start-rate: 0/s
  end-rate: 100/s
  distribution: none
`
	stagesToRun, err := file.ParseConfigFile([]byte(fileContent))

	require.NoError(t, err)
	require.Len(t, stagesToRun.Stages, 2)
	// The ramp starts at 09:01, so half way through it the rate is half way up, even when
	// that is the first time the rate is asked for.
	assert.Equal(t, 50, stagesToRun.Stages[1].Rate(stageStart.Add(time.Minute+50*time.Second)))
}
