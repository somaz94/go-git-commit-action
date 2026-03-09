package git

import (
	"fmt"
	"strconv"

	"github.com/somaz94/go-git-commit-action/internal/config"
	"github.com/somaz94/go-git-commit-action/internal/git/pr"
	"github.com/somaz94/go-git-commit-action/internal/output"
)

// CreatePullRequest is the main function to create a GitHub pull request.
// It handles the entire flow of preparing branches, creating the PR,
// and processing post-creation tasks like adding labels or closing the PR.
func CreatePullRequest(config *config.GitConfig, result *output.Result) error {
	fmt.Println("\nCreating Pull Request:")

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

	// Capture PR outputs
	if htmlURL, ok := prResponse["html_url"].(string); ok {
		result.Set(output.KeyPRURL, htmlURL)
	}
	if number, ok := prResponse["number"].(float64); ok {
		result.Set(output.KeyPRNumber, strconv.Itoa(int(number)))
	}

	// Step 4: Process the PR response (labels, closing, etc.)
	if err := creator.HandlePRResponse(prResponse, sourceBranch); err != nil {
		return err
	}

	fmt.Println("\nGit Commit Action Completed Successfully!\n" +
		"=========================================")

	return nil
}
