// Package scheduler provides a periodic execution loop for drift detection.
//
// A Scheduler wraps a [watcher.Watcher] and an [alert.DriftHandler], ticking
// at a configurable interval. On each tick it calls Watcher.Check, then
// forwards every DriftResult to the handler so that alerts are emitted for
// any files whose checksums have changed since the last baseline snapshot.
//
// Typical usage:
//
//	w, _ := watcher.New(cfg.WatchPaths)
//	h   := alert.NewDriftHandler(alert.New(os.Stdout))
//	s   := scheduler.New(30*time.Second, w, h)
//	s.Run(ctx)
package scheduler
