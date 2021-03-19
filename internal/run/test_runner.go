package run

import (
	"fmt"
	"io"
	stdlog "log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"text/template"
	"time"

	"github.com/form3tech-oss/f1/v2/internal/logging"
	"github.com/form3tech-oss/f1/v2/internal/options"
	"github.com/form3tech-oss/f1/v2/pkg/f1/scenarios"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/form3tech-oss/f1/v2/internal/trace"

	"github.com/form3tech-oss/f1/v2/internal/raterun"

	"github.com/form3tech-oss/f1/v2/internal/trigger/api"

	"github.com/aholic/ggtimer"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"

	"github.com/form3tech-oss/f1/v2/internal/metrics"
)

const NextIterationWindow = 10 * time.Millisecond
const IterationStage = "iteration"

func NewRun(options options.RunOptions, t *api.Trigger) (*Run, error) {
	run := Run{
		Options:         options,
		RateDescription: t.Description,
		trigger:         t,
	}
	prometheusUrl := os.Getenv("PROMETHEUS_PUSH_GATEWAY")
	if prometheusUrl != "" {
		run.pusher = push.New(prometheusUrl, "f1-"+options.Scenario).Gatherer(prometheus.DefaultGatherer)
	}
	if run.Options.RegisterLogHookFunc == nil {
		run.Options.RegisterLogHookFunc = logging.NoneRegisterLogHookFunc
	}
	run.result.IgnoreDropped = options.IgnoreDropped

	progressRunner, _ := raterun.New(func(rate time.Duration, t time.Time) {
		run.gatherProgressMetrics(rate)
		fmt.Println(run.result.Progress())
	}, []raterun.Rate{
		{Start: time.Nanosecond, Rate: time.Second},
		{Start: time.Minute, Rate: time.Second * 10},
		{Start: time.Minute * 5, Rate: time.Minute / 2},
		{Start: time.Minute * 10, Rate: time.Minute},
	})
	run.progressRunner = progressRunner

	return &run, nil
}

type Run struct {
	Options         options.RunOptions
	busyWorkers     int32
	iteration       int32
	failures        int32
	result          RunResult
	activeScenario  *ActiveScenario
	interrupt       chan os.Signal
	trigger         *api.Trigger
	RateDescription string
	pusher          *push.Pusher
	notifyDropped   sync.Once
	progressRunner  raterun.Runner
}

var startTemplate = template.Must(template.New("result parse").
	Funcs(templateFunctions).
	Parse(`{u}{bold}{intensive_blue}F1 Load Tester{-}
Running {yellow}{{.Options.Scenario}}{-} scenario for {{if .Options.MaxIterations}}up to {{.Options.MaxIterations}} iterations or up to {{end}}{{duration .Options.MaxDuration}} at a rate of {{.RateDescription}}.
`))

func (r *Run) Do(s *scenarios.Scenarios) *RunResult {
	fmt.Print(renderTemplate(startTemplate, r))
	defer r.printSummary()
	defer r.printLogOnFailure()

	r.configureLogging()

	metrics.Instance().Reset()

	r.activeScenario = NewActiveScenario(s.GetScenario(r.Options.Scenario))
	r.pushMetrics()
	defer r.teardownActiveScenario()

	if r.activeScenario.t.Failed() {
		return r.reportSetupFailure()
	}

	// set initial started timestamp so that the progress trackers work
	r.result.RecordStarted()
	r.progressRunner.Run()
	metricsTick := ggtimer.NewTicker(5*time.Second, func(t time.Time) {
		r.pushMetrics()
	})

	r.run()

	r.progressRunner.Terminate()
	metricsTick.Close()
	r.gatherMetrics()

	return &r.result
}

func (r *Run) reportSetupFailure() *RunResult {
	r.fail("setup failed")
	r.pushMetrics()
	fmt.Println(r.result.Setup())
	return &r.result
}

