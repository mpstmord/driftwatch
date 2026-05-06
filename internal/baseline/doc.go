// Package baseline provides persistence for the trusted reference checksums
// used by driftwatch to detect configuration drift.
//
// A Baseline is a map of file paths to their last-known-good checksums.
// It is saved to and loaded from a JSON file on disk, typically alongside
// the driftwatch state directory.
//
// Typical usage:
//
//	b, err := baseline.Load("/var/lib/driftwatch/baseline.json")
//	if errors.Is(err, baseline.ErrNoBaseline) {
//		b = baseline.New() // first run — no baseline yet
//	}
//
//	// After computing checksums, persist the trusted state:
//	b.Set("/etc/app.conf", checksum)
//	baseline.Save(b, "/var/lib/driftwatch/baseline.json")
package baseline
