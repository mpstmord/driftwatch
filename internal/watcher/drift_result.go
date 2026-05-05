package watcher

// DriftResult captures the outcome of a single file check performed by the
// Watcher. It is passed to any registered alert handlers.
type DriftResult struct {
	// Path is the absolute path of the monitored file.
	Path string

	// Drifted is true when the current checksum differs from the baseline.
	Drifted bool

	// Previous is the baseline checksum recorded at startup or last reset.
	Previous string

	// Current is the checksum computed during the most recent check.
	Current string
}