func (r *Run) teardownActiveScenario() {
	r.activeScenario.Teardown()
	if r.activeScenario.t.TeardownFailed() {
		r.fail("teardown failed")
	}
	r.pushMetrics()
	fmt.Println(r.result.Teardown())
}

func (r *Run) configureLogging() {
	r.Options.RegisterLogHookFunc(r.Options.Scenario)
	if !r.Options.Verbose {
		r.result.LogFile = redirectLoggingToFile(r.Options.Scenario)
		welcomeMessage := renderTemplate(startTemplate, r)
		log.Info(welcomeMessage)
		fmt.Printf("Saving logs to %s\n\n", r.result.LogFile)
	}
}

func (r *Run) printSummary() {
	summary := r.result.String()
	fmt.Println(summary)
	if !r.Options.Verbose {
		log.Info(summary)
		log.StandardLogger().SetOutput(os.Stdout)
		stdlog.SetOutput(os.Stdout)
	}
}

func (r *Run) run() {
	// handle ctrl-c interrupts
	r.interrupt = make(chan os.Signal, 1)
	signal.Notify(r.interrupt, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(r.interrupt)
	defer close(r.interrupt)

	// Build a worker group to process the iterations.
	workers := r.Options.Concurrency
	doWorkChannel := make(chan int32, workers)
	stopWorkers := make(chan struct{})

	wg := &sync.WaitGroup{}
	defer wg.Wait()

	r.busyWorkers = int32(0)
	workDone := make(chan bool, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go r.runWorker(doWorkChannel, stopWorkers, wg, fmt.Sprint(i), workDone)
	}

	// if the trigger has a limited duration, restrict the run to that duration.
	duration := r.Options.MaxDuration
	if r.trigger.Duration > 0 && r.trigger.Duration < r.Options.MaxDuration {
		duration = r.trigger.Duration
	}

	// Cancel work slightly before end of duration to avoid starting a new iteration
	durationElapsed := NewCancellableTimer(duration - NextIterationWindow)
	r.result.RecordStarted()
	defer r.result.RecordTestFinished()

	workTriggered := make(chan bool, workers)
	stopTrigger := make(chan bool, 1)
	go r.trigger.Trigger(workTriggered, stopTrigger, workDone, r.Options)

	// This blocks waiting for cancellable timer
	go func() {
		elapsed := <-durationElapsed.C
		trace.ReceivedFromChannel("C")
		if elapsed {
			fmt.Println(r.result.MaxDurationElapsed())
		}
		log.Info("Stopping worker")
		stopTrigger <- true
		close(stopWorkers)
		wg.Wait()
	}()

	// run more iterations on every tick, until duration has elapsed.
	for {
		trace.Event("Run loop ")
		select {
		case <-r.interrupt:
			fmt.Println(r.result.Interrupted())
			r.progressRunner.RestartRate()
			// stop listening to interrupts - second interrupt will terminate immediately
			signal.Stop(r.interrupt)
			durationElapsed.Cancel()
		case <-workTriggered:
			trace.ReceivedFromChannel("workTriggered")
			r.doWork(doWorkChannel, durationElapsed)
			trace.Event("Called do work")
		case <-stopWorkers:
			wg.Wait()
			return
		}
	}
}

func (r *Run) doWork(doWorkChannel chan<- int32, durationElapsed *CancellableTimer) {
	if atomic.LoadInt32(&r.busyWorkers) >= int32(r.Options.Concurrency) {
		r.activeScenario.RecordDroppedIteration()
		r.notifyDropped.Do(func() {
			// only log once.
			log.Warn("Dropping requests as workers are too busy. Considering increasing `--concurrency` argument")
		})
		return
	}
	iteration := atomic.AddInt32(&r.iteration, 1)
	if r.Options.MaxIterations > 0 && iteration > r.Options.MaxIterations {
		trace.Event("Max iterations exceeded Calling Cancel on iteration  '%v' .", iteration)
		if durationElapsed.Cancel() {
			fmt.Println(r.result.MaxIterationsReached())
		}
		trace.Event("Max iterations exceeded Called Cancel on iteration  '%v' .", iteration)
	} else if r.Options.MaxIterations <= 0 || iteration <= r.Options.MaxIterations {
		trace.Event("Within Max iterations So calling dowork() on iteration  '%v' .", iteration)
		doWorkChannel <- iteration
	}
}

func (r *Run) gatherMetrics() {
	m, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		r.result.AddError(errors.Wrap(err, "unable to gather metrics"))
	}
	for _, metric := range m {
		if *metric.Name == "form3_loadtest_iteration" {
			for _, m := range metric.Metric {
				result := "unknown"
				stage := IterationStage
				for _, label := range m.Label {
					if *label.Name == "result" {
						result = *label.Value
					}
					if *label.Name == "stage" {
						stage = *label.Value
					}
				}
				r.result.SetMetrics(result, stage, *m.Summary.SampleCount, m.Summary.Quantile)
			}
		}
	}
}
func (r *Run) gatherProgressMetrics(duration time.Duration) {
	m, err := metrics.Instance().ProgressRegistry.Gather()
	if err != nil {
		r.result.AddError(errors.Wrap(err, "unable to gather metrics"))
	}
	metrics.Instance().Progress.Reset()
	r.result.ClearProgressMetrics()
	for _, metric := range m {
		if *metric.Name == "form3_loadtest_iteration" {
			for _, m := range metric.Metric {
				result := "unknown"
				stage := IterationStage
				for _, label := range m.Label {
					if *label.Name == "result" {
						result = *label.Value
					}
					if *label.Name == "stage" {
						stage = *label.Value
					}
				}
				r.result.IncrementMetrics(duration, result, stage, *m.Summary.SampleCount, m.Summary.Quantile)
			}
		}
	}
}

