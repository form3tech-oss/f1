package run

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/form3tech-oss/f1/v2/internal/logging"
	"github.com/form3tech-oss/f1/v2/internal/options"
	"github.com/form3tech-oss/f1/v2/internal/trigger/api"
	"github.com/form3tech-oss/f1/v2/pkg/f1/scenarios"
)

func Cmd(s *scenarios.Scenarios, builders []api.Builder, hookFunc logging.RegisterLogHookFunc) *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "run <subcommand>",
		Short: "Runs a test scenario",
	}

	for _, t := range builders {
		if t.IgnoreCommonFlags {
			triggerCmd := &cobra.Command{
				Use:   t.Name,
				Short: t.Description,
				RunE:  runCmdExecute(s, t, hookFunc),
				Args:  cobra.MatchAll(cobra.ExactArgs(1)),
			}
			triggerCmd.Flags().BoolP("verbose", "v", false, "enables log output to stdout")
			triggerCmd.Flags().Bool("verbose-fail", false, "log output to stdout on failure")

			triggerCmd.Flags().AddFlagSet(t.Flags)
			runCmd.AddCommand(triggerCmd)
		} else {
			triggerCmd := &cobra.Command{
				Use:       t.Name,
				Short:     t.Description,
				RunE:      runCmdExecute(s, t, hookFunc),
				Args:      cobra.MatchAll(cobra.ExactArgs(1)),
				ValidArgs: s.GetScenarioNames(),
			}
			triggerCmd.Flags().BoolP("verbose", "v", false, "enables log output to stdout")
			triggerCmd.Flags().Bool("verbose-fail", false, "log output to stdout on failure")

			triggerCmd.Flags().Bool("ignore-dropped", false, "dropped requests will not fail the run")
			triggerCmd.Flags().DurationP("max-duration", "d", time.Second, "--max-duration 1s (stop after 1 second)")
			triggerCmd.Flags().IntP("concurrency", "c", 100, "--concurrency 2 (allow at most 2 groups of iterations to run concurrently)")
			triggerCmd.Flags().Int32P("max-iterations", "i", 0, "--max-iterations 100 (stop after 100 iterations, regardless of remaining duration)")
			triggerCmd.Flags().Int("max-failures", 0, "--max-failures 10 (load test will fail if more than 10 errors occurred, default is 0)")
			triggerCmd.Flags().Int("max-failures-rate", 0, "--max-failures-rate 5 (load test will fail if more than 5\\% requests failed, default is 0)")

			triggerCmd.Flags().AddFlagSet(t.Flags)
			runCmd.AddCommand(triggerCmd)
		}
	}

	return runCmd
}

func runCmdExecute(s *scenarios.Scenarios, t api.Builder, hookFunc logging.RegisterLogHookFunc) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		trig, err := t.New(cmd.Flags())
		if err != nil {
			return errors.Wrap(err, "error creating trigger command")
		}

		var scenarioName string
		var duration time.Duration
		var concurrency int
		var maxIterations int32
		var maxFailures int
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
			duration, err = cmd.Flags().GetDuration("max-duration")
			if err != nil {
				return fmt.Errorf("invalid max-duration value: %w", err)
			}
			concurrency, err = cmd.Flags().GetInt("concurrency")
			if err != nil || concurrency < 1 {
				return fmt.Errorf("invalid concurrency value: %w", err)
			}
			maxIterations, err = cmd.Flags().GetInt32("max-iterations")
			if err != nil {
				return fmt.Errorf("invalid max-iterations value: %w", err)
			}
			maxFailures, err = cmd.Flags().GetInt("max-failures")
			if err != nil {
				return fmt.Errorf("invalid max-failures value: %w", err)
			}
			maxFailuresRate, err = cmd.Flags().GetInt("max-failures-rate")
			if err != nil {
				return fmt.Errorf("invalid max-failures-rate value: %w", err)
			}
			ignoreDropped, err = cmd.Flags().GetBool("ignore-dropped")
			if err != nil {
				return fmt.Errorf("invalid ignore-dropped value: %w", err)
			}
		}

		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			return fmt.Errorf("invalid verbose value: %w", err)
		}

		verboseFail, err := cmd.Flags().GetBool("verbose-fail")
		if err != nil {
			return fmt.Errorf("invalid verbose-fail value: %w", err)
		}

		run, err := NewRun(options.RunOptions{
			Scenario:            scenarioName,
			MaxDuration:         duration,
			Concurrency:         concurrency,
			Verbose:             verbose,
			VerboseFail:         verboseFail,
			MaxIterations:       maxIterations,
			MaxFailures:         maxFailures,
			MaxFailuresRate:     maxFailuresRate,
			RegisterLogHookFunc: hookFunc,
			IgnoreDropped:       ignoreDropped,
		}, trig)
		if err != nil {
			return err
		}
		result := run.Do(s)
		if result.Error() != nil {
			return result.Error()
		} else if result.Failed() {
			return errors.New("load test failed - see log for details")
		}
		cmd.SilenceUsage = false
		return nil
	}
}
