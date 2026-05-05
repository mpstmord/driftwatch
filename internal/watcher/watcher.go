package watcher

import (
	"fmt"
	"sync"
	"time"

	"github.com/driftwatch/internal/checksum"
)

// FileState holds the last known checksum and metadata for a watched file.
type FileState struct {
	Path     string
	Checksum checksum.Result
	LastSeen time.Time
}

// DriftEvent is emitted when a file's checksum changes.
type DriftEvent struct {
	Path     string
	Previous checksum.Result
	Current  checksum.Result
	Detected time.Time
}

// Watcher monitors a set of file paths for checksum drift.
type Watcher struct {
	mu       sync.Mutex
	paths    []string
	state    map[string]FileState
	interval time.Duration
	Events   chan DriftEvent
	stopCh   chan struct{}
}

// New creates a new Watcher for the given paths and poll interval.
func New(paths []string, interval time.Duration) *Watcher {
	return &Watcher{
		paths:    paths,
		state:    make(map[string]FileState),
		interval: interval,
		Events:   make(chan DriftEvent, 16),
		stopCh:   make(chan struct{}),
	}
}

// Start begins polling in a background goroutine.
func (w *Watcher) Start() {
	go func() {
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				w.poll()
			case <-w.stopCh:
				return
			}
		}
	}()
}

// Stop halts the background polling goroutine.
func (w *Watcher) Stop() {
	close(w.stopCh)
}

// poll checks each file and emits DriftEvents for any changes.
func (w *Watcher) poll() {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, path := range w.paths {
		current, err := checksum.Compute(path)
		if err != nil {
			fmt.Printf("watcher: error computing checksum for %s: %v\n", path, err)
			continue
		}

		prev, seen := w.state[path]
		if !seen {
			w.state[path] = FileState{Path: path, Checksum: current, LastSeen: time.Now()}
			continue
		}

		if !checksum.Equal(prev.Checksum, current) {
			w.Events <- DriftEvent{
				Path:     path,
				Previous: prev.Checksum,
				Current:  current,
				Detected: time.Now(),
			}
			w.state[path] = FileState{Path: path, Checksum: current, LastSeen: time.Now()}
		}
	}
}
