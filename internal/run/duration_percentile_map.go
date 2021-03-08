package run

import (
	"fmt"
	"sort"
	"time"
)

type DurationPercentileMap map[float64]time.Duration

func (m *DurationPercentileMap) String() string {
	s := ""
	keys := make([]float64, 0, len(*m))
	for k := range *m {
		keys = append(keys, k)
	}
	sort.Float64s(keys)
	for _, percentile := range keys {
		s = fmt.Sprintf("%s p(%2.2f): %v,", s, percentile*100, (*m)[percentile])
	}
	return s
}

func (m *DurationPercentileMap) Get(pc float64) string {
	return fmt.Sprintf("%v", (*m)[pc])
}
