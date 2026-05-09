// Package suppress provides time-bounded alert suppression for specific file
// paths within driftwatch.
//
// During planned maintenance or known configuration rollouts, operators can
// suppress drift alerts for individual paths for a fixed duration. Once the
// duration elapses the suppression expires automatically and drift detection
// resumes as normal.
//
// Usage:
//
//	sm := suppress.New()
//	sm.Suppress("/etc/nginx/nginx.conf", 30*time.Minute)
//
//	if sm.IsSuppressed(path) {
//		// skip alerting
//	}
//
// Expired suppressions can be cleaned up at any time by calling Purge.
package suppress
