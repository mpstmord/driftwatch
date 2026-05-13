// Package watchlist provides a concurrency-safe registry of file paths
// that driftwatch monitors for configuration drift.
//
// Paths can be added and removed at runtime, allowing the daemon to adapt
// to configuration changes without restarting. All operations are safe for
// concurrent use by multiple goroutines.
//
// Example:
//
//	wl := watchlist.FromPaths(cfg.WatchPaths)
//	_ = wl.Add("/etc/nginx/nginx.conf")
//	for _, p := range wl.Paths() {
//		// monitor p
//	}
package watchlist
