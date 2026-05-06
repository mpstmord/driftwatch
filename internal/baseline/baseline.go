// Package baseline manages the trusted reference state for watched config files.
// It persists known-good checksums so that drift can be detected on subsequent runs.
package baseline

import (
	"encoding/json"
	"errors"
	"os"
	"time"
)

// ErrNoBaseline is returned when no baseline file exists at the given path.
var ErrNoBaseline = errors.New("baseline: no baseline file found")

// Record holds the trusted checksum for a single file path.
type Record struct {
	Path      string    `json:"path"`
	Checksum  string    `json:"checksum"`
	RecordedAt time.Time `json:"recorded_at"`
}

// Baseline is a collection of trusted records keyed by file path.
type Baseline struct {
	Records map[string]Record `json:"records"`
}

// New returns an empty Baseline ready for use.
func New() *Baseline {
	return &Baseline{Records: make(map[string]Record)}
}

// Set adds or replaces the trusted record for the given path.
func (b *Baseline) Set(path, checksum string) {
	b.Records[path] = Record{
		Path:       path,
		Checksum:   checksum,
		RecordedAt: time.Now().UTC(),
	}
}

// Get returns the Record for path and whether it was found.
func (b *Baseline) Get(path string) (Record, bool) {
	r, ok := b.Records[path]
	return r, ok
}

// Save writes the baseline to disk as JSON.
func Save(b *Baseline, dest string) error {
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(b)
}

// Load reads a baseline from disk. Returns ErrNoBaseline if the file is absent.
func Load(src string) (*Baseline, error) {
	f, err := os.Open(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoBaseline
		}
		return nil, err
	}
	defer f.Close()
	var b Baseline
	if err := json.NewDecoder(f).Decode(&b); err != nil {
		return nil, err
	}
	if b.Records == nil {
		b.Records = make(map[string]Record)
	}
	return &b, nil
}
