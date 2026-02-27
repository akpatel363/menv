package runner

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Run executes a command with the given environment variables and working directory.
// On Windows it uses cmd /c, on Unix it uses sh -c for shell expansion.
// If cmdParts has multiple elements, they are joined and run through the shell.
// If cmdParts has a single element, it is still run through the shell to support
// commands like "npm start" or piped commands.
func Run(cmdParts []string, envVars []string, workDir string) error {
	if len(cmdParts) == 0 {
		return fmt.Errorf("no command provided")
	}

	cmdStr := strings.Join(cmdParts, " ")
	var c *exec.Cmd

	if runtime.GOOS == "windows" {
		c = exec.Command("cmd", "/c", cmdStr)
	} else {
		c = exec.Command("sh", "-c", cmdStr)
	}

	c.Env = envVars
	c.Dir = workDir
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return c.Run()
}
