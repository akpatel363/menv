# menv

Environment manager for projects. Configure envs in YAML, run commands with variables auto-injected.

## Install

```bash
go install github.com/akpatel363/menv@latest
```

Or build from source:

```bash
git clone https://github.com/akpatel363/menv.git
cd menv
go install .
```

Make sure `$GOPATH/bin` is in your `PATH`.

## Quick Start

```bash
# Create config file (~/.menv.yaml)
menv init

# Add a project
menv project add my-api --path ~/code/my-api --command "go run ."

# Add environments
menv env add my-api dev --files .env.dev --override NODE_ENV=development
menv env add my-api prod --files .env.prod,.secrets --override NODE_ENV=production

# Run with env vars loaded
menv run my-api dev
menv run my-api dev -- go run ./cmd/server

# Context-aware: if you're inside the project directory, skip the project name
cd ~/code/my-api
menv run dev
menv run dev -- go run ./cmd/server

# Inspect env vars
menv env get my-api dev              # table view
menv env get dev DB_HOST API_KEY      # specific keys (CWD-aware)
menv env get dev --export             # eval-friendly: eval $(menv env get dev -x)
```

## Config

Default location: `~/.menv.yaml` (override with `$MENV_CONFIG` or `--config`)

```yaml
projects:
  my-api:
    path: /home/user/code/my-api
    command: go run .
    envs:
      dev:
        files:
          - .env.dev
        overrides:
          NODE_ENV: development
      prod:
        files:
          - .env.prod
          - .secrets
        overrides:
          NODE_ENV: production
```

- **files**: `.env`-style files (relative to project path). Supports `KEY=VALUE`, quoted values, `export` prefix, comments.
- **overrides**: Key-value pairs that take precedence over file values. Use this to override specific vars without touching your env files.

## Commands

```
menv init                                  # Create config file
menv project add <name> --path <p> --command <c>  # Add project
menv project list                          # List projects
menv project remove <name>                 # Remove project
menv env add <project> <env> --files <f> --override <K=V>  # Add env
menv env list <project>                    # List envs
menv env remove <project> <env>            # Remove env
menv run <project> <env>                   # Run default command
menv run <project> <env> -- <command>      # Run specific command
menv run <env>                             # Auto-detect project from CWD
menv run <env> -- <command>                # Auto-detect + custom command
menv env get [project] <env>               # Print all env vars
menv env get [project] <env> <key...>      # Print specific vars
menv env get [project] <env> --export      # Output as export statements
```

## Shell Completion

```bash
# Bash
menv completion bash > /etc/bash_completion.d/menv

# Zsh
menv completion zsh > "${fpath[1]}/_menv"

# Fish
menv completion fish > ~/.config/fish/completions/menv.fish

# PowerShell
menv completion powershell | Out-String | Invoke-Expression
```

## Cross-Platform

Works on Linux, macOS, and Windows. Uses `sh -c` on Unix and `cmd /c` on Windows.
