// Package replay provides facilities for replaying previously recorded drift
// events from an [eventlog.EventLog].
//
// This is useful in two scenarios:
//
//  1. Diagnostics — an operator can replay the event history to understand
//     what files drifted and when, without waiting for the next scan cycle.
//
//  2. Catch-up — after a period of downtime or a notifier outage, the replay
//     package lets downstream consumers reprocess missed events starting from
//     a specific point in time via [Options.Since].
//
// # Basic usage
//
//	r := replay.New(myEventLog, replay.Options{Since: lastProcessedAt})
//	err := r.Run(ctx, func(result watcher.DriftResult) {
//	    // handle replayed event
//	})
package replay
