package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

var (
	configPath string
	mu         sync.Mutex
)

// DefaultConfigPath returns the default config file path.
func DefaultConfigPath() string {
	if envPath := os.Getenv("MENV_CONFIG"); envPath != "" {
		return envPath
	}
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not determine home directory: %v\n", err)
		os.Exit(1)
	}
	return filepath.Join(home, ".menv.yaml")
}

// SetConfigPath overrides the config file path (e.g. from --config flag).
func SetConfigPath(path string) {
	mu.Lock()
	defer mu.Unlock()
	configPath = path
}

// GetConfigPath returns the resolved config file path.
func GetConfigPath() string {
	mu.Lock()
	defer mu.Unlock()
	if configPath != "" {
		return configPath
	}
	return DefaultConfigPath()
}

// Load reads and parses the config file.
func Load() (*Config, error) {
	path := GetConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	if cfg.Projects == nil {
		cfg.Projects = make(map[string]Project)
	}
	return &cfg, nil
}

// Save writes the config back to disk.
func Save(cfg *Config) error {
	path := GetConfigPath()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	return nil
}

// Exists checks if the config file already exists.
func Exists() bool {
	_, err := os.Stat(GetConfigPath())
	return err == nil
}

// NormalizePath resolves symlinks and returns the cleaned absolute path.
func NormalizePath(p string) string {
	// Expand ~ manually since filepath.Abs doesn't handle it.
	if p == "" {
		return ""
	}
	if p == "~" || strings.HasPrefix(p, "~/") || strings.HasPrefix(p, "~\\") {
		home, err := os.UserHomeDir()
		if err == nil {
			if p == "~" {
				p = home
			} else {
				p = filepath.Join(home, p[2:])
			}
		}
	}

	abs, err := filepath.Abs(p)
	if err != nil {
		return p
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		// Path might not exist yet, fall back to abs.
		return abs
	}
	return resolved
}
