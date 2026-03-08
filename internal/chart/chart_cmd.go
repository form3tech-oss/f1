package chart

import (
	"fmt"
	"time"

	"github.com/guptarohit/asciigraph"
	"github.com/spf13/cobra"

	"github.com/form3tech-oss/f1/v2/internal/trigger/api"
	"github.com/form3tech-oss/f1/v2/internal/ui"
)

const (
	flagChartStart    = "chart-start"
	flagChartDuration = "chart-duration"
)

func Cmd(builders []api.Builder, output *ui.Output) *cobra.Command {
	chartCmd := &cobra.Command{
		Use:   "chart <subcommand>",
		Short: "plots a chart of the test scenarios that would be triggered over time with the provided run function",
	}

	for _, t := range builders {
		triggerCmd := &cobra.Command{
			Use:   t.Name,
			Short: t.Description,
			RunE:  chartCmdExecute(t, output),
		}
		triggerCmd.Flags().String(flagChartStart, time.Now().Format(time.RFC3339), "Optional start time for the chart")
		triggerCmd.Flags().Duration(flagChartDuration, 10*time.Minute, "Duration for the chart")
		triggerCmd.Flags().AddFlagSet(t.Flags)
		chartCmd.AddCommand(triggerCmd)
	}

	return chartCmd
}

func chartCmdExecute(
	t api.Builder,
	output *ui.Output,
) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		cmd.SilenceUsage = true

		startString, err := cmd.Flags().GetString(flagChartStart)
		if err != nil {
			return fmt.Errorf("getting flag: %w", err)
		}
		start, err := time.Parse(time.RFC3339, startString)
		if err != nil {
			return fmt.Errorf("parsing start time: %w", err)
		}
		duration, err := cmd.Flags().GetDuration(flagChartDuration)
		if err != nil {
			return fmt.Errorf("getting flag: %w", err)
		}

		trig, err := t.New(cmd.Flags())
		if err != nil {
			return fmt.Errorf("creating builder: %w", err)
		}

		if trig.DryRun == nil {
			return fmt.Errorf("%s does not support charting predicted load", cmd.Name())
		}

		current := start
		end := current.Add(duration)
		width := 160
		sampleInterval := duration / time.Duration(width-1)

		rates := []float64{0.0}
		times := []time.Time{current}
		for ; current.Add(sampleInterval).Before(end); current = current.Add(sampleInterval) {
			rate := trig.DryRun(current)
			rates = append(rates, float64(rate))
			times = append(times, current)
		}

		output.Display(ui.InteractiveMessage{
			Message: asciigraph.Plot(rates, asciigraph.Height(15), asciigraph.Width(width)),
		})

		return nil
	}
}
