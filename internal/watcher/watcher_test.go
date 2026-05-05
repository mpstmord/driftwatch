package watcher_test

import (
	"os"
	"testing"
	"time"

	"github.com/driftwatch/internal/watcher"
)

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "driftwatch-watcher-*")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

func TestWatcher_NoDriftOnUnchangedFile(t *testing.T) {
	path := writeTempFile(t, "stable config content")
	w := watcher.New([]string{path}, 50*time.Millisecond)
	w.Start()
	defer w.Stop()

	time.Sleep(200 * time.Millisecond)

	select {
	case evt := <-w.Events:
		t.Errorf("unexpected drift event for unchanged file: %+v", evt)
	default:
		// expected: no events
	}
}

func TestWatcher_DetectsDrift(t *testing.T) {
	path := writeTempFile(t, "original content")
	w := watcher.New([]string{path}, 50*time.Millisecond)
	w.Start()
	defer w.Stop()

	// Allow the watcher to record the initial state.
	time.Sleep(100 * time.Millisecond)

	// Modify the file to trigger drift.
	if err := os.WriteFile(path, []byte("modified content"), 0644); err != nil {
		t.Fatalf("failed to modify file: %v", err)
	}

	select {
	case evt := <-w.Events:
		if evt.Path != path {
			t.Errorf("expected path %s, got %s", path, evt.Path)
		}
		if evt.Previous.SHA256 == evt.Current.SHA256 {
			t.Errorf("expected different checksums, both are %s", evt.Current.SHA256)
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("timed out waiting for drift event")
	}
}

func TestWatcher_MultipleFiles(t *testing.T) {
	path1 := writeTempFile(t, "file one content")
	path2 := writeTempFile(t, "file two content")

	w := watcher.New([]string{path1, path2}, 50*time.Millisecond)
	w.Start()
	defer w.Stop()

	time.Sleep(100 * time.Millisecond)

	if err := os.WriteFile(path2, []byte("file two changed"), 0644); err != nil {
		t.Fatalf("failed to modify file: %v", err)
	}

	select {
	case evt := <-w.Events:
		if evt.Path != path2 {
			t.Errorf("expected drift on %s, got %s", path2, evt.Path)
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("timed out waiting for drift event")
	}
}
