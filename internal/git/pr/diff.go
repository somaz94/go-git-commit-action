package pr

import (
	"fmt"
	"os"

	"github.com/somaz94/go-git-commit-action/internal/config"
	"github.com/somaz94/go-git-commit-action/internal/gitcmd"
)

// DiffChecker handles change detection between branches.
type DiffChecker struct {
	config *config.GitConfig
	runner gitcmd.Runner
}

// NewDiffChecker creates a new DiffChecker instance.
func NewDiffChecker(cfg *config.GitConfig) *DiffChecker {
	return NewDiffCheckerWithRunner(cfg, gitcmd.NewExecRunner())
}

// NewDiffCheckerWithRunner creates a DiffChecker with an explicit command
// Runner, allowing tests to assert the emitted git commands.
func NewDiffCheckerWithRunner(cfg *config.GitConfig, r gitcmd.Runner) *DiffChecker {
	return &DiffChecker{config: cfg, runner: r}
}

// CheckBranchDifferences checks the differences between the PR base branch and the source branch.
// It also prints the compare URL for opening the PR by hand.
func (dc *DiffChecker) CheckBranchDifferences() error {
	fmt.Printf("\nChanged files between %s and %s:\n", dc.config.PRBase, dc.config.PRBranch)

	branchMgr := NewBranchManagerWithRunner(dc.config, dc.runner)
	if err := branchMgr.FetchBranches(); err != nil {
		return err
	}

	return dc.displayChangedFiles()
}

// displayChangedFiles shows the changed files between branches and validates if changes exist.
func (dc *DiffChecker) displayChangedFiles() error {
	filesOutput, err := dc.runner.Output(gitcmd.CmdGit, gitcmd.DiffNameStatusArgs(
		fmt.Sprintf("origin/%s", dc.config.PRBase),
		fmt.Sprintf("origin/%s", dc.config.PRBranch),
	)...)
	if err != nil {
		fmt.Printf("[WARN] Failed to get diff: %v\n", err)
	}

	if len(filesOutput) == 0 {
		fmt.Println("No changes detected")
		if dc.config.SkipIfEmpty {
			return nil
		}
		return fmt.Errorf("no changes to create PR")
	}

	fmt.Printf("%s\n", string(filesOutput))

	dc.displayPRURL()

	return nil
}

// displayPRURL shows the URL for manual PR creation.
func (dc *DiffChecker) displayPRURL() {
	fmt.Printf("\nBranch '%s' is ready for PR.\n", dc.config.PRBranch)
	prURL := fmt.Sprintf("https://github.com/%s/compare/%s...%s",
		os.Getenv("GITHUB_REPOSITORY"),
		dc.config.PRBase,
		dc.config.PRBranch)
	fmt.Printf("You can create a pull request by visiting:\n   %s\n", prURL)
}
