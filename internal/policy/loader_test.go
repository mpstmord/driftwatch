package policy_test

import (
	"os"
	"testing"

	"github.com/example/driftwatch/internal/policy"
)

func writePolicyFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "policy-*.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

func TestLoadFile_ValidPolicy(t *testing.T) {
	path := writePolicyFile(t, `[
		{"name":"allow-motd","pattern":"/etc/motd","disposition":"allow"},
		{"name":"block-ssh","pattern":"sshd_config","disposition":"block"}
	]`)

	p, err := policy.LoadFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(p.Rules()) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(p.Rules()))
	}
}

func TestLoadFile_MissingFile(t *testing.T) {
	_, err := policy.LoadFile("/nonexistent/policy.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadFile_InvalidJSON(t *testing.T) {
	path := writePolicyFile(t, `not json`)
	_, err := policy.LoadFile(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestLoadFile_InvalidPattern(t *testing.T) {
	path := writePolicyFile(t, `[{"name":"bad","pattern":"[invalid","disposition":"warn"}]`)
	_, err := policy.LoadFile(path)
	if err == nil {
		t.Fatal("expected error for invalid regex pattern")
	}
}

func TestLoadFile_UnknownDisposition(t *testing.T) {
	path := writePolicyFile(t, `[{"name":"x","pattern":".*","disposition":"explode"}]`)
	_, err := policy.LoadFile(path)
	if err == nil {
		t.Fatal("expected error for unknown disposition")
	}
}

func TestLoadFile_EvaluatesCorrectly(t *testing.T) {
	path := writePolicyFile(t, `[
		{"name":"allow-tmp","pattern":"/tmp/","disposition":"allow"}
	]`)
	p, err := policy.LoadFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := p.Evaluate(driftAt("/tmp/test.conf")); got != policy.Allow {
		t.Fatalf("expected Allow, got %s", got)
	}
	if got := p.Evaluate(driftAt("/etc/passwd")); got != policy.Warn {
		t.Fatalf("expected Warn fallback, got %s", got)
	}
}
