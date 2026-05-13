// Package debounce provides a path-keyed debouncer for drift events.
//
// When config files are written by tools that perform multiple rapid writes
// (e.g. sed-in-place, atomic rename patterns), the watcher may emit a burst
// of change events for the same path in quick succession. Wrapping alert
// dispatch with a Debouncer ensures that only a single alert is sent once the
// burst settles, reducing noise for operators.
//
// Usage:
//
//	d := debounce.New(2 * time.Second)
//	// called on every watcher event:
//	d.Trigger(path, func() { handler.Handle(result) })
//
// The callback is invoked at most once per quiet window. Calling Trigger
// again before the window expires resets the countdown.
package debounce
