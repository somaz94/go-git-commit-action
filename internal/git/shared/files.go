package shared

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/somaz94/go-git-commit-action/internal/gitcmd"
)

// RunStep executes a single command with the standard
// "  - <desc>... " → "Done" / "FAILED" progress feedback used across the
// git packages. On failure it prints "FAILED" and returns the raw error so
// the caller can wrap it with the appropriate typed error.
//
// It does NOT modify cmd.Stdout / cmd.Stderr — configure those on cmd before
// calling if the command output should stream to the console.
func RunStep(desc string, cmd *exec.Cmd) error {
	fmt.Printf("  - %s... ", desc)
	if err := cmd.Run(); err != nil {
		fmt.Println("FAILED")
		return err
	}
	fmt.Println("Done")
	return nil
}

// StageFiles adds the specified files to the Git staging area.
// It handles multiple file patterns separated by spaces.
func StageFiles(filePattern string) error {
	fmt.Printf("  - Adding files... ")

	for _, pattern := range strings.Fields(filePattern) {
		if err := executeGitAdd(pattern); err != nil {
			fmt.Println("FAILED")
			return fmt.Errorf("failed to add pattern %s: %w", pattern, err)
		}
	}

	fmt.Println("Done")
	return nil
}

// CommitPushOptions configures CommitAndPush behavior.
type CommitPushOptions struct {
	// SetUpstream pushes with "-u" to set the upstream tracking reference
	// (used when pushing a freshly created branch).
	SetUpstream bool
	// TolerateNothingToCommit treats a "nothing to commit" result (git commit
	// exit code 1) as success and proceeds to push, instead of failing. Used on
	// the direct-commit path where an empty commit must not abort the action.
	TolerateNothingToCommit bool
}

// isNothingToCommitExit reports whether err is a "git commit" exit-code-1
// failure, which git returns when there is nothing staged to commit.
func isNothingToCommitExit(err error) bool {
	var exitErr *exec.ExitError
	return errors.As(err, &exitErr) && exitErr.ExitCode() == 1
}

// CommitAndPush commits the staged changes and pushes them to the remote branch.
// Behavior is controlled by opts (upstream tracking and empty-commit tolerance).
func CommitAndPush(commitMessage, branch string, opts CommitPushOptions) error {
	// Commit
	commitCmd := exec.Command(gitcmd.CmdGit, gitcmd.CommitArgs(commitMessage)...)
	commitCmd.Stdout = os.Stdout
	commitCmd.Stderr = os.Stderr
	fmt.Printf("  - Committing changes... ")
	if err := commitCmd.Run(); err != nil {
		if opts.TolerateNothingToCommit && isNothingToCommitExit(err) {
			fmt.Println("[WARN] Nothing to commit, skipping...")
		} else {
			fmt.Println("FAILED")
			return fmt.Errorf("failed to commit: %w", err)
		}
	} else {
		fmt.Println("Done")
	}

	// Push
	pushArgs := gitcmd.PushArgs(gitcmd.RefOrigin, branch)
	if opts.SetUpstream {
		pushArgs = gitcmd.PushUpstreamArgs(gitcmd.RefOrigin, branch)
	}
	pushCmd := exec.Command(gitcmd.CmdGit, pushArgs...)
	pushCmd.Stdout = os.Stdout
	pushCmd.Stderr = os.Stderr
	if err := RunStep("Pushing changes", pushCmd); err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}

	return nil
}

// CurrentCommitSHA retrieves the current HEAD commit SHA.
func CurrentCommitSHA() (string, error) {
	out, err := exec.Command(gitcmd.CmdGit, gitcmd.RevParseArgs("HEAD")...).Output()
	if err != nil {
		return "", fmt.Errorf("get commit SHA: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// executeGitAdd executes the git add command for a specific pattern.
func executeGitAdd(pattern string) error {
	addCmd := exec.Command(gitcmd.CmdGit, gitcmd.AddArgs(pattern)...)
	addCmd.Stdout = os.Stdout
	addCmd.Stderr = os.Stderr
	return addCmd.Run()
}
