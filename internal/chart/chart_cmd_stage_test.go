package chart_test

import (
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"

	"github.com/form3tech-oss/f1/v2/internal/chart"
	"github.com/form3tech-oss/f1/v2/internal/log"
	"github.com/form3tech-oss/f1/v2/internal/trigger"
	"github.com/form3tech-oss/f1/v2/internal/ui"
)

type ChartTestStage struct {
	t              *testing.T
	assert         *assert.Assertions
	err            error
	args           []string
	output         *ui.Output
	results        *lineStringBuilder
	expectedOutput string
}

func NewChartTestStage(t *testing.T) (*ChartTestStage, *ChartTestStage, *ChartTestStage) {
	t.Helper()

	sb := newLineStringBuilder()
	p := ui.Printer{
		Writer:    sb,
		ErrWriter: sb,
	}

	stage := &ChartTestStage{
		t:       t,
		assert:  assert.New(t),
		output:  ui.NewOutput(log.NewDiscardLogger(), &p, true, true),
		results: sb,
	}
	return stage, stage, stage
}

func (s *ChartTestStage) and() *ChartTestStage {
	return s
}

func (s *ChartTestStage) i_execute_the_chart_command() *ChartTestStage {
	cmd := chart.Cmd(trigger.GetBuilders(s.output), s.output)
	cmd.SetArgs(s.args)
	s.err = cmd.Execute()
	return s
}

func (s *ChartTestStage) the_command_is_successful() *ChartTestStage {
	s.assert.NoError(s.err)
	return s
}

func (s *ChartTestStage) the_output_is_correct() *ChartTestStage {
	s.assert.Equal(s.expectedOutput, s.results.String())
	return s
}

func (s *ChartTestStage) the_load_style_is_constant() *ChartTestStage {
	s.args = append(s.args, "constant", "--rate", "10/s", "--distribution", "none")
	f, err := os.ReadFile("../testdata/expected-constant-chart-output.txt")
	s.assert.NoError(err)
	s.expectedOutput = string(f)
	return s
}

func (s *ChartTestStage) jitter_is_applied() *ChartTestStage {
	s.args = append(s.args, "--jitter", "20", "--distribution", "none")
	return s
}

func (s *ChartTestStage) the_load_style_is_staged() *ChartTestStage {
	s.args = append(s.args, "staged", "--stages", "5m:100,2m:0,10s:100", "--distribution", "none")
	f, err := os.ReadFile("../testdata/expected-staged-chart-output.txt")
	s.assert.NoError(err)
	s.expectedOutput = string(f)
	return s
}

func (s *ChartTestStage) the_load_style_is_ramp() *ChartTestStage {
	s.args = append(s.args, "ramp", "--start-rate", "0/s", "--end-rate", "10/s", "--ramp-duration", "10s", "--chart-duration", "10s", "--distribution", "none")
	f, err := os.ReadFile("../testdata/expected-ramp-chart-output.txt")
	s.assert.NoError(err)
	s.expectedOutput = string(f)
	return s
}

func (s *ChartTestStage) the_load_style_is_gaussian_with_a_volume_of() *ChartTestStage {
	s.args = append(s.args, "gaussian", "--peak", "5m", "--repeat", "10m", "--volume", "100000", "--standard-deviation", "1m", "--distribution", "none")
	f, err := os.ReadFile("../testdata/expected-gaussian-chart-output.txt")
	s.assert.NoError(err)
	s.expectedOutput = string(f)
	return s
}

func (s *ChartTestStage) the_chart_starts_at_a_fixed_time() *ChartTestStage {
	s.args = append(s.args, "--chart-start", "2024-09-19T17:00:00Z")
	return s
}

func (s *ChartTestStage) the_load_style_is_defined_in_the_config_file() *ChartTestStage {
	s.args = append(s.args, "file", "../testdata/config-file.yaml", "--chart-duration", "5s")
	f, err := os.ReadFile("../testdata/expected-file-chart-output.txt")
	s.assert.NoError(err)
	s.expectedOutput = string(f)
	return s
}

type lineStringBuilder struct {
	sb *strings.Builder
}

func newLineStringBuilder() *lineStringBuilder {
	return &lineStringBuilder{sb: &strings.Builder{}}
}

func (l *lineStringBuilder) Write(p []byte) (n int, err error) {
	return l.sb.Write(p)
}

func (l *lineStringBuilder) String() string {
	return strings.Replace(l.sb.String(), "\\n", "\n", -1)
}
