package run

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/push"

	"github.com/form3tech-oss/f1/v2/internal/envsettings"
	"github.com/form3tech-oss/f1/v2/internal/log"
	"github.com/form3tech-oss/f1/v2/internal/logutils"
	"github.com/form3tech-oss/f1/v2/internal/metrics"
	"github.com/form3tech-oss/f1/v2/internal/options"
	"github.com/form3tech-oss/f1/v2/internal/progress"
	"github.com/form3tech-oss/f1/v2/internal/raterun"
	"github.com/form3tech-oss/f1/v2/internal/run/views"
	"github.com/form3tech-oss/f1/v2/internal/trigger/api"
	"github.com/form3tech-oss/f1/v2/internal/ui"
	"github.com/form3tech-oss/f1/v2/internal/workers"
	"github.com/form3tech-oss/f1/v2/internal/xcontext"
	"github.com/form3tech-oss/f1/v2/pkg/f1/scenarios"
)

const (
	nextIterationWindow    = 10 * time.Millisecond
	metricsRefreshInterval = 5 * time.Second
)

type Run struct {
	pusher         *push.Pusher
	progressRunner *raterun.Runner
	metrics        *metrics.Metrics
	views          *views.Views
	activeScenario *workers.ActiveScenario
	trigger        *api.Trigger
	output         *ui.Output
	scenarioLogger *ScenarioLogger
	result         *Result
	options        options.RunOptions
}

func NewRun(
	options options.RunOptions,
	scenarios *scenarios.Scenarios,
	trigger *api.Trigger,
	settings envsettings.Settings,
	metricsInstance *metrics.Metrics,
	parentOutput *ui.Output,
) (*Run, error) {
	progressStats := &progress.Stats{}
	viewsInstance := views.New()

	scenario := scenarios.GetScenario(options.Scenario)
	if scenario == nil {
		return nil, fmt.Errorf("scenario not defined: %s", options.Scenario)
	}

	result := NewResult(options, viewsInstance, progressStats)

	outputer := ui.NewOutput(
		parentOutput.Logger.With(log.ScenarioAttr(scenario.Name)),
		parentOutput.Printer,
		parentOutput.Interactive,
		options.LogToFile(),
	)

	scenarioLogger := NewScenarioLogger(outputer)
	result.LogFilePath = scenarioLogger.Open(
		LogFilePathOrDefault(settings.Log.FilePath, scenario.Name),
		logutils.NewLogConfigFromSettings(settings),
		scenario.Name,
		options.LogToFile(),
	)

	progressRunner, err := newProgressRunner(result, outputer)
	if err != nil {
		return nil, fmt.Errorf("creating progress runner: %w", err)
	}

	activeScenario := workers.NewActiveScenario(
		scenario,
		metricsInstance,
		progressStats,
		scenarioLogger.Logger,
		log.NewSlogLogrusLogger(scenarioLogger.Logger),
	)

	pusher := newMetricsPusher(settings, scenario.Name, metricsInstance)

	return &Run{
		options:        options,
		trigger:        trigger,
		metrics:        metricsInstance,
		views:          viewsInstance,
		result:         result,
		pusher:         pusher,
		output:         outputer,
		progressRunner: progressRunner,
		activeScenario: activeScenario,
		scenarioLogger: scenarioLogger,
	}, nil
}

func newMetricsPusher(
	settings envsettings.Settings,
	scenarioName string,
	metricsInstance *metrics.Metrics,
) *push.Pusher {
	if settings.Prometheus.PushGateway == "" {
		return nil
	}

	pusher := push.New(settings.Prometheus.PushGateway, "f1-"+scenarioName).
		Gatherer(metricsInstance.Registry)

	if settings.Prometheus.Namespace != "" {
		pusher = pusher.Grouping("namespace", settings.Prometheus.Namespace)
	}

	if settings.Prometheus.LabelID != "" {
		pusher = pusher.Grouping("id", settings.Prometheus.LabelID)
	}

	return pusher
}

