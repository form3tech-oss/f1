package chart

import (
	"testing"
)

func TestChartConstant(t *testing.T) {
	given, when, then := NewChartTestStage(t)

	given.
		the_load_style_is_constant().and().
		jitter_is_applied()

	when.
		i_execute_the_chart_command()

	then.
		the_command_is_successful()

}

func TestChartConstantNoJitter(t *testing.T) {
	given, when, then := NewChartTestStage(t)

	given.
		the_load_style_is_constant()

	when.
		i_execute_the_chart_command()

	then.
		the_command_is_successful()

}

func TestChartStaged(t *testing.T) {
	given, when, then := NewChartTestStage(t)

	given.
		the_load_style_is_staged("5m:100,2m:0,10s:100")

	when.
		i_execute_the_chart_command()

	then.
		the_command_is_successful()

}

func TestChartGaussian(t *testing.T) {
	given, when, then := NewChartTestStage(t)

	given.
		the_load_style_is_gaussian_with_a_volume_of(100000).and().
		the_chart_starts_at_a_fixed_time()

	when.
		i_execute_the_chart_command()

	then.
		the_command_is_successful()

}

func TestChartGaussianWithJitter(t *testing.T) {
	given, when, then := NewChartTestStage(t)

	given.
		the_load_style_is_gaussian_with_a_volume_of(100000).and().
		jitter_is_applied().and().
		the_chart_starts_at_a_fixed_time()

	when.
		i_execute_the_chart_command()

	then.
		the_command_is_successful()

}

func TestChartRamp(t *testing.T) {
	given, when, then := NewChartTestStage(t)

	given.
		the_load_style_is_ramp()

	when.
		i_execute_the_chart_command()

	then.
		the_command_is_successful()

}

func TestChartFileConfig(t *testing.T) {
	given, when, then := NewChartTestStage(t)

	given.
		the_load_style_is_defined_in_the_config_file("../testing/testdata/config-file.yaml")

	when.
		i_execute_the_chart_command()

	then.
		the_command_is_successful()

}
