package checksum

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"time"
)

// FileChecksum holds the checksum and metadata for a single file.
type FileChecksum struct {
	Path      string    `json:"path"`
	SHA256    string    `json:"sha256"`
	Size      int64     `json:"size"`
	ModTime   time.Time `json:"mod_time"`
	ComputedAt time.Time `json:"computed_at"`
}

// Compute calculates the SHA-256 checksum of the file at the given path
// and returns a FileChecksum with metadata.
func Compute(path string) (*FileChecksum, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("checksum: open %q: %w", path, err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("checksum: stat %q: %w", path, err)
	}

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, fmt.Errorf("checksum: read %q: %w", path, err)
	}

	return &FileChecksum{
		Path:       path,
		SHA256:     hex.EncodeToString(h.Sum(nil)),
		Size:       info.Size(),
		ModTime:    info.ModTime().UTC(),
		ComputedAt: time.Now().UTC(),
	}, nil
}

// Equal returns true when two FileChecksum values represent identical file
// content (compares SHA-256 digests only).
func Equal(a, b *FileChecksum) bool {
	if a == nil || b == nil {
		return false
	}
	return a.SHA256 == b.SHA256
}
