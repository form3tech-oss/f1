package run

import (
	"context"
	"errors"
	"fmt"
	"io"
	stdlog "log"
	"os"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/sirupsen/logrus"

	"github.com/form3tech-oss/f1/v2/internal/console"
	"github.com/form3tech-oss/f1/v2/internal/envsettings"
	"github.com/form3tech-oss/f1/v2/internal/logging"
	"github.com/form3tech-oss/f1/v2/internal/metrics"
	"github.com/form3tech-oss/f1/v2/internal/options"
	"github.com/form3tech-oss/f1/v2/internal/progress"
	"github.com/form3tech-oss/f1/v2/internal/raterun"
	"github.com/form3tech-oss/f1/v2/internal/run/templates"
	"github.com/form3tech-oss/f1/v2/internal/trigger/api"
	"github.com/form3tech-oss/f1/v2/internal/workers"
	"github.com/form3tech-oss/f1/v2/internal/xcontext"
	"github.com/form3tech-oss/f1/v2/pkg/f1/scenarios"
)

const (
	nextIterationWindow    = 10 * time.Millisecond
	metricsRefreshInterval = 5 * time.Second
)

type Run struct {
	progressRunner  *raterun.Runner
	progressStats   *progress.Stats
	metrics         *metrics.Metrics
	templates       *templates.Templates
	activeScenario  *workers.ActiveScenario
	trigger         *api.Trigger
	pusher          *push.Pusher
	printer         *console.Printer
	RateDescription string
	Settings        envsettings.Settings
	Options         options.RunOptions
	result          Result
	notifyDropped   sync.Once
}

func NewRun(
	options options.RunOptions,
	t *api.Trigger,
	settings envsettings.Settings,
	metricsInstane *metrics.Metrics,
	printer *console.Printer,
) (*Run, error) {
	run := Run{
		Options:         options,
		Settings:        settings,
		RateDescription: t.Description,
		trigger:         t,
		metrics:         metricsInstane,
		printer:         printer,
		progressStats:   &progress.Stats{},
	}

	run.templates = templates.Parse(templates.RenderTermColors)
	run.result = NewResult(options, run.templates, run.progressStats)

	if run.Settings.Prometheus.PushGateway != "" {
		run.pusher = push.New(settings.Prometheus.PushGateway, "f1-"+options.Scenario).
			Gatherer(run.metrics.Registry)

		if run.Settings.Prometheus.Namespace != "" {
			run.pusher = run.pusher.Grouping("namespace", run.Settings.Prometheus.Namespace)
		}

		if run.Settings.Prometheus.LabelID != "" {
			run.pusher = run.pusher.Grouping("id", run.Settings.Prometheus.LabelID)
		}
	}
	if run.Options.RegisterLogHookFunc == nil {
		run.Options.RegisterLogHookFunc = logging.NoneRegisterLogHookFunc
	}

	progressRunner, err := raterun.New(func(rate time.Duration) {
		run.result.SnapshotProgress(rate)
		run.printer.Println(run.result.Progress())
		if run.result.HasDroppedIterations() {
			run.notifyDropped.Do(func() {
				logrus.Warn("Dropping requests as workers are too busy. Considering increasing `--concurrency` argument")
			})
		}
	}, []raterun.Schedule{
		{StartDelay: 0, Frequency: time.Second},
		{StartDelay: time.Minute, Frequency: 10 * time.Second},
		{StartDelay: 5 * time.Minute, Frequency: 30 * time.Second},
		{StartDelay: 10 * time.Minute, Frequency: time.Minute},
	})
	if err != nil {
		return nil, fmt.Errorf("creating progress runner: %w", err)
	}

	run.progressRunner = progressRunner

	return &run, nil
}

func (r *Run) Do(ctx context.Context, s *scenarios.Scenarios) (*Result, error) {
	r.printer.Print(r.templates.Start(templates.StartData{
		Scenario:        r.Options.Scenario,
		MaxDuration:     r.Options.MaxDuration,
		MaxIterations:   r.Options.MaxIterations,
		RateDescription: r.RateDescription,
	}))

	defer r.printSummary()
	defer r.printLogOnFailure()

	if err := r.configureLogging(); err != nil {
		return nil, fmt.Errorf("configure logging: %w", err)
	}

	r.metrics.Reset()

	scenario := s.GetScenario(r.Options.Scenario)
	if scenario == nil {
		return nil, fmt.Errorf("scenario not defined: %s", r.Options.Scenario)
	}
	r.activeScenario = workers.NewActiveScenario(scenario, r.metrics, r.progressStats)
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

	return &r.result, nil
}

