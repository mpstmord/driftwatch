package policy_test

import (
	"regexp"
	"testing"

	"github.com/example/driftwatch/internal/policy"
	"github.com/example/driftwatch/internal/watcher"
)

func driftAt(path string) watcher.DriftResult {
	return watcher.DriftResult{Path: path, Drifted: true}
}

func TestEvaluate_DefaultsToWarnWhenNoRules(t *testing.T) {
	p := policy.New()
	got := p.Evaluate(driftAt("/etc/passwd"))
	if got != policy.Warn {
		t.Fatalf("expected Warn, got %s", got)
	}
}

func TestEvaluate_FirstMatchingRuleWins(t *testing.T) {
	p := policy.New()
	p.AddRule(policy.Rule{Name: "allow-cron", Pattern: regexp.MustCompile(`crontab`), Dispose: policy.Allow})
	p.AddRule(policy.Rule{Name: "block-all", Pattern: regexp.MustCompile(`.*`), Dispose: policy.Block})

	if got := p.Evaluate(driftAt("/etc/crontab")); got != policy.Allow {
		t.Fatalf("expected Allow for crontab, got %s", got)
	}
	if got := p.Evaluate(driftAt("/etc/shadow")); got != policy.Block {
		t.Fatalf("expected Block for shadow, got %s", got)
	}
}

func TestEvaluate_BlockDisposition(t *testing.T) {
	p := policy.New()
	p.AddRule(policy.Rule{Name: "block-ssh", Pattern: regexp.MustCompile(`sshd_config`), Dispose: policy.Block})

	got := p.Evaluate(driftAt("/etc/ssh/sshd_config"))
	if got != policy.Block {
		t.Fatalf("expected Block, got %s", got)
	}
}

func TestEvaluate_AllowDisposition(t *testing.T) {
	p := policy.New()
	p.AddRule(policy.Rule{Name: "allow-motd", Pattern: regexp.MustCompile(`motd`), Dispose: policy.Allow})

	got := p.Evaluate(driftAt("/etc/motd"))
	if got != policy.Allow {
		t.Fatalf("expected Allow, got %s", got)
	}
}

func TestRules_ReturnsCopy(t *testing.T) {
	p := policy.New()
	p.AddRule(policy.Rule{Name: "r1", Pattern: regexp.MustCompile(`x`), Dispose: policy.Warn})

	rules := p.Rules()
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	// mutating the returned slice must not affect the policy
	rules[0].Name = "mutated"
	if p.Rules()[0].Name != "r1" {
		t.Fatal("policy rules were mutated through returned slice")
	}
}

func TestDisposition_String(t *testing.T) {
	cases := []struct {
		d    policy.Disposition
		want string
	}{
		{policy.Allow, "allow"},
		{policy.Warn, "warn"},
		{policy.Block, "block"},
	}
	for _, tc := range cases {
		if got := tc.d.String(); got != tc.want {
			t.Errorf("Disposition(%d).String() = %q, want %q", tc.d, got, tc.want)
		}
	}
}
