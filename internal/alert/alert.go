package alert

import (
	"fmt"
	"io"
	"os"
	"time"
)

// Level represents the severity of an alert.
type Level string

const (
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

// Event holds the details of a single drift alert.
type Event struct {
	Timestamp time.Time
	Level     Level
	FilePath  string
	Message   string
}

// Alerter sends drift alert events to one or more sinks.
type Alerter struct {
	sinks []io.Writer
}

// New creates an Alerter that writes to the provided sinks.
// If no sinks are supplied it defaults to os.Stdout.
func New(sinks ...io.Writer) *Alerter {
	if len(sinks) == 0 {
		sinks = []io.Writer{os.Stdout}
	}
	return &Alerter{sinks: sinks}
}

// Send formats and dispatches an Event to all registered sinks.
func (a *Alerter) Send(e Event) error {
	line := fmt.Sprintf("%s [%s] path=%q msg=%q\n",
		e.Timestamp.UTC().Format(time.RFC3339),
		e.Level,
		e.FilePath,
		e.Message,
	)
	for _, w := range a.sinks {
		if _, err := fmt.Fprint(w, line); err != nil {
			return fmt.Errorf("alert: write to sink: %w", err)
		}
	}
	return nil
}

// Drift is a convenience helper that sends a WARN-level drift event.
func (a *Alerter) Drift(filePath, message string) error {
	return a.Send(Event{
		Timestamp: time.Now(),
		Level:     LevelWarn,
		FilePath:  filePath,
		Message:   message,
	})
}
