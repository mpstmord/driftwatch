// Package runner provides the top-level orchestration for a single
// driftwatch check cycle.
//
// A Runner is constructed from a [config.Config] and exposes two methods:
//
//   - Run: snapshots watched paths, diffs against the stored baseline, and
//     emits alerts for any files whose checksums have changed.
//
//   - UpdateBaseline: re-snapshots all watched paths and persists the result
//     as the new trusted baseline, effectively acknowledging current state.
//
// Typical usage:
//
//	r, err := runner.New(cfg)
//	if err != nil { ... }
//	if err := r.Run(ctx); err != nil { ... }
package runner
