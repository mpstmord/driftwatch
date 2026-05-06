package config

import (
	"fmt"
	"os"
)

// MustLoad loads a Config from path and panics on any error.
// Intended for use during daemon startup where misconfiguration is fatal.
func MustLoad(path string) *Config {
	cfg, err := Load(path)
	if err != nil {
		panic(fmt.Sprintf("driftwatch: fatal config error: %v", err))
	}
	return cfg
}

// LoadOrDefault attempts to load from path; if the file does not exist it
// returns a default Config populated with the provided file paths.
func LoadOrDefault(path string, watchPaths []string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if len(watchPaths) == 0 {
			return nil, fmt.Errorf("config: no config file and no watch paths provided")
		}
		entries := make([]FileEntry, len(watchPaths))
		for i, p := range watchPaths {
			entries[i] = FileEntry{Path: p}
		}
		cfg := &Config{
			Files: entries,
		}
		if err := cfg.validate(); err != nil {
			return nil, err
		}
		return cfg, nil
	}
	return Load(path)
}
