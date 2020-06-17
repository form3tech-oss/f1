package f1

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/form3tech-oss/f1/internal/support/errorh"

	"github.com/form3tech-oss/f1/pkg/f1/fluentd_hook"

	"github.com/form3tech-oss/f1/pkg/f1/scenarios"

	"github.com/form3tech-oss/f1/pkg/f1/chart"
	"github.com/form3tech-oss/f1/pkg/f1/run"
	"github.com/form3tech-oss/f1/pkg/f1/trigger"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	cpuProfile *os.File
	memProfile string
)

func buildRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:              "f1",
		Short:            "F1 load testing tool",
		PersistentPreRun: startProfiling,
	}
	builders := trigger.GetBuilders()
	rootCmd.PersistentFlags().String("cpuprofile", "", "write cpu profile to `file`")
	rootCmd.PersistentFlags().String("memprofile", "", "write memory profile to `file`")
	rootCmd.AddCommand(run.Cmd(builders, fluentd_hook.AddFluentdLoggingHook))
	rootCmd.AddCommand(chart.Cmd(builders))
	rootCmd.AddCommand(scenarios.Cmd())
	rootCmd.AddCommand(completionsCmd())
	return rootCmd
}

func Execute() {
	if err := buildRootCmd().Execute(); err != nil {
		writeProfiles()
		fmt.Println(err)
		os.Exit(1)
	}
}

func startProfiling(cmd *cobra.Command, args []string) {
	if file, ok := cmd.Flags().GetString("cpuprofile"); ok == nil && file != "" {
		var err error
		cpuProfile, err = os.Create(file)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(cpuProfile); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
	}
	if file, ok := cmd.Flags().GetString("memprofile"); ok == nil && file != "" {
		memProfile = file
	}
}

func writeProfiles() {
	if cpuProfile != nil {
		pprof.StopCPUProfile()
		errorh.Print(cpuProfile.Close(), "error closing cpu profile")
	}
	if memProfile != "" {
		f, err := os.Create(memProfile)
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

func ExecuteWithArgs(args []string) error {
	rootCmd := buildRootCmd()
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	writeProfiles()
	return err
}