func newProgressRunner(result *Result, output *ui.Output) (*raterun.Runner, error) {
	notifyDropped := sync.Once{}

	r, err := raterun.New(func(rate time.Duration) {
		result.SnapshotProgress(rate)
		output.Display(result.Progress())
		if result.HasDroppedIterations() {
			notifyDropped.Do(func() {
				output.Display(ui.WarningMessage{
					Message: "Dropping requests as workers are too busy. " +
						"Considering increasing `--concurrency` argument",
				})
			})
		}
	}, []raterun.Schedule{
		{StartDelay: 0, Frequency: time.Second},
		{StartDelay: time.Minute, Frequency: 10 * time.Second},
		{StartDelay: 5 * time.Minute, Frequency: 30 * time.Second},
		{StartDelay: 10 * time.Minute, Frequency: time.Minute},
	})
	if err != nil {
		return nil, fmt.Errorf("new progress runner: %w", err)
	}

	return r, nil
}

func (r *Run) Do(ctx context.Context) (*Result, error) {
	defer r.scenarioLogger.Close()

	welcomeMessage := r.views.Start(views.StartData{
		Scenario:        r.options.Scenario,
		MaxDuration:     r.options.MaxDuration,
		MaxIterations:   r.options.MaxIterations,
		RateDescription: r.trigger.Description,
	})

	r.output.Display(welcomeMessage)

	defer r.printSummary()

	r.metrics.Reset()

	r.activeScenario.Setup()

	r.pushMetrics(ctx)

	// run teardown even if the context is cancelled
	teardownContext := xcontext.Detach(ctx)
	defer r.teardownActiveScenario(teardownContext)

	if r.activeScenario.Failed() {
		return r.reportSetupFailure(ctx), nil
	}

	// set initial started timestamp so that the progress trackers work
	r.result.RecordStarted()

	metricsCloseCh := make(chan struct{})
	go func() {
		t := time.NewTicker(metricsRefreshInterval)
		defer t.Stop()

		for {
			select {
			case <-t.C:
				r.pushMetrics(ctx)
			case <-ctx.Done():
				return
			case <-metricsCloseCh:
				return
			}
		}
	}()

	r.progressRunner.Start(ctx)

	r.run(ctx)

	r.progressRunner.Stop()
	close(metricsCloseCh)
	r.result.GetTotals()

	return r.result, nil
}

func (r *Run) reportSetupFailure(ctx context.Context) *Result {
	r.fail("setup failed")
	r.pushMetrics(ctx)
	r.output.Display(r.result.Setup())
	return r.result
}

func (r *Run) teardownActiveScenario(ctx context.Context) {
	r.activeScenario.Teardown()
	if r.activeScenario.TeardownFailed() {
		r.fail("teardown failed")
	}
	r.pushMetrics(ctx)
	r.output.Display(r.result.Teardown())
}

func (r *Run) printSummary() {
	r.output.Display(r.result.Summary())
}

func (r *Run) run(ctx context.Context) {
	// if the trigger has a limited duration, restrict the run to that duration.
	duration := r.options.MaxDuration
	if r.trigger.Duration > 0 && r.trigger.Duration < r.options.MaxDuration {
		duration = r.trigger.Duration
	}

	// Cancel work slightly before end of duration to avoid starting a new iteration
	r.result.RecordStarted()
	defer r.result.RecordTestFinished()

	triggerCtx, triggerCancel := context.WithTimeout(ctx, duration-nextIterationWindow)
	defer triggerCancel()

	poolManager := workers.New(r.options.MaxIterations, r.activeScenario)
	r.trigger.Trigger(triggerCtx, r.output, poolManager, r.options)

	select {
	case <-ctx.Done():
		r.output.Display(r.result.Interrupted())
		r.progressRunner.Restart()
		<-poolManager.WaitForCompletion()
	case <-triggerCtx.Done():
		if triggerCtx.Err() == context.DeadlineExceeded {
			r.output.Display(r.result.MaxDurationElapsed())
		}
		<-poolManager.WaitForCompletion()
	case <-poolManager.WaitForCompletion():
		if poolManager.MaxIterationsReached() {
			r.output.Display(r.result.MaxIterationsReached())
		}
	}
}

func (r *Run) fail(message string) {
	r.result.AddError(errors.New(message))
}

func (r *Run) pushMetrics(ctx context.Context) {
	if r.pusher == nil {
		return
	}
	err := r.pusher.PushContext(ctx)
	if err != nil {
		r.output.Display(ui.ErrorMessage{
			Message: "unable to push metrics to prometheus",
			Error:   err,
		})
	}
}
