package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"menv/internal/config"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var projectCmd = &cobra.Command{
	Use:     "project",
	Aliases: []string{"p"},
	Short:   "Manage projects",
}

// --- project list ---

var projectListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all configured projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := loadConfig()

		if len(cfg.Projects) == 0 {
			color.Yellow("No projects configured. Use 'menv project add' to add one.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		bold := color.New(color.Bold)
		bold.Fprintf(w, "PROJECT\tPATH\tCOMMAND\tENVS\n")
		for name, p := range cfg.Projects {
			envNames := make([]string, 0, len(p.Envs))
			for e := range p.Envs {
				envNames = append(envNames, e)
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%v\n", name, p.Path, p.Command, envNames)
		}
		w.Flush()
		return nil
	},
}

// --- project add ---

var (
	projectAddPath    string
	projectAddCommand string
)

var projectAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a new project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		cfg := loadConfig()

		if _, exists := cfg.Projects[name]; exists {
			return fmt.Errorf("project %q already exists", name)
		}

		cfg.Projects[name] = config.Project{
			Path:    config.NormalizePath(projectAddPath),
			Command: projectAddCommand,
			Envs:    make(map[string]config.Env),
		}

		if err := config.Save(cfg); err != nil {
			return err
		}

		color.Green("✓ Project %q added.", name)
		return nil
	},
}

// --- project remove ---

var projectRemoveCmd = &cobra.Command{
	Use:     "remove <name>",
	Aliases: []string{"rm"},
	Short:   "Remove a project",
	Args:    cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return getProjectNames(), cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		cfg := loadConfig()

		if _, exists := cfg.Projects[name]; !exists {
			return fmt.Errorf("project %q not found", name)
		}

		delete(cfg.Projects, name)
		if err := config.Save(cfg); err != nil {
			return err
		}

		color.Green("✓ Project %q removed.", name)
		return nil
	},
}

func init() {
	projectAddCmd.Flags().StringVar(&projectAddPath, "path", ".", "project root directory")
	projectAddCmd.Flags().StringVar(&projectAddCommand, "command", "", "default run command")

	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectAddCmd)
	projectCmd.AddCommand(projectRemoveCmd)

	rootCmd.AddCommand(projectCmd)
}

// getProjectNames returns all configured project names for shell completion.
func getProjectNames() []string {
	cfg, err := config.Load()
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(cfg.Projects))
	for n := range cfg.Projects {
		names = append(names, n)
	}
	return names
}
