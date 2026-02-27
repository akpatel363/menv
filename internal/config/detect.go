package config

import (
	"os"
	"path/filepath"
	"strings"
)

// DetectProject finds a project whose path matches the current working directory.
// Returns the project name and project config, or empty string if no match.
// It resolves symlinks and normalises paths before comparing.
// A match occurs if cwd equals the project path or is a subdirectory of it.
func DetectProject(cfg *Config) (string, *Project) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", nil
	}
	cwd = NormalizePath(cwd)

	for name, p := range cfg.Projects {
		projectPath := NormalizePath(p.Path)
		if projectPath == "" {
			continue
		}
		if cwd == projectPath || strings.HasPrefix(cwd, projectPath+string(filepath.Separator)) {
			return name, &p
		}
	}
	return "", nil
}
