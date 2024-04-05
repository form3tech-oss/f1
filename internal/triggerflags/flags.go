package triggerflags

import (
	"strings"

	"github.com/spf13/pflag"

	"github.com/form3tech-oss/f1/v2/internal/trigger/api"
)

const (
	FlagVerbose         = "verbose"
	FlagVerboseFail     = "verbose-fail"
	FlagIgnoreDropped   = "ignore-dropped"
	FlagMaxDuration     = "max-duration"
	FlagMaxIterations   = "max-iterations"
	FlagConcurrency     = "concurrency"
	FlagMaxFailures     = "max-failures"
	FlagMaxFailuresRate = "max-failures-rate"
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
		"optional parameter to distribute the rate over steps of 100ms, which can be "+distributions)
}

const FlagJitter = "jitter"

func JitterFlag(flagSet *pflag.FlagSet) {
	flagSet.Float64P(FlagJitter, "j", 0.0,
		"vary the rate randomly by up to jitter percent")
}
