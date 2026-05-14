package rollup

import (
	"fmt"
	"strings"

	"github.com/example/driftwatch/internal/watcher"
)

// Summary holds aggregated statistics for a flush batch.
type Summary struct {
	Total   int
	Drifted int
	Paths   []string
}

// Summarise builds a Summary from a batch of DriftResults.
func Summarise(results []watcher.DriftResult) Summary {
	s := Summary{Total: len(results)}
	for _, r := range results {
		if r.Drifted {
			s.Drifted++
			s.Paths = append(s.Paths, r.Path)
		}
	}
	return s
}

// String returns a human-readable one-liner for the summary.
func (s Summary) String() string {
	if s.Drifted == 0 {
		return fmt.Sprintf("rollup: %d files checked, no drift detected", s.Total)
	}
	return fmt.Sprintf("rollup: %d/%d files drifted — %s",
		s.Drifted, s.Total, strings.Join(s.Paths, ", "))
}