func (r *Run) reportSetupFailure(ctx context.Context) *Result {
	r.fail("setup failed")
	r.pushMetrics(ctx)
	r.printer.Println(r.result.Setup())
	return &r.result
}

func (r *Run) teardownActiveScenario(ctx context.Context) {
	r.activeScenario.Teardown()
	if r.activeScenario.TeardownFailed() {
		r.fail("teardown failed")
	}
	r.pushMetrics(ctx)
	r.printer.Println(r.result.Teardown())
}

func (r *Run) configureLogging() error {
	err := r.Options.RegisterLogHookFunc(r.Options.Scenario)
	if err != nil {
		return fmt.Errorf("calling register log hook func: %w", err)
	}

	if !r.Options.Verbose {
		r.result.LogFile = redirectLoggingToFile(r.Options.Scenario, r.Settings.LogFilePath, r.printer.Writer)
		welcomeMessage := r.templates.Start(templates.StartData{
			Scenario:        r.Options.Scenario,
			MaxDuration:     r.Options.MaxDuration,
			MaxIterations:   r.Options.MaxIterations,
			RateDescription: r.RateDescription,
		})

		logrus.Info(welcomeMessage)
		r.printer.Printf("Saving logs to %s\n\n", r.result.LogFile)
	}

	return nil
}

func (r *Run) printSummary() {
	summary := r.result.String()
	r.printer.Println(summary)
	if !r.Options.Verbose {
		logrus.Info(summary)
		logrus.StandardLogger().SetOutput(r.printer.Writer)
		stdlog.SetOutput(r.printer.Writer)
	}
}

func (r *Run) run(ctx context.Context) {
	// if the trigger has a limited duration, restrict the run to that duration.
	duration := r.Options.MaxDuration
	if r.trigger.Duration > 0 && r.trigger.Duration < r.Options.MaxDuration {
		duration = r.trigger.Duration
	}

	// Cancel work slightly before end of duration to avoid starting a new iteration
	r.result.RecordStarted()
	defer r.result.RecordTestFinished()

	triggerCtx, triggerCancel := context.WithTimeout(ctx, duration-nextIterationWindow)
	defer triggerCancel()

	poolManager := workers.New(r.Options.MaxIterations, r.activeScenario)
	r.trigger.Trigger(triggerCtx, poolManager, r.Options)

	select {
	case <-ctx.Done():
		r.printer.Println(r.result.Interrupted())
		r.progressRunner.Restart()
		<-poolManager.WaitForCompletion()
	case <-triggerCtx.Done():
		if triggerCtx.Err() == context.DeadlineExceeded {
			r.printer.Println(r.result.MaxDurationElapsed())
		}
		<-poolManager.WaitForCompletion()
	case <-poolManager.WaitForCompletion():
		if poolManager.MaxIterationsReached() {
			r.printer.Println(r.result.MaxIterationsReached())
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
		logrus.Errorf("unable to push metrics to prometheus: %v", err)
	}
}

func (r *Run) printLogOnFailure() {
	if r.Options.Verbose || !r.Options.VerboseFail || !r.result.Failed() {
		return
	}

	if err := r.printResultLogs(); err != nil {
		logrus.WithError(err).Error("error printing logs")
	}
}

func (r *Run) printResultLogs() error {
	fd, err := os.Open(r.result.LogFile)
	if err != nil {
		return fmt.Errorf("opening log file: %w", err)
	}
	defer func() {
		if fd == nil {
			return
		}
		if err := fd.Close(); err != nil {
			logrus.WithError(err).Error("error closing log file")
		}
	}()

	if fd != nil {
		if _, err := io.Copy(r.printer.Writer, fd); err != nil {
			return fmt.Errorf("printing logs: %w", err)
		}
	}

	return nil
}
