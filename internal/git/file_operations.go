package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/somaz94/go-git-commit-action/internal/config"
	"github.com/somaz94/go-git-commit-action/internal/gitcmd"
)

// StageFiles adds the specified files to the Git staging area.
// It handles multiple file patterns separated by spaces.
func StageFiles(filePattern string) error {
	fmt.Printf("  • Adding files... ")

	// Handle multiple patterns separated by spaces
	if strings.Contains(filePattern, " ") {
		patterns := strings.Fields(filePattern)
		for _, pattern := range patterns {
			if err := executeGitAdd(pattern); err != nil {
				fmt.Println("❌ Failed")
				return fmt.Errorf("failed to add pattern %s: %v", pattern, err)
			}
		}
	} else {
		// Single pattern case
		if err := executeGitAdd(filePattern); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to add files: %v", err)
		}
	}

	fmt.Println("✅ Done")
	return nil
}

// executeGitAdd executes the git add command for a specific pattern.
func executeGitAdd(pattern string) error {
	addCmd := exec.Command(gitcmd.CmdGit, gitcmd.AddArgs(pattern)...)
	addCmd.Stdout = os.Stdout
	addCmd.Stderr = os.Stderr
	return addCmd.Run()
}

// CommitAndPush commits the staged changes and pushes them to the remote branch.
func CommitAndPush(config *config.GitConfig, branch string) error {
	commitPushCommands := []Command{
		{gitcmd.CmdGit, gitcmd.CommitArgs(config.CommitMessage), "Committing changes"},
		{gitcmd.CmdGit, gitcmd.PushUpstreamArgs(gitcmd.RefOrigin, branch), "Pushing changes"},
	}

	return ExecuteCommandBatch(commitPushCommands, "")
}
