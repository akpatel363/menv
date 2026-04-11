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
//
// A single-element cmdParts is passed to the shell verbatim, so things like
// "npm start" or "go run ." configured as a project's default command work as
// expected. With multiple elements (typically from `menv run ... -- cmd args`),
// each element is shell-quoted before joining so that arguments containing
// spaces or shell metacharacters are preserved.
func Run(cmdParts []string, envVars []string, workDir string) error {
	if len(cmdParts) == 0 {
		return fmt.Errorf("no command provided")
	}

	var cmdStr string
	if len(cmdParts) == 1 {
		cmdStr = cmdParts[0]
	} else {
		quoted := make([]string, len(cmdParts))
		for i, p := range cmdParts {
			quoted[i] = shellQuote(p)
		}
		cmdStr = strings.Join(quoted, " ")
	}

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

// shellQuote escapes a single argument so it survives a round-trip through
// `sh -c` (or `cmd /c` on Windows). It's intentionally conservative: any
// argument containing whitespace or shell metacharacters is wrapped, so
// the resulting command string keeps the original token boundaries.
func shellQuote(s string) string {
	if s == "" {
		if runtime.GOOS == "windows" {
			return `""`
		}
		return "''"
	}

	if runtime.GOOS == "windows" {
		// cmd.exe quoting is gnarly; this handles the common cases
		// (spaces, embedded quotes) without trying to be a full parser.
		if !strings.ContainsAny(s, " \t\"&|<>^") {
			return s
		}
		return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
	}

	// POSIX: single-quote everything except bare safe chars. Embedded
	// single quotes are closed, escaped, and reopened: '\''.
	const safe = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789@%+=:,./-_"
	for _, r := range s {
		if !strings.ContainsRune(safe, r) {
			return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
		}
	}
	return s
}
