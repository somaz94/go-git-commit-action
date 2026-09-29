package git

import (
	"context"
	stderrors "errors"
	"fmt"
	"strconv"

	"github.com/somaz94/go-git-commit-action/internal/config"
	"github.com/somaz94/go-git-commit-action/internal/git/pr"
	"github.com/somaz94/go-git-commit-action/internal/git/shared"
	"github.com/somaz94/go-git-commit-action/internal/gitcmd"
	"github.com/somaz94/go-git-commit-action/internal/output"
)

// errPRStage marks a pull request step that still failed after its in-place
// retries, so withRetry returns it without rerunning the workflow.
var errPRStage = stderrors.New("pull request steps failed")

// newPRCreator lets tests point the Creator at a fake API.
var newPRCreator = pr.NewCreatorWithRunner

// CreatePullRequest is the main function to create a GitHub pull request.
// It handles the entire flow of preparing branches, creating the PR,
// and processing post-creation tasks like adding labels or closing the PR.
// Every step after the commit is retried in place, resuming where it failed.
func CreatePullRequest(ctx context.Context, r gitcmd.Runner, config *config.GitConfig, result *output.Result) error {
	fmt.Println("\nCreating Pull Request:")

	branchMgr := pr.NewBranchManagerWithRunner(config, r)
	var sourceBranch string
	if config.AutoBranch {
		// Commits and pushes a new branch, so only the outer retry may repeat it.
		branch, err := branchMgr.PrepareSourceBranch(ctx)
		if err != nil {
			return err
		}
		sourceBranch = branch
	}

	diffChecker := pr.NewDiffCheckerWithRunner(config, r)
	creator := newPRCreator(config, r)
	var posted *pr.PRResponse
	err := withRetry(ctx, config.RetryCount, func() error {
		if sourceBranch == "" {
			branch, err := branchMgr.PrepareSourceBranch(ctx)
			if err != nil {
				return err
			}
			sourceBranch = branch
		}

		response := posted
		if response == nil {
			fresh, err := openPullRequest(ctx, r, diffChecker, creator, result)
			if err != nil {
				return err
			}
			// Never POST again once GitHub has a PR for this head: if pr_closed
			// closed it, a repeat opens a duplicate.
			if (fresh.Message == "" && fresh.HTMLURL != "") || fresh.AlreadyExists() {
				posted = &fresh
			}
			response = &fresh
		}
		return creator.HandlePRResponse(ctx, *response, sourceBranch)
	})
	if err != nil {
		return fmt.Errorf("%w: %w", errPRStage, err)
	}

	// A 422 carries no PR fields; the lookup that resolved it does.
	if existing, ok := creator.ExistingPR(); ok {
		recordPROutputs(result, existing)
	}

	fmt.Println("\nGit Commit Action Completed Successfully!\n" +
		"=========================================")

	return nil
}

// openPullRequest checks the branch diff, sends the PR creation request and
// records its outputs.
func openPullRequest(ctx context.Context, r gitcmd.Runner, diffChecker *pr.DiffChecker, creator *pr.Creator, result *output.Result) (pr.PRResponse, error) {
	if err := diffChecker.CheckBranchDifferences(); err != nil {
		return pr.PRResponse{}, err
	}

	response, err := creator.CreatePullRequest(ctx)
	if err != nil {
		return pr.PRResponse{}, err
	}

	// Capture commit SHA (works for both auto-branch and manual branch flows)
	if commitSHA, err := shared.CurrentCommitSHA(r); err == nil {
		result.Set(output.KeyCommitSHA, commitSHA)
	}

	recordPROutputs(result, response)
	return response, nil
}

// recordPROutputs sets pr_url and pr_number from the fields the PR carries.
func recordPROutputs(result *output.Result, response pr.PRResponse) {
	if response.HTMLURL != "" {
		result.Set(output.KeyPRURL, response.HTMLURL)
	}
	if response.HasNumber {
		result.Set(output.KeyPRNumber, strconv.Itoa(response.Number))
	}
}
