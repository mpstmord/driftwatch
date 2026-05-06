package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the full driftwatch daemon configuration.
type Config struct {
	Interval time.Duration `yaml:"interval"`
	Alert    AlertConfig   `yaml:"alert"`
	Files    []FileEntry   `yaml:"files"`
}

// AlertConfig configures where alerts are sent.
type AlertConfig struct {
	Level  string `yaml:"level"`
	Output string `yaml:"output"` // "stdout" or a file path
}

// FileEntry describes a single file to watch.
type FileEntry struct {
	Path string `yaml:"path"`
	Tag  string `yaml:"tag"`
}

// Load reads and parses a YAML config file from the given path.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("config: open %q: %w", path, err)
	}
	defer f.Close()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("config: decode %q: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config: invalid: %w", err)
	}

	return &cfg, nil
}

// validate checks required fields and applies defaults.
func (c *Config) validate() error {
	if len(c.Files) == 0 {
		return fmt.Errorf("at least one file entry is required")
	}
	for i, fe := range c.Files {
		if fe.Path == "" {
			return fmt.Errorf("files[%d]: path must not be empty", i)
		}
	}
	if c.Interval <= 0 {
		c.Interval = 30 * time.Second
	}
	if c.Alert.Level == "" {
		c.Alert.Level = "warn"
	}
	if c.Alert.Output == "" {
		c.Alert.Output = "stdout"
	}
	return nil
}
