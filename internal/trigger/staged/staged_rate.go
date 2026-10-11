package staged

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/pflag"

	"github.com/form3tech-oss/f1/v2/internal/options"
	"github.com/form3tech-oss/f1/v2/internal/trigger/api"
	"github.com/form3tech-oss/f1/v2/internal/triggerflags"
	"github.com/form3tech-oss/f1/v2/internal/ui"
	"github.com/form3tech-oss/f1/v2/internal/workers"
)

const (
	flagStages             = "stages"
	flagIterationFrequency = "iterationFrequency"
	flagStartTime          = "startTime"
)

func Rate() api.Builder {
	flags := pflag.NewFlagSet("staged", pflag.ContinueOnError)
	flags.StringP("stages", "s", "0s:1, 10s:1",
		"Comma separated list of <stage_duration>:<target_concurrent_iterations>. "+
			"During the stage, the number of concurrent iterations will ramp up or down to the target.")
	flags.DurationP(flagIterationFrequency, "f", 1*time.Second,
		"How frequently iterations should be started")
	flags.String(flagStartTime, "",
		"Starting point of stage calculation in RFC3339 format, defaults to now. "+
			"For compatibility, a +07:00 suffix is read as UTC, as in earlier versions")

	triggerflags.JitterFlag(flags)
	triggerflags.DistributionFlag(flags)

	return api.Builder{
		Name:        "staged <scenario>",
		Description: "triggers iterations at varying rates",
		Flags:       flags,
		New: func(params *pflag.FlagSet) (*api.Trigger, error) {
			jitterArg, err := params.GetFloat64(triggerflags.FlagJitter)
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}
			stg, err := params.GetString(flagStages)
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}
			frequency, err := params.GetDuration(flagIterationFrequency)
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}
			distributionTypeArg, err := params.GetString(triggerflags.FlagDistribution)
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}
			var startTime *time.Time
			startTimeStr, err := params.GetString(flagStartTime)
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}
			var startTimeErr error
			if startTimeStr != "" {
				parsedStartTime, err := parseStartTime(startTimeStr)
				if err == nil {
					startTime = &parsedStartTime
				} else {
					startTimeErr = err
				}
			}

			rates, err := CalculateStagedRate(jitterArg, frequency, stg, distributionTypeArg, startTime)
			if err != nil {
				return nil, err
			}

			trigger := api.NewIterationWorker(rates.IterationDuration, rates.Rate)
			if startTimeErr != nil {
				trigger = warnStartTimeIgnored(startTimeErr, trigger)
			}

			return &api.Trigger{
					Trigger: trigger,
					DryRun:  rates.Rate,
					Description: fmt.Sprintf(
						"Starting iterations every %s in numbers varying by time: %s, using distribution %s",
						frequency, stg, distributionTypeArg),
					Duration: rates.Duration,
				},
				nil
		},
	}
}

// legacyStartTimeLayout is the layout --startTime used to require. Its "+07:00" suffix is
// matched literally rather than as a zone offset, so such values have always meant UTC.
// They deliberately keep that meaning so existing scripts don't shift by 7 hours, even
// though RFC3339 would read +07:00 as an offset. Revisit in v3.
const legacyStartTimeLayout = "2006-01-02T15:04:05+07:00"

func parseStartTime(value string) (time.Time, error) {
	if startTime, err := time.Parse(legacyStartTimeLayout, value); err == nil {
		return startTime, nil
	}

	startTime, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parsing %s: %w", flagStartTime, err)
	}

	return startTime, nil
}

// warnStartTimeIgnored warns, when the run starts, that --startTime couldn't be parsed
// and the stages start now instead.
//
// Remove in v3: return the parse error from the builder instead. Earlier versions
// silently ignored values they couldn't parse, so for v2 compatibility this only warns.
func warnStartTimeIgnored(startTimeErr error, trigger api.WorkTriggerer) api.WorkTriggerer {
	return func(ctx context.Context, output *ui.Output, workers *workers.PoolManager, opts options.RunOptions) {
		output.Display(ui.WarningMessage{
			Message: fmt.Sprintf("ignoring --%s, stages start now: %s", flagStartTime, startTimeErr),
		})
		trigger(ctx, output, workers, opts)
	}
}

func CalculateStagedRate(
	jitterArg float64,
	frequency time.Duration,
	stg string,
	distributionTypeArg string,
	startTime *time.Time,
) (*api.Rates, error) {
	stages, err := ParseStages(stg)
	if err != nil {
		return nil, fmt.Errorf("parsing stages: %w", err)
	}

	calculator := NewRateCalculator(stages, startTime)
	rateFn := api.WithJitter(calculator.Rate, jitterArg)
	distributedIterationDuration, distributedRateFn, err := api.NewDistribution(
		api.DistributionType(distributionTypeArg), frequency, rateFn, nil,
	)
	if err != nil {
		return nil, fmt.Errorf("new distribution: %w", err)
	}

	return &api.Rates{
		IterationDuration: distributedIterationDuration,
		Rate:              distributedRateFn,
		// Known limitation: the duration is the sum of the stages, counted from now rather
		// than from --startTime, so a start time in the past or future ends the run early or
		// late relative to the stages. Left for v3.
		Duration: calculator.MaxDuration(),
	}, nil
}
