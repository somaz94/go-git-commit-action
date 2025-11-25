package pr

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/somaz94/go-git-commit-action/internal/config"
	"github.com/somaz94/go-git-commit-action/internal/gitcmd"
)

// DiffChecker handles change detection between branches.
type DiffChecker struct {
	config *config.GitConfig
}

// NewDiffChecker creates a new DiffChecker instance.
func NewDiffChecker(cfg *config.GitConfig) *DiffChecker {
	return &DiffChecker{config: cfg}
}

// CheckBranchDifferences checks the differences between the PR base branch and the source branch.
// It also shows the potential PR URL for manual creation if the API fails.
func (dc *DiffChecker) CheckBranchDifferences() error {
	fmt.Printf("\n📊 Changed files between %s and %s:\n", dc.config.PRBase, dc.config.PRBranch)

	// Fetch the latest from both branches
	branchMgr := NewBranchManager(dc.config)
	if err := branchMgr.FetchBranches(); err != nil {
		return err
	}

	// Display the changed files
	return dc.displayChangedFiles()
}

// displayChangedFiles shows the changed files between branches and validates if changes exist.
func (dc *DiffChecker) displayChangedFiles() error {
	// Check the changed files
	diffFiles := exec.Command(gitcmd.CmdGit, gitcmd.DiffNameStatusArgs(
		fmt.Sprintf("origin/%s", dc.config.PRBase),
		fmt.Sprintf("origin/%s", dc.config.PRBranch),
	)...)
	filesOutput, _ := diffFiles.Output()

	if len(filesOutput) == 0 {
		fmt.Println("No changes detected")
		if dc.config.SkipIfEmpty {
			return nil
		}
		return fmt.Errorf("no changes to create PR")
	}

	fmt.Printf("%s\n", string(filesOutput))

	// Display the PR URL for manual creation if needed
	dc.displayPRURL()

	return nil
}

// displayPRURL shows the URL for manual PR creation.
func (dc *DiffChecker) displayPRURL() {
	fmt.Printf("\n✅ Branch '%s' is ready for PR.\n", dc.config.PRBranch)
	prURL := fmt.Sprintf("https://github.com/%s/compare/%s...%s",
		os.Getenv("GITHUB_REPOSITORY"),
		dc.config.PRBase,
		dc.config.PRBranch)
	fmt.Printf("✅ You can create a pull request by visiting:\n   %s\n", prURL)
}
