package healthcheck_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/driftwatch/internal/healthcheck"
)

func decode(t *testing.T, body []byte) healthcheck.Status {
	t.Helper()
	var s healthcheck.Status
	if err := json.Unmarshal(body, &s); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return s
}

func TestNew_DefaultsToHealthy(t *testing.T) {
	h := healthcheck.New()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	s := decode(t, rec.Body.Bytes())
	if !s.Healthy {
		t.Error("expected healthy=true by default")
	}
}

func TestServeHTTP_Returns503WhenDriftFound(t *testing.T) {
	h := healthcheck.New()
	h.Update(healthcheck.Status{
		Healthy:    true,
		DriftFound: true,
		DriftPaths: []string{"/etc/hosts"},
		Message:    "drift detected",
	})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
	s := decode(t, rec.Body.Bytes())
	if len(s.DriftPaths) != 1 || s.DriftPaths[0] != "/etc/hosts" {
		t.Errorf("unexpected drift paths: %v", s.DriftPaths)
	}
}

func TestServeHTTP_Returns200WhenNoDrift(t *testing.T) {
	h := healthcheck.New()
	now := time.Now()
	h.Update(healthcheck.Status{
		Healthy:     true,
		DriftFound:  false,
		LastChecked: now,
		Message:     "all clear",
	})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	s := decode(t, rec.Body.Bytes())
	if s.Message != "all clear" {
		t.Errorf("unexpected message: %q", s.Message)
	}
}

func TestServeHTTP_ContentTypeIsJSON(t *testing.T) {
	h := healthcheck.New()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
}
