package git

import (
	"fmt"

	"github.com/somaz94/go-git-commit-action/internal/config"
	"github.com/somaz94/go-git-commit-action/internal/git/pr"
)

// CreatePullRequest is the main function to create a GitHub pull request.
// It handles the entire flow of preparing branches, creating the PR,
// and processing post-creation tasks like adding labels or closing the PR.
func CreatePullRequest(config *config.GitConfig) error {
	fmt.Println("\n🔄 Creating Pull Request:")

	// Step 1: Prepare the source branch
	branchMgr := pr.NewBranchManager(config)
	sourceBranch, err := branchMgr.PrepareSourceBranch()
	if err != nil {
		return err
	}

	// Step 2: Check for differences between branches
	diffChecker := pr.NewDiffChecker(config)
	if err := diffChecker.CheckBranchDifferences(); err != nil {
		return err
	}

	// Step 3: Create the actual pull request via GitHub API
	creator := pr.NewCreator(config)
	prResponse, err := creator.CreatePullRequest()
	if err != nil {
		return err
	}

	// Step 4: Process the PR response (labels, closing, etc.)
	if err := creator.HandlePRResponse(prResponse, sourceBranch); err != nil {
		return err
	}

	fmt.Println("\n✨ Git Commit Action Completed Successfully!\n" +
		"=========================================")

	return nil
}
