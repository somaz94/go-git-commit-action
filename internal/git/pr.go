package git

import (
	"context"
	"fmt"
	"strconv"

	"github.com/somaz94/go-git-commit-action/internal/config"
	"github.com/somaz94/go-git-commit-action/internal/git/pr"
	"github.com/somaz94/go-git-commit-action/internal/git/shared"
	"github.com/somaz94/go-git-commit-action/internal/gitcmd"
	"github.com/somaz94/go-git-commit-action/internal/output"
)

// CreatePullRequest is the main function to create a GitHub pull request.
// It handles the entire flow of preparing branches, creating the PR,
// and processing post-creation tasks like adding labels or closing the PR.
func CreatePullRequest(ctx context.Context, r gitcmd.Runner, config *config.GitConfig, result *output.Result) error {
	fmt.Println("\nCreating Pull Request:")

	branchMgr := pr.NewBranchManagerWithRunner(config, r)
	sourceBranch, err := branchMgr.PrepareSourceBranch()
	if err != nil {
		return err
	}

	diffChecker := pr.NewDiffCheckerWithRunner(config, r)
	if err := diffChecker.CheckBranchDifferences(); err != nil {
		return err
	}

	creator := pr.NewCreatorWithRunner(config, r)
	prResponse, err := creator.CreatePullRequest(ctx)
	if err != nil {
		return err
	}

	// Capture commit SHA (works for both auto-branch and manual branch flows)
	if commitSHA, err := shared.CurrentCommitSHA(r); err == nil {
		result.Set(output.KeyCommitSHA, commitSHA)
	}

	if prResponse.HTMLURL != "" {
		result.Set(output.KeyPRURL, prResponse.HTMLURL)
	}
	if prResponse.HasNumber {
		result.Set(output.KeyPRNumber, strconv.Itoa(prResponse.Number))
	}

	if err := creator.HandlePRResponse(ctx, prResponse, sourceBranch); err != nil {
		return err
	}

	fmt.Println("\nGit Commit Action Completed Successfully!\n" +
		"=========================================")

	return nil
}
