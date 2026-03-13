package run

import (
	"fmt"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/form3tech-oss/f1/v3/internal/triggerflags"
)

func flagGroupOrder() []string {
	return []string{
		"Output", "Duration & limits", "Concurrency", "Failure handling", "Shutdown",
		"Trigger options", "Help",
	}
}

func commonFlagGroups() map[string]string {
	return map[string]string{
		triggerflags.FlagVerbose:                  "Output",
		triggerflags.FlagMaxDuration:              "Duration & limits",
		triggerflags.FlagMaxIterations:            "Duration & limits",
		triggerflags.FlagConcurrency:              "Concurrency",
		triggerflags.FlagMaxFailures:              "Failure handling",
		triggerflags.FlagMaxFailuresRate:          "Failure handling",
		triggerflags.FlagIgnoreDropped:            "Failure handling",
		triggerflags.FlagWaitForCompletionTimeout: "Shutdown",
		"help": "Help",
	}
}

var registerHelpTemplateFunc = sync.OnceFunc(func() {
	cobra.AddTemplateFunc("groupedFlagUsages", groupedFlagUsages)
})

func groupedFlagUsages(cmd *cobra.Command) string {
	if cmd == nil || !cmd.HasAvailableLocalFlags() {
		return ""
	}
	fs := cmd.LocalFlags()
	groups := commonFlagGroups()

	var out strings.Builder
	for _, groupName := range flagGroupOrder() {
		groupFS := pflag.NewFlagSet("", pflag.ContinueOnError)
		groupFS.SortFlags = false
		fs.VisitAll(func(flag *pflag.Flag) {
			if flag.Hidden {
				return
			}
			g := groups[flag.Name]
			if g == "" {
				g = "Trigger options"
			}
			if g == groupName {
				groupFS.AddFlag(flag)
			}
		})
		if groupFS.HasFlags() {
			fmt.Fprintf(&out, "\n%s:\n%s", groupName, groupFS.FlagUsages())
		}
	}
	return strings.TrimSpace(out.String()) + "\n"
}
