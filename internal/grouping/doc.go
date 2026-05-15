// Package grouping classifies drift results into named groups based on
// file-path prefix rules.
//
// A Grouper holds an ordered list of prefix→label rules. When a DriftResult
// is recorded, its path is matched against each rule in insertion order and
// placed into the first matching group. Paths that match no rule are placed
// in the built-in "default" group.
//
// Example:
//
//	g := grouping.New()
//	_ = g.AddRule("/etc/nginx", "nginx")
//	_ = g.AddRule("/etc/app",   "app")
//
//	g.Record(result) // routed by prefix
//
//	for _, grp := range g.All() {
//		fmt.Printf("%s: %d drift(s)\n", grp.Label, len(grp.Results))
//	}
package grouping
