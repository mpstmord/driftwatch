// Package pipeline provides a composable event-processing chain for
// driftwatch drift results.
//
// A Pipeline gates each drifted-file event through:
//
//  1. filter.Filter  – path include/exclude glob rules
//  2. dedupe.Dedupe  – suppresses repeated identical checksums within a window
//  3. cooldown.Cooldown – enforces a minimum quiet period per path
//  4. notifier.Notifier – delivers the surviving event
//
// Construct a Pipeline with New, then call Process on every batch of
// DriftResults produced by the watcher or runner.
package pipeline
