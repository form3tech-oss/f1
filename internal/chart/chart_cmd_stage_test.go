package chart_test

import (
	"io"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/form3tech-oss/f1/v2/internal/chart"
	"github.com/form3tech-oss/f1/v2/internal/console"
	"github.com/form3tech-oss/f1/v2/internal/trigger"
)

type ChartTestStage struct {
	t      *testing.T
	assert *assert.Assertions
	err    error
	args   []string
}

func NewChartTestStage(t *testing.T) (*ChartTestStage, *ChartTestStage, *ChartTestStage) {
	t.Helper()

	stage := &ChartTestStage{
		t:      t,
		assert: assert.New(t),
	}
	return stage, stage, stage
}

func (s *ChartTestStage) and() *ChartTestStage {
	return s
}

func (s *ChartTestStage) i_execute_the_chart_command() *ChartTestStage {
	printer := console.NewPrinter(io.Discard, io.Discard)
	cmd := chart.Cmd(trigger.GetBuilders(printer), printer)
	cmd.SetArgs(s.args)
	s.err = cmd.Execute()
	return s
}

func (s *ChartTestStage) the_command_is_successful() *ChartTestStage {
	s.assert.NoError(s.err)
	return s
}

func (s *ChartTestStage) the_load_style_is_constant() *ChartTestStage {
	s.args = append(s.args, "constant", "--rate", "10/s", "--distribution", "none")
	return s
}

func (s *ChartTestStage) jitter_is_applied() *ChartTestStage {
	s.args = append(s.args, "--jitter", "20", "--distribution", "none")
	return s
}

func (s *ChartTestStage) the_load_style_is_staged(stages string) *ChartTestStage {
	s.args = append(s.args, "staged", "--stages", stages, "--distribution", "none")
	return s
}

func (s *ChartTestStage) the_load_style_is_ramp() *ChartTestStage {
	s.args = append(s.args, "ramp", "--start-rate", "0/s", "--end-rate", "10/s", "--ramp-duration", "10s", "--chart-duration", "10s", "--distribution", "none")
	return s
}

func (s *ChartTestStage) the_load_style_is_gaussian_with_a_volume_of(volume int) *ChartTestStage {
	s.args = append(s.args, "gaussian", "--peak", "5m", "--repeat", "10m", "--volume", strconv.Itoa(volume), "--standard-deviation", "1m", "--distribution", "none")
	return s
}

func (s *ChartTestStage) the_chart_starts_at_a_fixed_time() *ChartTestStage {
	s.args = append(s.args, "--chart-start", time.Now().Truncate(10*time.Minute).Format(time.RFC3339))
	return s
}

func (s *ChartTestStage) the_load_style_is_defined_in_the_config_file(filename string) *ChartTestStage {
	s.args = append(s.args, "file", filename, "--chart-duration", "5s")
	return s
}
