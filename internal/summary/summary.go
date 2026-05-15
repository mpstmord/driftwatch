// Package summary provides a periodic drift summary reporter that
// aggregates watcher results over a configurable window and emits
// human-readable or JSON summaries to a writer.
package summary

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/driftwatch/internal/watcher"
)

// Format controls the output format of the summary.
type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

// Report holds aggregated drift counts for a single flush window.
type Report struct {
	WindowStart time.Time `json:"window_start"`
	WindowEnd   time.Time `json:"window_end"`
	Total       int       `json:"total"`
	Drifted     int       `json:"drifted"`
	Clean       int       `json:"clean"`
	Paths       []string  `json:"drifted_paths,omitempty"`
}

// Summariser collects drift results and periodically writes a report.
type Summariser struct {
	mu      sync.Mutex
	results []watcher.DriftResult
	window  time.Duration
	format  Format
	sink    io.Writer
}

// New creates a Summariser that flushes every window duration.
// If sink is nil it defaults to os.Stdout.
func New(window time.Duration, format Format, sink io.Writer) *Summariser {
	if sink == nil {
		sink = os.Stdout
	}
	if format == "" {
		format = FormatText
	}
	return &Summariser{window: window, format: format, sink: sink}
}

// Add records a drift result for inclusion in the next report.
func (s *Summariser) Add(r watcher.DriftResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.results = append(s.results, r)
}

// Run starts the flush loop, writing a report every window until ctx is cancelled.
func (s *Summariser) Run(ctx context.Context) {
	ticker := time.NewTicker(s.window)
	defer ticker.Stop()
	start := time.Now()
	for {
		select {
		case <-ctx.Done():
			s.flush(start, time.Now())
			return
		case t := <-ticker.C:
			s.flush(start, t)
			start = t
		}
	}
}

func (s *Summariser) flush(start, end time.Time) {
	s.mu.Lock()
	results := s.results
	s.results = nil
	s.mu.Unlock()

	rep := build(results, start, end)
	s.write(rep)
}

func build(results []watcher.DriftResult, start, end time.Time) Report {
	rep := Report{WindowStart: start, WindowEnd: end, Total: len(results)}
	seen := map[string]bool{}
	for _, r := range results {
		if r.Drifted {
			rep.Drifted++
			if !seen[r.Path] {
				seen[r.Path] = true
				rep.Paths = append(rep.Paths, r.Path)
			}
		} else {
			rep.Clean++
		}
	}
	return rep
}

func (s *Summariser) write(rep Report) {
	switch s.format {
	case FormatJSON:
		b, _ := json.Marshal(rep)
		fmt.Fprintf(s.sink, "%s\n", b)
	default:
		fmt.Fprintf(s.sink, "[summary] %s — %s | total=%d drifted=%d clean=%d paths=%v\n",
			rep.WindowStart.Format(time.RFC3339),
			rep.WindowEnd.Format(time.RFC3339),
			rep.Total, rep.Drifted, rep.Clean, rep.Paths)
	}
}
