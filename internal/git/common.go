package git

import (
	"fmt"
	"os"
	"os/exec"
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
func ExecuteCommandBatch(commands []Command, headerMessage string) error {
	if headerMessage != "" {
		fmt.Println(headerMessage)
	}

	for _, cmd := range commands {
		fmt.Printf("  • %s... ", cmd.Desc)
		command := exec.Command(cmd.Name, cmd.Args...)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr

		if err := command.Run(); err != nil {
			// Special handling for "nothing to commit" case
			if len(cmd.Args) > 0 && cmd.Args[0] == "commit" && err.Error() == "exit status 1" {
				fmt.Println("⚠️  Nothing to commit, skipping...")
				continue
			}

			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to execute %s: %v", cmd.Name, err)
		}

		fmt.Println("✅ Done")
	}

	return nil
}
