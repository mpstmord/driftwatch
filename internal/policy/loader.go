package policy

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
)

// ruleSpec is the JSON-serialisable form of a Rule.
type ruleSpec struct {
	Name    string `json:"name"`
	Pattern string `json:"pattern"`
	Dispose string `json:"disposition"`
}

func dispositionFromString(s string) (Disposition, error) {
	switch s {
	case "allow":
		return Allow, nil
	case "warn":
		return Warn, nil
	case "block":
		return Block, nil
	default:
		return Warn, fmt.Errorf("policy: unknown disposition %q", s)
	}
}

// LoadFile reads a JSON policy file and returns a populated Policy.
// The file must contain a JSON array of rule objects, e.g.:
//
//	[{"name":"allow-motd","pattern":"/etc/motd","disposition":"allow"}]
func LoadFile(path string) (*Policy, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("policy: open %s: %w", path, err)
	}
	defer f.Close()

	var specs []ruleSpec
	if err := json.NewDecoder(f).Decode(&specs); err != nil {
		return nil, fmt.Errorf("policy: decode %s: %w", path, err)
	}

	p := New()
	for i, s := range specs {
		re, err := regexp.Compile(s.Pattern)
		if err != nil {
			return nil, fmt.Errorf("policy: rule %d pattern: %w", i, err)
		}
		d, err := dispositionFromString(s.Dispose)
		if err != nil {
			return nil, err
		}
		p.AddRule(Rule{Name: s.Name, Pattern: re, Dispose: d})
	}
	return p, nil
}
