package f1

import (
	"os"
	"path"
	"runtime"
	"runtime/pprof"

	"github.com/form3tech-oss/f1/v2/internal/support/errorh"

	"github.com/form3tech-oss/f1/v2/internal/fluentd_hook"

	"github.com/form3tech-oss/f1/v2/pkg/f1/scenarios"

	"github.com/form3tech-oss/f1/v2/internal/chart"
	"github.com/form3tech-oss/f1/v2/internal/run"
	"github.com/form3tech-oss/f1/v2/internal/trigger"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func buildRootCmd(s *scenarios.Scenarios, p *profiling) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:              getCmdName(),
		Short:            "F1 load testing tool",
		PersistentPreRun: startProfiling(p),
	}
	builders := trigger.GetBuilders()
	rootCmd.PersistentFlags().String("cpuprofile", "", "write cpu profile to `file`")
	rootCmd.PersistentFlags().String("memprofile", "", "write memory profile to `file`")
	rootCmd.AddCommand(run.Cmd(s, builders, fluentd_hook.AddFluentdLoggingHook))
	rootCmd.AddCommand(chart.Cmd(builders))
	rootCmd.AddCommand(scenarios.Cmd(s))
	rootCmd.AddCommand(completionsCmd(s, p))
	return rootCmd
}

func startProfiling(p *profiling) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		if file, ok := cmd.Flags().GetString("cpuprofile"); ok == nil && file != "" {
			var err error
			p.cpuProfile, err = os.Create(file)
			if err != nil {
				log.Fatal("could not create CPU profile: ", err)
			}
			if err := pprof.StartCPUProfile(p.cpuProfile); err != nil {
				log.Fatal("could not start CPU profile: ", err)
			}
		}
		if file, ok := cmd.Flags().GetString("memprofile"); ok == nil && file != "" {
			p.memProfile = file
		}
	}
}

func writeProfiles(p *profiling) {
	if p.cpuProfile != nil {
		pprof.StopCPUProfile()
		errorh.Print(p.cpuProfile.Close(), "error closing cpu profile")
	}
	if p.memProfile != "" {
		f, err := os.Create(p.memProfile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer errorh.SafeClose(f)
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}

func getCmdName() string {
	return path.Base(os.Args[0])
}
