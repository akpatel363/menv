package cmd

import (
	"fmt"

	"github.com/akpatel363/menv/internal/config"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new menv config file",
	Long:  `Creates a new .menv.yaml config file with a sample project structure.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.GetConfigPath()
		if err != nil {
			return err
		}
		if config.Exists() {
			color.Yellow("Config file already exists at %s", path)
			return nil
		}

		cfg := &config.Config{
			Projects: map[string]config.Project{
				"example": {
					Path:    "/path/to/project",
					Command: "echo hello",
					Envs: map[string]config.Env{
						"dev": {
							Files:     []string{".env.dev"},
							Overrides: map[string]string{"NODE_ENV": "development"},
						},
					},
				},
			},
		}

		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("failed to create config: %w", err)
		}

		color.Green("✓ Config file created at %s", path)
		color.Cyan("  Edit it to add your projects and environments.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
