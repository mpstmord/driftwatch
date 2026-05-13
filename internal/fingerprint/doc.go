// Package fingerprint provides a concurrency-safe tracker that records the
// most recent checksum seen for each watched file path.
//
// On every check cycle the caller calls Update with the current checksum.
// Update returns true only when the checksum differs from the previously
// stored value, making it easy to gate downstream alerting logic.
//
// The zero value is not usable; always construct a Tracker via New.
package fingerprint
