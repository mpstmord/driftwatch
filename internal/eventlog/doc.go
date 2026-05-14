// Package eventlog provides an in-memory, bounded ring buffer for recording
// drift events detected by driftwatch.
//
// The EventLog is safe for concurrent use. It retains at most N entries
// (configurable at construction time, defaulting to 100). When the buffer is
// full the oldest entry is evicted to make room for the newest, ensuring
// memory usage stays bounded over long-running daemon sessions.
//
// Typical usage:
//
//	log := eventlog.New(200)
//	// inside your drift handler:
//	log.Record(result)
//	// query for observability:
//	entries := log.All()
package eventlog
