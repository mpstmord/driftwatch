// Package audit provides a persistent audit log for drift detection events.
// Each entry records when a drift check occurred, which files were affected,
// and whether drift was detected.
package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Entry represents a single audit log record.
type Entry struct {
	Timestamp time.Time `json:"timestamp"`
	FilePath  string    `json:"file_path"`
	Drifted   bool      `json:"drifted"`
	Message   string    `json:"message,omitempty"`
}

// Log holds a collection of audit entries and writes them to a file.
type Log struct {
	mu      sync.Mutex
	entries []Entry
	path    string
}

// New creates a new Log that persists entries to the given file path.
// If the file already exists its contents are loaded.
func New(path string) (*Log, error) {
	l := &Log{path: path}
	if err := l.load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("audit: load existing log: %w", err)
	}
	return l, nil
}

// Record appends a new entry to the log and flushes it to disk.
func (l *Log) Record(filePath string, drifted bool, message string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	e := Entry{
		Timestamp: time.Now().UTC(),
		FilePath:  filePath,
		Drifted:   drifted,
		Message:   message,
	}
	l.entries = append(l.entries, e)
	return l.flush()
}

// Entries returns a copy of all recorded audit entries.
func (l *Log) Entries() []Entry {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]Entry, len(l.entries))
	copy(out, l.entries)
	return out
}

func (l *Log) flush() error {
	f, err := os.Create(l.path)
	if err != nil {
		return fmt.Errorf("audit: open for write: %w", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(l.entries)
}

func (l *Log) load() error {
	f, err := os.Open(l.path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(&l.entries)
}
