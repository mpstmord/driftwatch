// Package notifier provides pluggable notification backends for drift alerts.
package notifier

import (
	"fmt"
	"io"
	"os"
	"time"
)

// Level represents the severity of a notification.
type Level string

const (
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

// Event holds the data for a single drift notification.
type Event struct {
	Level     Level
	File      string
	Message   string
	Timestamp time.Time
}

// Notifier sends drift events to a destination.
type Notifier interface {
	Notify(e Event) error
}

// LogNotifier writes events as structured log lines to a writer.
type LogNotifier struct {
	out io.Writer
}

// New returns a LogNotifier that writes to w.
// If w is nil, os.Stderr is used.
func New(w io.Writer) *LogNotifier {
	if w == nil {
		w = os.Stderr
	}
	return &LogNotifier{out: w}
}

// Notify formats and writes the event to the configured writer.
func (n *LogNotifier) Notify(e Event) error {
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now().UTC()
	}
	_, err := fmt.Fprintf(
		n.out,
		"[%s] %s file=%s msg=%s\n",
		e.Timestamp.Format(time.RFC3339),
		e.Level,
		e.File,
		e.Message,
	)
	return err
}
