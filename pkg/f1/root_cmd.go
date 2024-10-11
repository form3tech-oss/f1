package f1

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"

	"github.com/form3tech-oss/f1/v2/internal/chart"
	"github.com/form3tech-oss/f1/v2/internal/envsettings"
	"github.com/form3tech-oss/f1/v2/internal/metrics"
	"github.com/form3tech-oss/f1/v2/internal/run"
	"github.com/form3tech-oss/f1/v2/internal/trigger"
	"github.com/form3tech-oss/f1/v2/internal/ui"
	"github.com/form3tech-oss/f1/v2/pkg/f1/scenarios"
)

const (
	flagCPUProfile = "cpuprofile"
	flagMemProfile = "memprofile"
)

func buildRootCmd(
	scenarioList *scenarios.Scenarios,
	settings envsettings.Settings,
	p *profiling,
	output *ui.Output,
	staticMetrics map[string]string,
) (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:               getCmdName(),
		Short:             "F1 load testing tool",
		PersistentPreRunE: startProfiling(p),
		SilenceErrors:     true,
	}

	rootCmd.PersistentFlags().String(flagCPUProfile, "", "write cpu profile to `file`")
	if err := rootCmd.MarkPersistentFlagFilename(flagCPUProfile); err != nil {
		return nil, fmt.Errorf("marking flag as filename: %w", err)
	}
	rootCmd.PersistentFlags().String(flagMemProfile, "", "write memory profile to `file`")
	if err := rootCmd.MarkPersistentFlagFilename(flagMemProfile); err != nil {
		return nil, fmt.Errorf("marking flag as filename: %w", err)
	}

	metrics.InitWithStaticMetrics(settings.PrometheusEnabled(), staticMetrics)
	metricsInstance := metrics.Instance()

	builders := trigger.GetBuilders(output)

	rootCmd.AddCommand(run.Cmd(
		scenarioList,
		builders,
		settings,
		metricsInstance,
		output,
	))
	rootCmd.AddCommand(chart.Cmd(builders, output))
	rootCmd.AddCommand(scenarios.Cmd(scenarioList))
	rootCmd.AddCommand(completionsCmd(rootCmd))
	return rootCmd, nil
}

func startProfiling(p *profiling) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		var err error
		p.cpuProfileFileName, err = cmd.Flags().GetString(flagCPUProfile)
		if err != nil {
			return fmt.Errorf("getting flag: %w", err)
		}

		p.memProfileFileName, err = cmd.Flags().GetString(flagMemProfile)
		if err != nil {
			return fmt.Errorf("getting flag: %w", err)
		}

		if err := p.start(); err != nil {
			return fmt.Errorf("starting profiling: %w", err)
		}

		return nil
	}
}

func getCmdName() string {
	return path.Base(os.Args[0])
}
