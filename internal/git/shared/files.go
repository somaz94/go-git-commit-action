package shared

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/somaz94/go-git-commit-action/internal/gitcmd"
)

// StageFiles adds the specified files to the Git staging area.
// It handles multiple file patterns separated by spaces.
func StageFiles(filePattern string) error {
	fmt.Printf("  • Adding files... ")

	for _, pattern := range strings.Fields(filePattern) {
		if err := executeGitAdd(pattern); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to add pattern %s: %v", pattern, err)
		}
	}

	fmt.Println("✅ Done")
	return nil
}

// CommitAndPush commits the staged changes and pushes them to the remote branch.
func CommitAndPush(commitMessage, branch string) error {
	// Commit
	fmt.Printf("  • Committing changes... ")
	commitCmd := exec.Command(gitcmd.CmdGit, gitcmd.CommitArgs(commitMessage)...)
	commitCmd.Stdout = os.Stdout
	commitCmd.Stderr = os.Stderr
	if err := commitCmd.Run(); err != nil {
		fmt.Println("❌ Failed")
		return fmt.Errorf("failed to commit: %v", err)
	}
	fmt.Println("✅ Done")

	// Push
	fmt.Printf("  • Pushing changes... ")
	pushCmd := exec.Command(gitcmd.CmdGit, gitcmd.PushUpstreamArgs(gitcmd.RefOrigin, branch)...)
	pushCmd.Stdout = os.Stdout
	pushCmd.Stderr = os.Stderr
	if err := pushCmd.Run(); err != nil {
		fmt.Println("❌ Failed")
		return fmt.Errorf("failed to push: %v", err)
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
