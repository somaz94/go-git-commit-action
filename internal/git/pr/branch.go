package pr

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/somaz94/go-git-commit-action/internal/config"
	"github.com/somaz94/go-git-commit-action/internal/git/shared"
	"github.com/somaz94/go-git-commit-action/internal/gitcmd"
)

const (
	timestampFormat = "20060102-150405"
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
	sourceBranch := fmt.Sprintf("update-files-%s", time.Now().Format(timestampFormat))
	bm.config.PRBranch = sourceBranch

	// Create and switch to a new branch
	fmt.Printf("  - Creating new branch %s... ", sourceBranch)
	if err := exec.Command(gitcmd.CmdGit, gitcmd.CheckoutNewBranchArgs(sourceBranch)...).Run(); err != nil {
		fmt.Println("FAILED")
		return "", fmt.Errorf("failed to create branch: %v", err)
	}
	fmt.Println("Done")

	// Stage files using shared utility
	if err := shared.StageFiles(bm.config.FilePattern); err != nil {
		return "", err
	}

	// Commit and push using shared utility
	if err := shared.CommitAndPush(bm.config.CommitMessage, sourceBranch); err != nil {
		return "", err
	}

	return sourceBranch, nil
}

// checkoutExistingBranch checks out the specified PR branch.
func (bm *BranchManager) checkoutExistingBranch() (string, error) {
	sourceBranch := bm.config.PRBranch
	fmt.Printf("  - Checking out branch %s... ", sourceBranch)
	if err := exec.Command(gitcmd.CmdGit, gitcmd.CheckoutArgs(sourceBranch)...).Run(); err != nil {
		fmt.Println("FAILED")
		return "", fmt.Errorf("failed to checkout branch: %v", err)
	}
	fmt.Println("Done")

	return sourceBranch, nil
}

// DeleteSourceBranch deletes the source branch from remote.
func (bm *BranchManager) DeleteSourceBranch(sourceBranch string) error {
	if bm.config.PRDryRun {
		fmt.Printf("\n  - [DRY RUN] Would delete source branch %s... Skipped\n", sourceBranch)
		return nil
	}

	// Only delete if auto-branch is enabled (safety check)
	if !bm.config.AutoBranch {
		return nil
	}

	fmt.Printf("\n  - Deleting source branch %s... ", sourceBranch)
	deleteCommand := exec.Command(gitcmd.CmdGit, gitcmd.SubCmdPush, gitcmd.RefOrigin, "--delete", sourceBranch)
	if err := deleteCommand.Run(); err != nil {
		fmt.Println("FAILED")
		return fmt.Errorf("failed to delete source branch %s: %v", sourceBranch, err)
	}

	fmt.Println("Done")
	return nil
}

// FetchBranches fetches the latest from both the base and source branches.
func (bm *BranchManager) FetchBranches() error {
	branches := []struct {
		branch string
		desc   string
	}{
		{bm.config.PRBase, "Fetching base branch"},
		{bm.config.PRBranch, "Fetching source branch"},
	}

	for _, b := range branches {
		if err := exec.Command(gitcmd.CmdGit, gitcmd.FetchArgs(gitcmd.RefOrigin, b.branch)...).Run(); err != nil {
			return fmt.Errorf("%s: %v", b.desc, err)
		}
	}

	return nil
}
