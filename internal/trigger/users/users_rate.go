package users

import (
	"time"

	"github.com/form3tech-oss/f1/v2/internal/options"
	"github.com/form3tech-oss/f1/v2/internal/trigger/api"
	"github.com/spf13/pflag"
)

func UsersRate() api.Builder {
	flags := pflag.NewFlagSet("users", pflag.ContinueOnError)

	return api.Builder{
		Name:        "users <scenario>",
		Description: "triggers test iterations from a static set of users controlled by the --concurrency flag",
		Flags:       flags,
		New: func(params *pflag.FlagSet) (*api.Trigger, error) {
			trigger := func(workTriggered chan<- bool, stop <-chan bool, workDone <-chan bool, options options.RunOptions) {
				doWork := NewWorker(options.Concurrency)
				doWork(workTriggered, stop, workDone, options)
			}

			return &api.Trigger{
					Trigger:     trigger,
					Description: "Makes requests from a set of users specified by --concurrency",
					// The rate function used by the `users` mode, is actually dependent
					// on the number of users specified in the `--concurrency` flag.
					// This flag is not required for the `chart` command, which uses the `DryRun`
					// function, so its not possible to provide an accurate rate function here.
					DryRun: func(t time.Time) int { return 1 },
				},
				nil
		},
	}
}

func NewWorker(concurrency int) api.WorkTriggerer {
	return func(workTriggered chan<- bool, stop <-chan bool, workDone <-chan bool, options options.RunOptions) {
		for i := 0; i < concurrency; i++ {
			workTriggered <- true
		}

		for {
			select {
			case <-stop:
				return
			case <-workDone:
				workTriggered <- true
			}
		}
	}
}
