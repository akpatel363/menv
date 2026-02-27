package config

// Config represents the top-level menv configuration.
type Config struct {
	Projects map[string]Project `yaml:"projects"`
}

// Project represents a single project entry.
type Project struct {
	Path    string         `yaml:"path"`
	Command string         `yaml:"command"`
	Envs    map[string]Env `yaml:"envs"`
}

// Env represents an environment within a project.
type Env struct {
	Files     []string          `yaml:"files"`
	Overrides map[string]string `yaml:"overrides"`
}
