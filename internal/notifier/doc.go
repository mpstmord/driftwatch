// Package notifier provides pluggable notification backends used by driftwatch
// to surface config-drift events to operators.
//
// The core interface is Notifier, which accepts an Event and delivers it to
// a configured destination. The built-in LogNotifier writes RFC 3339-stamped
// log lines to any io.Writer, making it easy to integrate with log aggregators
// or redirect output to files.
//
// Usage:
//
//	n := notifier.New(os.Stdout)
//	n.Notify(notifier.Event{
//		Level:   notifier.LevelWarn,
//		File:    "/etc/myapp/config.yaml",
//		Message: "checksum mismatch detected",
//	})
package notifier
