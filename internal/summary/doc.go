// Package summary provides a periodic drift summary reporter for driftwatch.
//
// A Summariser collects [watcher.DriftResult] values via Add and flushes
// aggregated [Report] values to a writer on a configurable time window.
//
// Output can be formatted as plain text (FormatText) or newline-delimited
// JSON (FormatJSON), making it easy to pipe into log aggregators or
// monitoring systems.
//
// Typical usage:
//
//	s := summary.New(5*time.Minute, summary.FormatJSON, logWriter)
//	go s.Run(ctx)
//	// feed results from the pipeline:
//	s.Add(result)
package summary
