// Package healthcheck provides a simple HTTP health endpoint that reports
// the current status of the driftwatch daemon, including the last check
// time and whether any drift was detected in the most recent run.
package healthcheck

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// Status holds the current health state of the daemon.
type Status struct {
	Healthy     bool      `json:"healthy"`
	LastChecked time.Time `json:"last_checked"`
	DriftFound  bool      `json:"drift_found"`
	DriftPaths  []string  `json:"drift_paths,omitempty"`
	Message     string    `json:"message"`
}

// Handler exposes an HTTP endpoint that returns the current Status as JSON.
type Handler struct {
	mu     sync.RWMutex
	status Status
}

// New returns a Handler with a default healthy status.
func New() *Handler {
	return &Handler{
		status: Status{
			Healthy: true,
			Message: "no checks run yet",
		},
	}
}

// Update replaces the current status. It is safe for concurrent use.
func (h *Handler) Update(s Status) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.status = s
}

// ServeHTTP writes the current status as a JSON response.
// It returns HTTP 200 when healthy and HTTP 503 when drift is found.
func (h *Handler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	h.mu.RLock()
	s := h.status
	h.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	if !s.Healthy || s.DriftFound {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	_ = json.NewEncoder(w).Encode(s)
}
