package alert

import (
	"fmt"

	"github.com/yourorg/driftwatch/internal/watcher"
)

// DriftHandler wraps an Alerter and implements handling of watcher.DriftResult
// values produced by the file watcher.
type DriftHandler struct {
	alerter *Alerter
}

// NewDriftHandler returns a DriftHandler backed by the given Alerter.
func NewDriftHandler(a *Alerter) *DriftHandler {
	return &DriftHandler{alerter: a}
}

// Handle inspects a DriftResult and dispatches an alert when drift is detected.
// It is designed to be called in the watcher's result loop.
func (h *DriftHandler) Handle(result watcher.DriftResult) error {
	if !result.Drifted {
		return nil
	}

	msg := fmt.Sprintf("checksum changed: previous=%s current=%s",
		result.Previous, result.Current)

	return h.alerter.Drift(result.Path, msg)
}
