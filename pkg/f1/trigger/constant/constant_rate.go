package constant

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/form3tech-oss/f1/pkg/f1/trigger/api"

	"github.com/asaskevich/govalidator"
	"github.com/spf13/pflag"
)

func ConstantRate() api.Builder {
	flags := pflag.NewFlagSet("constant", pflag.ContinueOnError)
	flags.StringP("rate", "r", "1/s", "number of iterations to start per interval, in the form <request>/<duration>")
	flags.Float64P("jitter", "j", 0.0, "vary the rate randomly by up to jitter percent")
	return api.Builder{
		Name:        "constant",
		Description: "triggers test iterations at a constant rate",
		Flags:       flags,
		New: func(params *pflag.FlagSet) (*api.Trigger, error) {
			rateArg, err := params.GetString("rate")
			if err != nil {
				return nil, err
			}
			jitterArg, err := params.GetFloat64("jitter")
			if err != nil {
				return nil, err
			}
			rate := 0
			iterationDuration := 1 * time.Second
			if strings.Contains(rateArg, "/") {
				rate, err = strconv.Atoi((rateArg)[0:strings.Index(rateArg, "/")])
				if err != nil {
					return nil, fmt.Errorf("unable to parse rate %s", rateArg)
				}
				unit := (rateArg)[strings.Index(rateArg, "/")+1:]
				if !govalidator.IsNumeric(unit[0:1]) {
					unit = "1" + unit
				}
				iterationDuration, err = time.ParseDuration(unit)
				if err != nil {
					return nil, fmt.Errorf("unable to parse unit %s", rateArg)
				}
			} else {
				rate, err = strconv.Atoi(rateArg)
				if err != nil {
					return nil, fmt.Errorf("unable to parse rate %s", rateArg)
				}
			}

			rateFn := api.WithJitter(func(time.Time) int { return rate }, jitterArg)
			return &api.Trigger{
					Trigger:     api.NewIterationWorker(iterationDuration, rateFn),
					Description: fmt.Sprintf("%d iterations every %s", rate, iterationDuration),
					DryRun:      rateFn,
				},
				nil
		},
	}
}
