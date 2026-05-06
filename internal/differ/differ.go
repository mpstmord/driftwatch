// Package differ compares two snapshots and returns a list of drift results.
package differ

import (
	"fmt"

	"github.com/example/driftwatch/internal/snapshot"
	"github.com/example/driftwatch/internal/watcher"
)

// Diff compares a baseline snapshot against a current snapshot and returns
// a DriftResult for every file that was added, removed, or modified.
func Diff(baseline, current *snapshot.Snapshot) []watcher.DriftResult {
	var results []watcher.DriftResult

	baselineRecords := baseline.Records()
	currentRecords := current.Records()

	// Check for modified or removed files.
	for path, baseRec := range baselineRecords {
		curRec, exists := currentRecords[path]
		if !exists {
			results = append(results, watcher.DriftResult{
				Path:    path,
				Drifted: true,
				Reason:  fmt.Sprintf("file removed (was %s)", baseRec.Checksum),
			})
			continue
		}
		if curRec.Checksum != baseRec.Checksum {
			results = append(results, watcher.DriftResult{
				Path:    path,
				Drifted: true,
				Reason:  fmt.Sprintf("checksum changed (%s -> %s)", baseRec.Checksum, curRec.Checksum),
			})
		}
	}

	// Check for newly added files.
	for path, curRec := range currentRecords {
		if _, exists := baselineRecords[path]; !exists {
			results = append(results, watcher.DriftResult{
				Path:    path,
				Drifted: true,
				Reason:  fmt.Sprintf("file added (checksum %s)", curRec.Checksum),
			})
		}
	}

	return results
}
