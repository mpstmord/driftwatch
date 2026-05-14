// Package rollup provides a time-windowed aggregator for drift events.
//
// Instead of emitting one alert per drifted file, rollup collects
// DriftResult values over a configurable window and delivers a single
// batch to a handler function when the window expires.
//
// Typical usage:
//
//	r := rollup.New(30*time.Second, func(batch []watcher.DriftResult) {
//		summary := rollup.Summarise(batch)
//		log.Println(summary)
//	})
//	go r.Run(ctx)
//
//	// elsewhere, feed results into the rollup:
//	r.Add(result)
package rollup
