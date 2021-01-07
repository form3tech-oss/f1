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
		Name:        "users <scenario>",
		Description: "triggers test iterations from a static set of virtual users controlled by the --concurrency flag",
		Flags:       flags,
		New: func(params *pflag.FlagSet) (*api.Trigger, error) {
			trigger := func(workTriggered chan<- bool, stop <-chan bool, workDone <-chan bool, options options.RunOptions) {
				DoWork(workTriggered, stop, workDone, options.Concurrency, options.MaxDuration)

				//for i := 0; i < options.Concurrency; i++ {
				//	workTriggered <- true
				//}
				//for {
				//	select {
				//	case <-stop:
				//		return
				//	case <-workDone:
				//		workTriggered <- true
				//	}
				//}
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

func DoWork(workTriggered chan<- bool, stop <-chan bool, workDone <-chan bool, concurrency int, duration time.Duration) {
	if concurrency == 0 || duration == 0 {
		return
	}

	for i := 0; i < concurrency; i++ {
		workTriggered <- true
	}

	totalDurationTicker := time.NewTicker(duration)

	for {
		select {
		case <-stop:
			totalDurationTicker.Stop()
			return
		case <-workDone:
			workTriggered <- true
		case <-totalDurationTicker.C:
			totalDurationTicker.Stop()
			return
		}
	}
}
