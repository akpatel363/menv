package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"menv/internal/config"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:     "env",
	Aliases: []string{"e"},
	Short:   "Manage environments for a project",
}

// --- env list ---

var envListCmd = &cobra.Command{
	Use:     "list <project>",
	Aliases: []string{"ls"},
	Short:   "List environments for a project",
	Args:    cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return getProjectNames(), cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]
		cfg := loadConfig()

		project, exists := cfg.Projects[projectName]
		if !exists {
			return fmt.Errorf("project %q not found", projectName)
		}

		if len(project.Envs) == 0 {
			color.Yellow("No environments configured for %q. Use 'menv env add' to add one.", projectName)
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		bold := color.New(color.Bold)
		bold.Fprintf(w, "ENV\tFILES\tOVERRIDES\n")
		for name, e := range project.Envs {
			overrides := make([]string, 0, len(e.Overrides))
			for k, v := range e.Overrides {
				overrides = append(overrides, k+"="+v)
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", name, strings.Join(e.Files, ", "), strings.Join(overrides, ", "))
		}
		w.Flush()
		return nil
	},
}

// --- env add ---

var (
	envAddFiles     []string
	envAddOverrides []string
)

var envAddCmd = &cobra.Command{
	Use:   "add <project> <env>",
	Short: "Add an environment to a project",
	Args:  cobra.ExactArgs(2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return getProjectNames(), cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]
		envName := args[1]
		cfg := loadConfig()

		project, exists := cfg.Projects[projectName]
		if !exists {
			return fmt.Errorf("project %q not found", projectName)
		}

		if project.Envs == nil {
			project.Envs = make(map[string]config.Env)
		}

		if _, exists := project.Envs[envName]; exists {
			return fmt.Errorf("environment %q already exists in project %q", envName, projectName)
		}

		overrides := make(map[string]string)
		for _, o := range envAddOverrides {
			parts := strings.SplitN(o, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid override format %q (expected KEY=VALUE)", o)
			}
			overrides[parts[0]] = parts[1]
		}

		project.Envs[envName] = config.Env{
			Files:     envAddFiles,
			Overrides: overrides,
		}
		cfg.Projects[projectName] = project

		if err := config.Save(cfg); err != nil {
			return err
		}

		color.Green("✓ Environment %q added to project %q.", envName, projectName)
		return nil
	},
}

// --- env remove ---

var envRemoveCmd = &cobra.Command{
	Use:     "remove <project> <env>",
	Aliases: []string{"rm"},
	Short:   "Remove an environment from a project",
	Args:    cobra.ExactArgs(2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return getProjectNames(), cobra.ShellCompDirectiveNoFileComp
		}
		if len(args) == 1 {
			return getEnvNames(args[0]), cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]
		envName := args[1]
		cfg := loadConfig()

		project, exists := cfg.Projects[projectName]
		if !exists {
			return fmt.Errorf("project %q not found", projectName)
		}

		if _, exists := project.Envs[envName]; !exists {
			return fmt.Errorf("environment %q not found in project %q", envName, projectName)
		}

		delete(project.Envs, envName)
		cfg.Projects[projectName] = project

		if err := config.Save(cfg); err != nil {
			return err
		}

		color.Green("✓ Environment %q removed from project %q.", envName, projectName)
		return nil
	},
}
