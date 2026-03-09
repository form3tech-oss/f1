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

// runTriggerUsageTemplate uses groupedFlagUsages for the Flags section.
const runTriggerUsageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Commands:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

{{groupedFlagUsages . | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

func Cmd(
	s *scenarios.Scenarios,
	builders []api.Builder,
	settings envsettings.Settings,
	metricsInstance *metrics.Metrics,
	output *ui.Output,
) *cobra.Command {
	registerHelpTemplateFunc()

	runCmd := &cobra.Command{
		Use:   "run <subcommand>",
		Short: "Runs a test scenario",
	}

	for _, t := range builders {
		triggerCmd := &cobra.Command{
			Use:   t.Name,
			Short: t.Description,
			Long:  t.Long,
			RunE:  runCmdExecute(s, t, settings, metricsInstance, output),
			Args:  cobra.MatchAll(cobra.ExactArgs(1)),
		}

		triggerCmd.Flags().SortFlags = false

		// Output
		triggerCmd.Flags().BoolP(triggerflags.FlagVerbose, "v", false, "enable log output to stdout")

		if !t.IgnoreCommonFlags {
			triggerCmd.ValidArgs = s.GetScenarioNames()

			// Duration & limits
			triggerCmd.Flags().DurationP(triggerflags.FlagMaxDuration, "d", time.Second,
				"stop after duration (e.g. 1s, 5m)")
			triggerCmd.Flags().Uint64P(triggerflags.FlagMaxIterations, "i", 0,
				"stop after N iterations (0 = unlimited)")

			// Concurrency
			triggerCmd.Flags().IntP(triggerflags.FlagConcurrency, "c", 100,
				"max concurrent iteration groups (e.g. 2, 100)")

			// Failure handling
			triggerCmd.Flags().Uint64(triggerflags.FlagMaxFailures, 0,
				"fail run if error count exceeds N (0 = disabled)")
			triggerCmd.Flags().Int(triggerflags.FlagMaxFailuresRate, 0,
				"fail run if error rate exceeds N%% (0 = disabled)")
			triggerCmd.Flags().Bool(triggerflags.FlagIgnoreDropped, false,
				"do not fail run when requests are dropped")

			// Shutdown
			triggerCmd.Flags().Duration(triggerflags.FlagWaitForCompletionTimeout, 10*time.Second,
				"wait for active iterations before exit (e.g. 10s)")
		}

		triggerCmd.Flags().AddFlagSet(t.Flags)
		triggerCmd.SetUsageTemplate(runTriggerUsageTemplate)
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
		var waitForCompletionTimeout time.Duration
		if t.IgnoreCommonFlags {
			scenarioName = trig.Options.Scenario
			duration = trig.Options.MaxDuration
			concurrency = trig.Options.Concurrency
			maxIterations = trig.Options.MaxIterations
			maxFailures = trig.Options.MaxFailures
			maxFailuresRate = trig.Options.MaxFailuresRate
			ignoreDropped = trig.Options.IgnoreDropped
			waitForCompletionTimeout = trig.Options.WaitForCompletionTimeout
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
			waitForCompletionTimeout, err = cmd.Flags().GetDuration(triggerflags.FlagWaitForCompletionTimeout)
			if err != nil {
				return fmt.Errorf("getting flag: %w", err)
			}
		}

		verbose, err := cmd.Flags().GetBool(triggerflags.FlagVerbose)
		if err != nil {
			return fmt.Errorf("getting flag: %w", err)
		}

		run, err := NewRun(options.RunOptions{
			Scenario:                 scenarioName,
			MaxDuration:              duration,
			Concurrency:              concurrency,
			Verbose:                  verbose,
			MaxIterations:            maxIterations,
			MaxFailures:              maxFailures,
			MaxFailuresRate:          maxFailuresRate,
			IgnoreDropped:            ignoreDropped,
			WaitForCompletionTimeout: waitForCompletionTimeout,
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
