package users

import (
	"time"

	"github.com/form3tech-oss/f1/pkg/f1/options"
	"github.com/form3tech-oss/f1/pkg/f1/trigger/api"
	"github.com/spf13/pflag"
)

func UsersRate() api.Builder {
	flags := pflag.NewFlagSet("users", pflag.ContinueOnError)

	return api.Builder{
		Name:        "users",
		Description: "triggers test iterations from a static set of virtual users controlled by the --concurrency flag",
		Flags:       flags,
		New: func(params *pflag.FlagSet) (*api.Trigger, error) {
			trigger := func(doWork chan<- bool, stop <-chan bool, workDone <-chan bool, options options.RunOptions) {
				for i := 0; i < options.Concurrency; i++ {
					doWork <- true
				}
				for {
					select {
					case <-stop:
						return
					case <-workDone:
						doWork <- true
					}
				}
			}

			return &api.Trigger{
					Trigger:     trigger,
					Description: "Makes requests from a set of virtual users specified by --concurrency",
					// The rate function used by the `users` mode, is actually dependent
					// on the number of virtual users specified in the `--concurrency` flag.
					// This flag is not required for the `chart` command, which uses the `DryRun`
					// function, so its not possible to provide an accurate rate function here.
					DryRun: func(t time.Time) int { return 1 },
				},
				nil
		},
	}
}
