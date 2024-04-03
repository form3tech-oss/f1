package chart

import (
	"fmt"
	"os"
	"time"

	"github.com/guptarohit/asciigraph"
	"github.com/spf13/cobra"
	"github.com/wcharczuk/go-chart/v2"

	"github.com/form3tech-oss/f1/v2/internal/console"
	"github.com/form3tech-oss/f1/v2/internal/support/errorh"
	"github.com/form3tech-oss/f1/v2/internal/trace"
	"github.com/form3tech-oss/f1/v2/internal/trigger/api"
)

func Cmd(builders []api.Builder, tracer trace.Tracer, printer *console.Printer) *cobra.Command {
	chartCmd := &cobra.Command{
		Use:   "chart <subcommand>",
		Short: "plots a chart of the test scenarios that would be triggered over time with the provided run function",
	}

	for _, t := range builders {
		triggerCmd := &cobra.Command{
			Use:   t.Name,
			Short: t.Description,
			RunE:  chartCmdExecute(t, tracer, printer),
		}
		triggerCmd.Flags().String("chart-start", time.Now().Format(time.RFC3339), "Optional start time for the chart")
		triggerCmd.Flags().Duration("chart-duration", 10*time.Minute, "Duration for the chart")
		triggerCmd.Flags().String("filename", "", fmt.Sprintf("Filename for optional detailed chart, e.g. %s.png", t.Name))
		triggerCmd.Flags().AddFlagSet(t.Flags)
		chartCmd.AddCommand(triggerCmd)
	}

	return chartCmd
}

func chartCmdExecute(
	t api.Builder,
	tracer trace.Tracer,
	printer *console.Printer,
) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		cmd.SilenceUsage = true

		startString, err := cmd.Flags().GetString("chart-start")
		if err != nil {
			return fmt.Errorf("getting flag: %w", err)
		}
		start, err := time.Parse(time.RFC3339, startString)
		if err != nil {
			return fmt.Errorf("parsing start time: %w", err)
		}
		duration, err := cmd.Flags().GetDuration("chart-duration")
		if err != nil {
			return fmt.Errorf("getting flag: %w", err)
		}
		filename, err := cmd.Flags().GetString("filename")
		if err != nil {
			return fmt.Errorf("getting flag: %w", err)
		}

		trig, err := t.New(cmd.Flags(), tracer)
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

		printer.Println(asciigraph.Plot(rates, asciigraph.Height(15), asciigraph.Width(width)))

		if filename == "" {
			return nil
		}
		graph := chart.Chart{
			Title:      trig.Description,
			TitleStyle: chart.StyleTextDefaults(),
			Width:      1920,
			Height:     1024,
			YAxis: chart.YAxis{
				Name:      "Triggered Test Iterations",
				NameStyle: chart.StyleTextDefaults(),
				Style:     chart.StyleTextDefaults(),
				AxisType:  chart.YAxisSecondary,
			},
			XAxis: chart.XAxis{
				Name:           "Time",
				NameStyle:      chart.StyleTextDefaults(),
				ValueFormatter: chart.TimeMinuteValueFormatter,
				Style:          chart.StyleTextDefaults(),
			},
			Series: []chart.Series{
				chart.TimeSeries{
					Style: chart.Style{
						StrokeColor: chart.GetDefaultColor(0).WithAlpha(64),
					},
					Name:    "testing",
					XValues: times,
					YValues: rates,
				},
			},
		}

		f, err := os.Create(filename)
		if err != nil {
			return fmt.Errorf("creting file: %w", err)
		}
		defer errorh.SafeClose(f)

		err = graph.Render(chart.PNG, f)
		if err != nil {
			return fmt.Errorf("rendering graph: %w", err)
		}
		printer.Printf("Detailed chart written to %s\n", filename)
		return nil
	}
}
