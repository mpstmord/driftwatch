package reporter

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/driftwatch/internal/watcher"
)

// Format defines the output format for drift reports.
type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

// Reporter writes drift summaries to a sink.
type Reporter struct {
	sink   io.Writer
	format Format
}

// New creates a Reporter writing to the given sink in the specified format.
// If sink is nil, os.Stdout is used. If format is empty, FormatText is used.
func New(sink io.Writer, format Format) *Reporter {
	if sink == nil {
		sink = os.Stdout
	}
	if format == "" {
		format = FormatText
	}
	return &Reporter{sink: sink, format: format}
}

// Report writes a summary of the provided drift results to the sink.
func (r *Reporter) Report(results []watcher.DriftResult) error {
	switch r.format {
	case FormatJSON:
		return r.writeJSON(results)
	default:
		return r.writeText(results)
	}
}

func (r *Reporter) writeText(results []watcher.DriftResult) error {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	drifted := 0
	for _, res := range results {
		if res.Drifted {
			drifted++
		}
	}
	_, err := fmt.Fprintf(r.sink, "[%s] drift report: %d/%d files drifted\n", timestamp, drifted, len(results))
	if err != nil {
		return err
	}
	for _, res := range results {
		status := "ok"
		if res.Drifted {
			status = "DRIFTED"
		}
		_, err := fmt.Fprintf(r.sink, "  [%s] %s\n", status, res.Path)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Reporter) writeJSON(results []watcher.DriftResult) error {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	_, err := fmt.Fprintf(r.sink, `{"timestamp":%q,"results":[`, timestamp)
	if err != nil {
		return err
	}
	for i, res := range results {
		if i > 0 {
			_, err = fmt.Fprint(r.sink, ",")
			if err != nil {
				return err
			}
		}
		_, err = fmt.Fprintf(r.sink, `{"path":%q,"drifted":%v}`, res.Path, res.Drifted)
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprintln(r.sink, "]}")
	return err
}
