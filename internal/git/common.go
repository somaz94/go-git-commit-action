package git

import (
	"fmt"

	"github.com/somaz94/go-git-commit-action/internal/gitcmd"
)

// Command defines a command to be executed.
// It encapsulates the command name, arguments, and a description
// for consistent command execution across the git package.
type Command struct {
	Name string   // Command name (e.g., "git")
	Args []string // Command arguments
	Desc string   // Human-readable description
}

// ExecuteCommandBatch runs a batch of commands with consistent output
// formatting and error handling. It provides visual feedback for each
// command execution and handles errors gracefully.
func ExecuteCommandBatch(r gitcmd.Runner, commands []Command, headerMessage string) error {
	if headerMessage != "" {
		fmt.Println(headerMessage)
	}

	for _, cmd := range commands {
		fmt.Printf("  - %s... ", cmd.Desc)

		if err := r.Run(cmd.Name, cmd.Args...); err != nil {
			// Special handling for "nothing to commit" case
			if isNothingToCommitError(cmd, err) {
				fmt.Println("[WARN] Nothing to commit, skipping...")
				continue
			}

			fmt.Println("FAILED")
			return fmt.Errorf("failed to execute %s: %w", cmd.Name, err)
		}

		fmt.Println("Done")
	}

	return nil
}

// isNothingToCommitError checks if an error is caused by "nothing to commit".
// It uses the exit code from the command failure instead of brittle string matching.
func isNothingToCommitError(cmd Command, err error) bool {
	if len(cmd.Args) == 0 || cmd.Args[0] != "commit" {
		return false
	}
	code, ok := gitcmd.ExitCodeOf(err)
	return ok && code == 1
}
