package pr

import (
	"fmt"
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
	runner gitcmd.Runner
}

// NewBranchManager creates a new BranchManager instance.
func NewBranchManager(cfg *config.GitConfig) *BranchManager {
	return NewBranchManagerWithRunner(cfg, gitcmd.NewExecRunner())
}

// NewBranchManagerWithRunner creates a BranchManager with an explicit command
// Runner, allowing tests to assert the emitted git commands.
func NewBranchManagerWithRunner(cfg *config.GitConfig, r gitcmd.Runner) *BranchManager {
	return &BranchManager{config: cfg, runner: r}
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
	if err := shared.RunStep(bm.runner, fmt.Sprintf("Creating new branch %s", sourceBranch),
		gitcmd.CmdGit, gitcmd.CheckoutNewBranchArgs(sourceBranch)...); err != nil {
		return "", fmt.Errorf("failed to create branch: %w", err)
	}

	// Stage files using shared utility
	if err := shared.StageFiles(bm.runner, bm.config.FilePattern); err != nil {
		return "", err
	}

	// Commit and push using shared utility (new branch — set upstream tracking)
	if err := shared.CommitAndPush(bm.runner, bm.config.CommitMessage, sourceBranch,
		shared.CommitPushOptions{SetUpstream: true}); err != nil {
		return "", err
	}

	return sourceBranch, nil
}

// checkoutExistingBranch checks out the specified PR branch.
func (bm *BranchManager) checkoutExistingBranch() (string, error) {
	sourceBranch := bm.config.PRBranch
	if err := shared.RunStep(bm.runner, fmt.Sprintf("Checking out branch %s", sourceBranch),
		gitcmd.CmdGit, gitcmd.CheckoutArgs(sourceBranch)...); err != nil {
		return "", fmt.Errorf("failed to checkout branch: %w", err)
	}

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

	fmt.Println()
	if err := shared.RunStep(bm.runner, fmt.Sprintf("Deleting source branch %s", sourceBranch),
		gitcmd.CmdGit, gitcmd.PushDeleteBranchArgs(gitcmd.RefOrigin, sourceBranch)...); err != nil {
		return fmt.Errorf("failed to delete source branch %s: %w", sourceBranch, err)
	}

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
		if err := bm.runner.Run(gitcmd.CmdGit, gitcmd.FetchArgs(gitcmd.RefOrigin, b.branch)...); err != nil {
			return fmt.Errorf("%s: %w", b.desc, err)
		}
	}

	return nil
}
