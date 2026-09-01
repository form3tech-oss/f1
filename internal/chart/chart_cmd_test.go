package chart_test

import (
	"testing"
)

func TestChartConstant(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	given, when, then := NewChartTestStage(t)

	given.
		the_load_style_is_constant()

	when.
		i_execute_the_chart_command()

	then.
		the_command_is_successful().and().
		the_output_is_correct()
}

func TestChartStaged(t *testing.T) {
	t.Parallel()

	given, when, then := NewChartTestStage(t)

	given.
		the_load_style_is_staged()

	when.
		i_execute_the_chart_command()

	then.
		the_command_is_successful().and().
		the_output_is_correct()
}

func TestChartGaussian(t *testing.T) {
	t.Parallel()

	given, when, then := NewChartTestStage(t)

	given.
		the_load_style_is_gaussian_with_a_volume_of().and().
		the_chart_starts_at_a_fixed_time()

	when.
		i_execute_the_chart_command()

	then.
		the_command_is_successful().and().
		the_output_is_correct()
}

func TestChartGaussianWithJitter(t *testing.T) {
	t.Parallel()

	given, when, then := NewChartTestStage(t)

	given.
		the_load_style_is_gaussian_with_a_volume_of().and().
		jitter_is_applied().and().
		the_chart_starts_at_a_fixed_time()

	when.
		i_execute_the_chart_command()

	then.
		the_command_is_successful()
}

func TestChartRamp(t *testing.T) {
	t.Parallel()

	given, when, then := NewChartTestStage(t)

	given.
		the_load_style_is_ramp()

	when.
		i_execute_the_chart_command()

	then.
		the_command_is_successful().and().
		the_output_is_correct()
}

func TestChartFileConfig(t *testing.T) {
	t.Parallel()

	given, when, then := NewChartTestStage(t)

	given.
		the_load_style_is_defined_in_the_config_file()

	when.
		i_execute_the_chart_command()

	then.
		the_command_is_successful().and().
		the_output_is_correct()
}
