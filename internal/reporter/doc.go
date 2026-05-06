// Package reporter provides formatted output of drift scan results.
//
// A Reporter summarises the outcome of a watcher scan cycle, listing each
// monitored file alongside its drift status. Two output formats are supported:
//
//	- FormatText: human-readable, suitable for log files and terminals.
//	- FormatJSON: machine-readable, suitable for log aggregators and dashboards.
//
// Basic usage:
//
//	r := reporter.New(os.Stdout, reporter.FormatText)
//	if err := r.Report(results); err != nil {
//		log.Fatal(err)
//	}
package reporter
