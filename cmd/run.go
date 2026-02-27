package cmd

import (
	"fmt"

	"menv/internal/config"
	"menv/internal/env"
	"menv/internal/runner"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [project] <env> [-- command ...]",
	Short: "Run a command with environment variables loaded",
	Long: `Loads environment variables from the configured files and overrides
for the given project/env, then executes the command.

If no command is provided after --, the project's default command is used.
If you are inside a project directory, the project name can be omitted.

Examples:
  menv run my-app dev
  menv run dev                         # auto-detect project from CWD
  menv run my-app dev -- npm run build
  menv run dev -- npm run build        # auto-detect + custom command`,
	Args:                  cobra.MinimumNArgs(1),
	DisableFlagParsing:    false,
	DisableFlagsInUseLine: true,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		switch len(args) {
		case 0:
			// Could be a project name or env name (if CWD-detected).
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

		argsBeforeDash := args
		dashIdx := cmd.ArgsLenAtDash()
		if dashIdx >= 0 {
			argsBeforeDash = args[:dashIdx]
		}

		switch len(argsBeforeDash) {
		case 1:
			// Only env name provided — detect project from CWD.
			envName = argsBeforeDash[0]
			var err error
			projectName, project, err = resolveProject(cfg, "")
			if err != nil {
				return err
			}
		case 2:
			// Both project and env provided.
			projectName = argsBeforeDash[0]
			envName = argsBeforeDash[1]
			var err error
			_, project, err = resolveProject(cfg, projectName)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("expected 1 or 2 positional arguments (env or project env), got %d", len(argsBeforeDash))
		}

		envCfg, exists := project.Envs[envName]
		if !exists {
			return fmt.Errorf("environment %q not found in project %q", envName, projectName)
		}

		// Determine the command to run.
		var cmdToRun []string
		if dashIdx >= 0 && dashIdx < len(args) {
			cmdToRun = args[dashIdx:]
		} else if project.Command != "" {
			cmdToRun = []string{project.Command}
		} else {
			return fmt.Errorf("no command provided and no default command configured for project %q", projectName)
		}

		// Load env variables.
		loaded, err := env.LoadEnv(project, envCfg)
		if err != nil {
			return err
		}

		envVars := env.BuildEnv(loaded)

		color.Cyan("» project: %s | env: %s", projectName, envName)
		color.Cyan("» directory: %s", project.Path)
		if len(loaded) > 0 {
			color.HiBlack("  loaded %d env variable(s)", len(loaded))
		}
		color.Cyan("» running: %v", cmdToRun)
		fmt.Println()

		return runner.Run(cmdToRun, envVars, project.Path)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
