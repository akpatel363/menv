package env

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/akpatel363/menv/internal/config"
)

// LoadEnv loads environment variables from the given project and env config.
// It reads .env-style files relative to the project path, then applies overrides.
// Returns a merged map of key=value pairs.
func LoadEnv(project config.Project, envCfg config.Env) (map[string]string, error) {
	result := make(map[string]string)

	for _, f := range envCfg.Files {
		filePath := f
		if !filepath.IsAbs(filePath) {
			filePath = filepath.Join(project.Path, f)
		}

		vars, err := parseEnvFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load env file %s: %w", filePath, err)
		}
		for k, v := range vars {
			result[k] = v
		}
	}

	// Overrides take precedence over file-loaded values.
	for k, v := range envCfg.Overrides {
		result[k] = v
	}

	return result, nil
}

// BuildEnv merges the current OS environment with the loaded env vars.
// Loaded vars override existing OS vars with the same key.
func BuildEnv(loaded map[string]string) []string {
	existing := make(map[string]string)
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			existing[parts[0]] = parts[1]
		}
	}

	// Apply loaded env vars (overrides OS vars).
	for k, v := range loaded {
		existing[k] = v
	}

	env := make([]string, 0, len(existing))
	for k, v := range existing {
		env = append(env, k+"="+v)
	}
	return env
}

// parseEnvFile reads a .env file and returns the key-value pairs.
// Supports:
//   - KEY=VALUE
//   - KEY="VALUE" (double-quoted, strips quotes)
//   - KEY='VALUE' (single-quoted, strips quotes)
//   - Comments (#)
//   - Empty lines
//   - export KEY=VALUE (strips the `export` prefix)
func parseEnvFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	result := make(map[string]string)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments.
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Strip optional 'export ' prefix.
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimPrefix(line, "export ")
			line = strings.TrimSpace(line)
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove surrounding quotes.
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		result[key] = value
	}

	return result, scanner.Err()
}