func (r *Run) runWorker(input <-chan int32, stop <-chan struct{}, wg *sync.WaitGroup, worker string, workDone chan<- bool) {
	defer wg.Done()
	trace.Event("Started worker (%v)", worker)
	for {
		select {
		case <-stop:
			trace.Event("Stopping worker (%v)", worker)
			return
		case iteration := <-input:
			trace.Event("Received work (%v) from Channel 'doWork' iteration (%v)", worker, iteration)
			atomic.AddInt32(&r.busyWorkers, 1)

			scenario := r.activeScenario.scenario
			successful := r.activeScenario.Run(metrics.IterationResult, IterationStage, fmt.Sprintf("iteration %d", iteration), scenario.RunFn)
			if !successful {
				atomic.AddInt32(&r.failures, 1)
			}
			atomic.AddInt32(&r.busyWorkers, -1)

			// if we need to stop - no one is listening for workDone,
			// so it will block forever. check the stop channel to stop the worker
			select {
			case workDone <- true:
			case <-stop:
				return
			}

			trace.Event("Completed iteration (%v).", iteration)

		}
	}
}

func (r *Run) fail(message string) *RunResult {
	r.result.AddError(fmt.Errorf(message))
	return &r.result
}

func (r *Run) pushMetrics() {
	if r.pusher == nil {
		return
	}
	err := r.pusher.Push()
	if err != nil {
		log.Errorf("unable to push metrics to prometheus: %v", err)
	}
}

func (r *Run) printLogOnFailure() {
	if r.Options.Verbose || !r.Options.VerboseFail || !r.result.Failed() {
		return
	}

	if err := r.printResultLogs(); err != nil {
		log.WithError(err).Error("error printing logs")
	}
}

func (r *Run) printResultLogs() error {
	fd, err := os.Open(r.result.LogFile)
	if err != nil {
		return errors.Wrap(err, "error opening log file")
	}
	defer func() {
		if fd == nil {
			return
		}
		if err := fd.Close(); err != nil {
			log.WithError(err).Error("error closing log file")
		}
	}()

	if fd != nil {
		if _, err := io.Copy(os.Stdout, fd); err != nil {
			return errors.Wrap(err, "error printing printing logs")
		}
	}

	return nil
}
