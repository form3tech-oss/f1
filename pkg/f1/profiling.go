package f1

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
)

type profiling struct {
	cpuProfileFile     *os.File
	cpuProfileFileName string
	memProfileFileName string
}

func (p *profiling) start() error {
	if len(p.cpuProfileFileName) == 0 {
		return nil
	}

	var err error
	p.cpuProfileFile, err = os.Create(p.cpuProfileFileName)
	if err != nil {
		return fmt.Errorf("creating cpuprofile file '%s': %w", p.cpuProfileFileName, err)
	}

	if err := pprof.StartCPUProfile(p.cpuProfileFile); err != nil {
		return fmt.Errorf("starting cpu profile: %w", err)
	}

	return nil
}

func (p *profiling) stop() error {
	if p.cpuProfileFile != nil {
		pprof.StopCPUProfile()
		if err := p.cpuProfileFile.Close(); err != nil {
			return fmt.Errorf("closing cpu profile file: %w", err)
		}
	}

	if p.memProfileFileName != "" {
		f, err := os.Create(p.memProfileFileName)
		if err != nil {
			return fmt.Errorf("creating memprofile file '%s': %w", p.memProfileFileName, err)
		}
		defer f.Close()

		runtime.GC() // get up-to-date statistics

		if err := pprof.WriteHeapProfile(f); err != nil {
			return fmt.Errorf("writing mem profile: %w", err)
		}
	}

	return nil
}
