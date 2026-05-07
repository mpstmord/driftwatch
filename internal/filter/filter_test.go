package filter_test

import (
	"testing"

	"github.com/driftwatch/internal/filter"
)

func TestAllow_NoPatterns_AllowsAll(t *testing.T) {
	f := filter.New(nil, nil)

	paths := []string{"/etc/nginx/nginx.conf", "/etc/hosts", "/var/app/config.yaml"}
	for _, p := range paths {
		ok, err := f.Allow(p)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", p, err)
		}
		if !ok {
			t.Errorf("expected %q to be allowed, but it was blocked", p)
		}
	}
}

func TestAllow_IncludePattern_BlocksNonMatch(t *testing.T) {
	f := filter.New([]string{"/etc/*"}, nil)

	ok, err := f.Allow("/var/app/config.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected /var/app/config.yaml to be blocked by include filter")
	}
}

func TestAllow_IncludePattern_AllowsMatch(t *testing.T) {
	f := filter.New([]string{"/etc/*"}, nil)

	ok, err := f.Allow("/etc/hosts")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected /etc/hosts to be allowed by include filter")
	}
}

func TestAllow_ExcludePattern_BlocksMatch(t *testing.T) {
	f := filter.New(nil, []string{"/etc/passwd"})

	ok, err := f.Allow("/etc/passwd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected /etc/passwd to be blocked by exclude filter")
	}
}

func TestAllow_ExcludePattern_AllowsNonMatch(t *testing.T) {
	f := filter.New(nil, []string{"/etc/passwd"})

	ok, err := f.Allow("/etc/hosts")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected /etc/hosts to be allowed when it does not match the exclude pattern")
	}
}

func TestAllow_IncludeAndExclude_ExcludeTakesPrecedence(t *testing.T) {
	// Include all /etc/* but exclude /etc/shadow specifically.
	f := filter.New([]string{"/etc/*"}, []string{"/etc/shadow"})

	ok, err := f.Allow("/etc/shadow")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected /etc/shadow to be blocked even though it matches the include pattern")
	}

	ok, err = f.Allow("/etc/hosts")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected /etc/hosts to be allowed")
	}
}

func TestAllow_InvalidPattern_ReturnsError(t *testing.T) {
	f := filter.New([]string{"[invalid"}, nil)

	_, err := f.Allow("/etc/hosts")
	if err == nil {
		t.Error("expected an error for an invalid glob pattern, got nil")
	}
}
