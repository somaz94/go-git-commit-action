package pr

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/somaz94/go-git-commit-action/internal/config"
	"github.com/somaz94/go-git-commit-action/internal/gitcmd"
)

// BranchManager handles branch operations for pull requests.
type BranchManager struct {
	config *config.GitConfig
}

// NewBranchManager creates a new BranchManager instance.
func NewBranchManager(cfg *config.GitConfig) *BranchManager {
	return &BranchManager{config: cfg}
}

// PrepareSourceBranch sets up the branch that will be used as the source for the PR.
// If auto_branch is enabled, it creates a new branch with a timestamp.
// Otherwise, it uses the specified PR branch.
func (bm *BranchManager) PrepareSourceBranch() (string, error) {
	if bm.config.AutoBranch {
		return bm.createAutoBranch()
	}
	return bm.checkoutExistingBranch()
}

// createAutoBranch creates a new branch with a timestamp and commits changes to it.
func (bm *BranchManager) createAutoBranch() (string, error) {
	// Create a branch name with a timestamp
	sourceBranch := fmt.Sprintf("update-files-%s", time.Now().Format("20060102-150405"))
	bm.config.PRBranch = sourceBranch

	// Create and switch to a new branch
	fmt.Printf("  • Creating new branch %s... ", sourceBranch)
	if err := exec.Command(gitcmd.CmdGit, gitcmd.CheckoutNewBranchArgs(sourceBranch)...).Run(); err != nil {
		fmt.Println("❌ Failed")
		return "", fmt.Errorf("failed to create branch: %v", err)
	}
	fmt.Println("✅ Done")

	// Stage files
	if err := stageFiles(bm.config.FilePattern); err != nil {
		return "", err
	}

	// Commit and push changes
	if err := commitAndPush(bm.config, sourceBranch); err != nil {
		return "", err
	}

	return sourceBranch, nil
}

// checkoutExistingBranch checks out the specified PR branch.
func (bm *BranchManager) checkoutExistingBranch() (string, error) {
	sourceBranch := bm.config.PRBranch
	fmt.Printf("  • Checking out branch %s... ", sourceBranch)
	if err := exec.Command(gitcmd.CmdGit, gitcmd.CheckoutArgs(sourceBranch)...).Run(); err != nil {
		fmt.Println("❌ Failed")
		return "", fmt.Errorf("failed to checkout branch: %v", err)
	}
	fmt.Println("✅ Done")

	return sourceBranch, nil
}

// DeleteSourceBranch deletes the source branch from remote.
func (bm *BranchManager) DeleteSourceBranch(sourceBranch string) error {
	// Skip if in dry run mode
	if bm.config.PRDryRun {
		fmt.Printf("\n  • [DRY RUN] Would delete source branch %s... ✅ Skipped\n", sourceBranch)
		return nil
	}

	// Only delete if auto-branch is enabled (safety check)
	if !bm.config.AutoBranch {
		return nil
	}

	fmt.Printf("\n  • Deleting source branch %s... ", sourceBranch)
	deleteCommand := exec.Command(gitcmd.CmdGit, gitcmd.SubCmdPush, gitcmd.RefOrigin, "--delete", sourceBranch)
	if err := deleteCommand.Run(); err != nil {
		fmt.Println("❌ Failed")
		return fmt.Errorf("failed to delete source branch %s: %v", sourceBranch, err)
	}

	fmt.Println("✅ Done")
	return nil
}

// FetchBranches fetches the latest from both the base and source branches.
func (bm *BranchManager) FetchBranches() error {
	fetchCommands := []struct {
		branch string
		desc   string
	}{
		{bm.config.PRBase, "Fetching base branch"},
		{bm.config.PRBranch, "Fetching source branch"},
	}

	for _, cmd := range fetchCommands {
		if err := exec.Command(gitcmd.CmdGit, gitcmd.FetchArgs(gitcmd.RefOrigin, cmd.branch)...).Run(); err != nil {
			return fmt.Errorf("%s: %v", cmd.desc, err)
		}
	}

	return nil
}

// stageFiles adds the specified files to the Git staging area.
// It handles multiple file patterns separated by spaces.
func stageFiles(filePattern string) error {
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

// commitAndPush commits the staged changes and pushes them to the remote branch.
func commitAndPush(cfg *config.GitConfig, branch string) error {
	// Commit
	fmt.Printf("  • Committing changes... ")
	commitCmd := exec.Command(gitcmd.CmdGit, gitcmd.CommitArgs(cfg.CommitMessage)...)
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
