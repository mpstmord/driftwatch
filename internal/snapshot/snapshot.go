// Package snapshot provides functionality for persisting and loading
// file checksum snapshots to disk, enabling drift detection across daemon restarts.
package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/yourusername/driftwatch/internal/checksum"
)

// Record holds the checksum and metadata for a single watched file.
type Record struct {
	Path      string            `json:"path"`
	Checksum  checksum.Checksum `json:"checksum"`
	CapturedAt time.Time        `json:"captured_at"`
}

// Snapshot is a collection of Records representing a point-in-time
// view of all watched files.
type Snapshot struct {
	Version   int      `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	Records   []Record  `json:"records"`
}

// Save serialises the Snapshot to the given file path as JSON.
func Save(path string, snap Snapshot) error {
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return fmt.Errorf("snapshot: marshal: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("snapshot: write %s: %w", path, err)
	}
	return nil
}

// Load reads and deserialises a Snapshot from the given file path.
// Returns os.ErrNotExist if the file does not exist.
func Load(path string) (Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Snapshot{}, fmt.Errorf("snapshot: read %s: %w", path, err)
	}
	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return Snapshot{}, fmt.Errorf("snapshot: unmarshal: %w", err)
	}
	return snap, nil
}

// New builds a Snapshot by computing checksums for the provided file paths.
func New(paths []string) (Snapshot, error) {
	records := make([]Record, 0, len(paths))
	for _, p := range paths {
		cs, err := checksum.Compute(p)
		if err != nil {
			return Snapshot{}, fmt.Errorf("snapshot: compute checksum for %s: %w", p, err)
		}
		records = append(records, Record{
			Path:       p,
			Checksum:   cs,
			CapturedAt: time.Now().UTC(),
		})
	}
	return Snapshot{
		Version:   1,
		CreatedAt: time.Now().UTC(),
		Records:   records,
	}, nil
}
