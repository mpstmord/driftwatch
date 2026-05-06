// Package audit provides a persistent, append-only audit log for driftwatch.
//
// Each time the scheduler runs a drift check the results are recorded as
// structured JSON entries on disk. This allows operators to review a history
// of when drift was first detected and which files were affected, without
// relying solely on real-time alerting.
//
// Usage:
//
//	log, err := audit.New("/var/lib/driftwatch/audit.json")
//	if err != nil { ... }
//
//	log.Record("/etc/nginx/nginx.conf", true, "checksum mismatch")
package audit
