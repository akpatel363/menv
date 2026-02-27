package cmd

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"menv/internal/config"
	"menv/internal/env"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var envGetCmd = &cobra.Command{
	Use:   "get [project] <env> [key...]",
	Short: "Print environment variables for a project/env",
	Long: `Loads and prints environment variables for the given project and environment.
If no keys are specified, all variables are printed.
If you are inside a project directory, the project name can be omitted.

This is useful for inspecting what variables will be injected, or for
piping into other tools with eval:

  eval $(menv env get dev)

Examples:
  menv env get my-app dev                # print all vars
  menv env get dev                       # auto-detect project from CWD
  menv env get dev DB_HOST API_KEY       # print specific vars
  menv env get my-app dev DB_HOST        # print specific var`,
	Args: cobra.MinimumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		switch len(args) {
		case 0:
			suggestions := getProjectNames()
			cfg, err := config.Load()
			if err == nil {
				if detected, _ := config.DetectProject(cfg); detected != "" {
					suggestions = append(suggestions, getEnvNames(detected)...)
				}
			}
			return suggestions, cobra.ShellCompDirectiveNoFileComp
		case 1:
			return getEnvNames(args[0]), cobra.ShellCompDirectiveNoFileComp
		default:
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := loadConfig()

		var projectName, envName string
		var project config.Project
		var keys []string

		// Determine if first arg is a project name or an env name.
		_, isProject := cfg.Projects[args[0]]

		if isProject && len(args) >= 2 {
			projectName = args[0]
			envName = args[1]
			keys = args[2:]
			var err error
			_, project, err = resolveProject(cfg, projectName)
			if err != nil {
				return err
			}
		} else {
			// First arg is env name — detect project from CWD.
			envName = args[0]
			keys = args[1:]
			var err error
			projectName, project, err = resolveProject(cfg, "")
			if err != nil {
				return err
			}
		}

		envCfg, exists := project.Envs[envName]
		if !exists {
			return fmt.Errorf("environment %q not found in project %q", envName, projectName)
		}

		loaded, err := env.LoadEnv(project, envCfg)
		if err != nil {
			return err
		}

		exportFormat, _ := cmd.Flags().GetBool("export")

		if len(keys) > 0 {
			// Print only requested keys.
			for _, k := range keys {
				v, ok := loaded[k]
				if !ok {
					color.Yellow("# %s not set", k)
					continue
				}
				if exportFormat {
					fmt.Fprintf(os.Stdout, "export %s=%q\n", k, v)
				} else {
					fmt.Fprintf(os.Stdout, "%s=%s\n", k, v)
				}
			}
			return nil
		}

		// Print all variables.
		sortedKeys := make([]string, 0, len(loaded))
		for k := range loaded {
			sortedKeys = append(sortedKeys, k)
		}
		sort.Strings(sortedKeys)

		if exportFormat {
			for _, k := range sortedKeys {
				fmt.Fprintf(os.Stdout, "export %s=%q\n", k, loaded[k])
			}
		} else {
			color.Cyan("» project: %s | env: %s", projectName, envName)
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			bold := color.New(color.Bold)
			bold.Fprintf(w, "KEY\tVALUE\tSOURCE\n")
			for _, k := range sortedKeys {
				fmt.Fprintf(w, "%s\t%s\t\n", k, loaded[k])
			}
			w.Flush()
		}

		return nil
	},
}

func init() {
	envGetCmd.Flags().BoolP("export", "x", false, "output in export format (for eval)")

	envAddCmd.Flags().StringSliceVarP(&envAddFiles, "files", "f", nil, "env files (comma-separated or repeated)")
	envAddCmd.Flags().StringSliceVarP(&envAddOverrides, "override", "o", nil, "env overrides as KEY=VALUE (comma-separated or repeated)")

	envCmd.AddCommand(envListCmd)
	envCmd.AddCommand(envAddCmd)
	envCmd.AddCommand(envRemoveCmd)
	envCmd.AddCommand(envGetCmd)

	rootCmd.AddCommand(envCmd)
}

// getEnvNames returns all env names for a given project (for shell completion).
func getEnvNames(projectName string) []string {
	cfg, err := config.Load()
	if err != nil {
		return nil
	}
	project, exists := cfg.Projects[projectName]
	if !exists {
		return nil
	}
	names := make([]string, 0, len(project.Envs))
	for n := range project.Envs {
		names = append(names, n)
	}
	return names
}
