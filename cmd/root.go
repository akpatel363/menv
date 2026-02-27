package cmd

import (
	"fmt"
	"os"

	"menv/internal/config"

	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "menv",
	Short: "menv – manage project environments from the terminal",
	Long: `menv is a CLI tool for managing project environments.

Configure projects, define environments with .env files and overrides,
then run your commands with the correct env vars auto-populated.

  menv project add my-app --path ./my-app --command "npm start"
  menv env add my-app dev --files .env.dev
  menv run my-app dev -- npm start`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: $MENV_CONFIG or ~/.menv.yaml)")
}

func initConfig() {
	if cfgFile != "" {
		config.SetConfigPath(cfgFile)
	}
}

// loadConfig is a helper used by sub-commands.
func loadConfig() *config.Config {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintln(os.Stderr, "Run 'menv init' to create a config file.")
		os.Exit(1)
	}
	return cfg
}

// resolveProject resolves a project by name, or by CWD detection if name is empty.
func resolveProject(cfg *config.Config, name string) (string, config.Project, error) {
	if name != "" {
		p, exists := cfg.Projects[name]
		if !exists {
			return "", config.Project{}, fmt.Errorf("project %q not found", name)
		}
		return name, p, nil
	}

	detected, p := config.DetectProject(cfg)
	if detected == "" {
		return "", config.Project{}, fmt.Errorf("could not detect project from current directory; specify a project name or cd into a project path")
	}
	return detected, *p, nil
}
