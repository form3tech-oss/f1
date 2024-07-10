package run

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/form3tech-oss/f1/v2/internal/envsettings"
	"github.com/form3tech-oss/f1/v2/internal/metrics"
	"github.com/form3tech-oss/f1/v2/internal/options"
	"github.com/form3tech-oss/f1/v2/internal/trigger/api"
	"github.com/form3tech-oss/f1/v2/internal/triggerflags"
	"github.com/form3tech-oss/f1/v2/internal/ui"
	"github.com/form3tech-oss/f1/v2/pkg/f1/scenarios"
)

func Cmd(
	s *scenarios.Scenarios,
	builders []api.Builder,
	settings envsettings.Settings,
	metricsInstance *metrics.Metrics,
	output *ui.Output,
) *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "run <subcommand>",
		Short: "Runs a test scenario",
	}

	for _, t := range builders {
		triggerCmd := &cobra.Command{
			Use:   t.Name,
			Short: t.Description,
			RunE:  runCmdExecute(s, t, settings, metricsInstance, output),
			Args:  cobra.MatchAll(cobra.ExactArgs(1)),
		}

		triggerCmd.Flags().BoolP(triggerflags.FlagVerbose, "v", false, "enables log output to stdout")
		triggerCmd.Flags().Bool(triggerflags.FlagVerboseFail, false, "DEPRECATED: log output to stdout on failure")

		if !t.IgnoreCommonFlags {
			triggerCmd.ValidArgs = s.GetScenarioNames()

			triggerCmd.Flags().Bool(triggerflags.FlagIgnoreDropped, false, "dropped requests will not fail the run")
			triggerCmd.Flags().DurationP(triggerflags.FlagMaxDuration, "d", time.Second,
				"--max-duration 1s (stop after 1 second)")
			triggerCmd.Flags().IntP(triggerflags.FlagConcurrency, "c", 100,
				"--concurrency 2 (allow at most 2 groups of iterations to run concurrently)")
			triggerCmd.Flags().Uint64P(triggerflags.FlagMaxIterations, "i", 0,
				"--max-iterations 100 (stop after 100 iterations, regardless of remaining duration)")
			triggerCmd.Flags().Uint64(triggerflags.FlagMaxFailures, 0,
				"--max-failures 10 (load test will fail if more than 10 errors occurred, default is 0)")
			triggerCmd.Flags().Int(triggerflags.FlagMaxFailuresRate, 0,
				"--max-failures-rate 5 (load test will fail if more than 5\\% requests failed, default is 0)")
		}

		triggerCmd.Flags().AddFlagSet(t.Flags)
		runCmd.AddCommand(triggerCmd)
	}

	return runCmd
}

func runCmdExecute(
	s *scenarios.Scenarios,
	t api.Builder,
	settings envsettings.Settings,
	metricsInstance *metrics.Metrics,
	output *ui.Output,
) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true

		trig, err := t.New(cmd.Flags())
		if err != nil {
			return fmt.Errorf("creating trigger command: %w", err)
		}

		var scenarioName string
		var duration time.Duration
		var concurrency int
		var maxIterations uint64
		var maxFailures uint64
		var maxFailuresRate int
		var ignoreDropped bool
		if t.IgnoreCommonFlags {
			scenarioName = trig.Options.Scenario
			duration = trig.Options.MaxDuration
			concurrency = trig.Options.Concurrency
			maxIterations = trig.Options.MaxIterations
			maxFailures = trig.Options.MaxFailures
			maxFailuresRate = trig.Options.MaxFailuresRate
			ignoreDropped = trig.Options.IgnoreDropped
		} else {
			scenarioName = args[0]
			duration, err = cmd.Flags().GetDuration(triggerflags.FlagMaxDuration)
			if err != nil {
				return fmt.Errorf("getting flag: %w", err)
			}
			concurrency, err = cmd.Flags().GetInt(triggerflags.FlagConcurrency)
			if err != nil {
				return fmt.Errorf("getting flag: %w", err)
			}
			if concurrency < 1 {
				return fmt.Errorf("concurrency %d can't be less than 1", concurrency)
			}

			maxIterations, err = cmd.Flags().GetUint64(triggerflags.FlagMaxIterations)
			if err != nil {
				return fmt.Errorf("getting flag: %w", err)
			}
			maxFailures, err = cmd.Flags().GetUint64(triggerflags.FlagMaxFailures)
			if err != nil {
				return fmt.Errorf("getting flag: %w", err)
			}
			maxFailuresRate, err = cmd.Flags().GetInt(triggerflags.FlagMaxFailuresRate)
			if err != nil {
				return fmt.Errorf("getting flag: %w", err)
			}
			ignoreDropped, err = cmd.Flags().GetBool(triggerflags.FlagIgnoreDropped)
			if err != nil {
				return fmt.Errorf("getting flag: %w", err)
			}
		}

		verbose, err := cmd.Flags().GetBool(triggerflags.FlagVerbose)
		if err != nil {
			return fmt.Errorf("getting flag: %w", err)
		}

		verboseFail, err := cmd.Flags().GetBool(triggerflags.FlagVerboseFail)
		if err != nil {
			return fmt.Errorf("getting flag: %w", err)
		}
		if verboseFail {
			output.Display(ui.WarningMessage{Message: "--verbose-fail option has been removed"})
		}

		if settings.Fluentd.Present() {
			output.Display(ui.WarningMessage{
				Message: fmt.Sprintf("WARNING: fluentd integration has been removed. %s and %s have no effect.",
					envsettings.EnvFluentdHost,
					envsettings.EnvFluentdPort,
				),
			},
			)
		}

		run, err := NewRun(options.RunOptions{
			Scenario:        scenarioName,
			MaxDuration:     duration,
			Concurrency:     concurrency,
			Verbose:         verbose,
			MaxIterations:   maxIterations,
			MaxFailures:     maxFailures,
			MaxFailuresRate: maxFailuresRate,
			IgnoreDropped:   ignoreDropped,
		}, s, trig, settings, metricsInstance, output)
		if err != nil {
			return fmt.Errorf("new run: %w", err)
		}
		result, err := run.Do(cmd.Context())
		if err != nil {
			return fmt.Errorf("internal error on run: %w", err)
		}

		if result.Error() != nil {
			return result.Error()
		} else if result.Failed() {
			return errors.New("load test failed - see log for details")
		}
		cmd.SilenceUsage = false
		return nil
	}
}
