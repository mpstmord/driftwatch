// Package policy provides rule-based evaluation of drift results.
//
// A Policy holds an ordered list of Rules, each pairing a regular-expression
// path pattern with a Disposition (allow, warn, or block). When a drift result
// is evaluated, the first matching rule determines the disposition; if no rule
// matches, Warn is returned as a safe default.
//
// Rules can be loaded programmatically via AddRule or read from a JSON file
// with LoadFile, enabling operators to manage drift response policy outside
// the compiled binary.
//
// Typical usage:
//
//	p, err := policy.LoadFile("/etc/driftwatch/policy.json")
//	if err != nil { ... }
//	disp := p.Evaluate(result)
//	if disp == policy.Block { /* page on-call */ }
package policy
