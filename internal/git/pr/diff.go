package pr

import (
	"fmt"
	"os"
	"strings"

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
// It also prints the compare URL for opening the PR by hand, except for a dry-run
// auto branch, which does not exist on the remote.
func (dc *DiffChecker) CheckBranchDifferences() error {
	fmt.Printf("\nChanged files between %s and %s:\n", dc.config.PRBase, dc.config.PRBranch)

	// A dry-run auto branch is never pushed, so list the uncommitted changes
	// through StageFiles' pathspecs: exactly what its commit would take.
	if dc.dryRunAutoBranch() {
		return dc.displayChangedFiles(gitcmd.StatusPorcelainArgs(strings.Fields(dc.config.FilePattern)...))
	}

	branchMgr := NewBranchManagerWithRunner(dc.config, dc.runner)
	if err := branchMgr.FetchBranches(); err != nil {
		return err
	}

	return dc.displayChangedFiles(gitcmd.DiffNameStatusArgs(
		fmt.Sprintf("origin/%s", dc.config.PRBase),
		fmt.Sprintf("origin/%s", dc.config.PRBranch),
	))
}

func (dc *DiffChecker) dryRunAutoBranch() bool {
	return dc.config.PRDryRun && dc.config.AutoBranch
}

// displayChangedFiles shows the files git lists for args and validates if changes exist.
func (dc *DiffChecker) displayChangedFiles(args []string) error {
	filesOutput, err := dc.runner.Output(gitcmd.CmdGit, args...)
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

	if !dc.dryRunAutoBranch() {
		dc.displayPRURL()
	}

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
