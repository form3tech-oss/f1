package triggerflags

import (
	"strings"

	"github.com/spf13/pflag"

	"github.com/form3tech-oss/f1/v2/internal/trigger/api"
)

const (
	FlagVerbose                  = "verbose"
	FlagIgnoreDropped            = "ignore-dropped"
	FlagMaxDuration              = "max-duration"
	FlagMaxIterations            = "max-iterations"
	FlagConcurrency              = "concurrency"
	FlagMaxFailures              = "max-failures"
	FlagMaxFailuresRate          = "max-failures-rate"
	FlagWaitForCompletionTimeout = "wait-for-completion-timeout"
)

const FlagDistribution = "distribution"

func DistributionFlag(flagSet *pflag.FlagSet) {
	distributionTypes := []string{
		string(api.NoneDistribution),
		string(api.RegularDistribution),
		string(api.RandomDistribution),
	}

	distributions := strings.Join(distributionTypes, "|")
	flagSet.String(FlagDistribution, string(api.RegularDistribution),
		"rate distribution: "+distributions)
}

const FlagJitter = "jitter"

func JitterFlag(flagSet *pflag.FlagSet) {
	flagSet.Float64P(FlagJitter, "j", 0.0,
		"random rate variation, e.g. 5 for ±5%")
}
