package users

import (
	"context"
	"time"

	"github.com/spf13/pflag"

	"github.com/form3tech-oss/f1/v2/internal/options"
	"github.com/form3tech-oss/f1/v2/internal/trigger/api"
	"github.com/form3tech-oss/f1/v2/internal/ui"
	"github.com/form3tech-oss/f1/v2/internal/workers"
)

func Rate() api.Builder {
	flags := pflag.NewFlagSet("users", pflag.ContinueOnError)

	return api.Builder{
		Name:        "users <scenario>",
		Description: "triggers test iterations from a static set of users controlled by the --concurrency flag",
		Flags:       flags,
		New: func(*pflag.FlagSet) (*api.Trigger, error) {
			trigger := func(
				ctx context.Context,
				output *ui.Output,
				workers *workers.PoolManager,
				options options.RunOptions,
			) {
				doWork := NewWorker(options.Concurrency)
				doWork(ctx, output, workers, options)
			}

			return &api.Trigger{
					Trigger:     trigger,
					Description: "Makes requests from a set of users specified by --concurrency",
					// The rate function used by the `users` mode, is actually dependent
					// on the number of users specified in the `--concurrency` flag.
					Rate: func(time.Time) int { return 1 },
				},
				nil
		},
	}
}

func NewWorker(concurrency int) api.WorkTriggerer {
	return func(ctx context.Context, _ *ui.Output, workers *workers.PoolManager, _ options.RunOptions) {
		pool := workers.NewContinuousPool(concurrency)
		pool.Start(ctx)
		<-workers.WaitForCompletion()
	}
}
