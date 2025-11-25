package pr

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/somaz94/go-git-commit-action/internal/config"
	"github.com/somaz94/go-git-commit-action/internal/errors"
	"github.com/somaz94/go-git-commit-action/internal/gitcmd"
)

// Creator handles pull request creation and management.
type Creator struct {
	config *config.GitConfig
}

// NewCreator creates a new Creator instance.
func NewCreator(cfg *config.GitConfig) *Creator {
	return &Creator{config: cfg}
}

// CreatePullRequest creates a GitHub pull request via API.
// It handles both dry run and actual PR creation.
func (c *Creator) CreatePullRequest() (map[string]interface{}, error) {
	if c.config.PRDryRun {
		return c.createDryRunPR()
	}
	return c.createActualPR()
}

// createDryRunPR simulates PR creation in dry run mode.
func (c *Creator) createDryRunPR() (map[string]interface{}, error) {
	fmt.Printf("  • [DRY RUN] Would create pull request from %s to %s... ", c.config.PRBranch, c.config.PRBase)
	fmt.Println("✅ Skipped (Dry Run mode)")

	// Create a mock response with the PR URL
	mockResponse := map[string]interface{}{
		"html_url": fmt.Sprintf("https://github.com/%s/compare/%s...%s?dry_run=1",
			os.Getenv("GITHUB_REPOSITORY"),
			c.config.PRBase,
			c.config.PRBranch),
		"number":  float64(0),
		"dry_run": true,
	}

	return mockResponse, nil
}

// createActualPR creates an actual pull request via GitHub API.
func (c *Creator) createActualPR() (map[string]interface{}, error) {
	fmt.Printf("  • Creating pull request from %s to %s... ", c.config.PRBranch, c.config.PRBase)

	// Prepare the PR data
	prData, err := c.preparePRData()
	if err != nil {
		return nil, err
	}

	// Call the GitHub API
	return c.callGitHubAPI(prData)
}

// preparePRData creates the data structure needed for the PR creation API call.
func (c *Creator) preparePRData() (map[string]interface{}, error) {
	// Get the GitHub Run ID for reference
	runID := os.Getenv("GITHUB_RUN_ID")

	// Get the current commit SHA
	commitSHA, err := getCurrentCommitSHA()
	if err != nil {
		return nil, err
	}

	// Generate title and body if not provided
	title, body := c.generatePRTitleAndBody(runID, commitSHA)

	// Create the PR data
	prData := map[string]interface{}{
		"title": title,
		"head":  c.config.PRBranch,
		"base":  c.config.PRBase,
		"body":  body,
	}

	return prData, nil
}

// getCurrentCommitSHA retrieves the current commit SHA.
func getCurrentCommitSHA() (string, error) {
	commitCmd := exec.Command(gitcmd.CmdGit, gitcmd.RevParseArgs("HEAD")...)
	commitSHA, err := commitCmd.Output()
	if err != nil {
		return "", errors.New("get commit SHA", err)
	}
	return strings.TrimSpace(string(commitSHA)), nil
}

// generatePRTitleAndBody creates default PR title and body if not specified.
func (c *Creator) generatePRTitleAndBody(runID string, commitSHA string) (string, string) {
	title := c.config.PRTitle
	if title == "" {
		title = fmt.Sprintf("Auto PR: %s to %s (Run ID: %s)", c.config.PRBranch, c.config.PRBase, runID)
	}

	body := c.config.PRBody
	if body == "" {
		body = fmt.Sprintf("Created by Go Git Commit Action\nSource: %s\nTarget: %s\nCommit: %s\nGitHub Run ID: %s",
			c.config.PRBranch, c.config.PRBase, commitSHA, runID)
	}

	return title, body
}

// callGitHubAPI makes the actual GitHub API call to create a PR.
func (c *Creator) callGitHubAPI(prData map[string]interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(prData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal PR data: %v", err)
	}

	// Create a pull request using the GitHub API via curl
	curlCmd := exec.Command("curl", "-s", "-X", "POST",
		"-H", fmt.Sprintf("Authorization: Bearer %s", c.config.GitHubToken),
		"-H", "Accept: application/vnd.github+json",
		"-H", "Content-Type: application/json",
		fmt.Sprintf("https://api.github.com/repos/%s/pulls", os.Getenv("GITHUB_REPOSITORY")),
		"-d", string(jsonData))

	output, err := curlCmd.CombinedOutput()
	if err != nil {
		fmt.Println("⚠️  Failed to create PR automatically")
		fmt.Printf("Error: %v\n", err)
		fmt.Printf("Response: %s\n", string(output))
		return nil, fmt.Errorf("failed to execute curl command: %v", err)
	}

	// Parse the response JSON
	var response map[string]interface{}
	if err := json.Unmarshal(output, &response); err != nil {
		fmt.Printf("Raw response: %s\n", string(output))
		return nil, errors.New("parse PR response", err)
	}

	return response, nil
}

// HandlePRResponse processes the PR creation response and performs follow-up actions.
func (c *Creator) HandlePRResponse(response map[string]interface{}, sourceBranch string) error {
	// Handle dry run responses separately
	if dryRun, ok := response["dry_run"].(bool); ok && dryRun {
		return c.handleDryRunResponse(response)
	}

	// Check for error messages in the response
	if errMsg, ok := response["message"].(string); ok {
		return c.handleErrorResponse(response, errMsg)
	}

	// Handle successful PR creation
	return c.handleSuccessfulPR(response, sourceBranch)
}

// handleDryRunResponse handles the dry run response without making actual changes.
func (c *Creator) handleDryRunResponse(response map[string]interface{}) error {
	fmt.Printf("\n🔍 [DRY RUN] Pull request would be created at: %s\n", response["html_url"])
	fmt.Printf("👉 No actual PR was created (dry run mode)\n")

	// Print mock PR details
	fmt.Printf("\n📋 PR details that would be submitted:\n")
	fmt.Printf("  • Title: %s\n", c.config.PRTitle)
	fmt.Printf("  • Source branch: %s\n", c.config.PRBranch)
	fmt.Printf("  • Target branch: %s\n", c.config.PRBase)

	if len(c.config.PRLabels) > 0 {
		fmt.Printf("  • Labels: %s\n", strings.Join(c.config.PRLabels, ", "))
	}

	if c.config.PRClosed {
		fmt.Printf("  • Would be closed immediately: Yes\n")
	}

	if c.config.DeleteSourceBranch {
		if c.config.AutoBranch {
			fmt.Printf("  • Source branch would be deleted: Yes (auto-generated branch)\n")
		} else {
			fmt.Printf("  • Warning: delete_source_branch is set but requires auto_branch:true to take effect\n")
		}
	}

	return nil
}

// handleErrorResponse processes error responses from the GitHub API.
func (c *Creator) handleErrorResponse(response map[string]interface{}, errMsg string) error {
	fmt.Printf("GitHub API Error: %s\n", errMsg)

	// Extract detailed error information if available
	if errors, ok := response["errors"].([]interface{}); ok {
		fmt.Println("Error details:")
		for _, err := range errors {
			if errMap, ok := err.(map[string]interface{}); ok {
				fmt.Printf("  • %v\n", errMap)

				// Special handling for when a PR already exists
				if message, ok := errMap["message"].(string); ok &&
					strings.Contains(message, "A pull request already exists") {
					return c.handleExistingPR()
				}
			}
		}
	}

	return fmt.Errorf("GitHub API error: %s", errMsg)
}

// handleSuccessfulPR processes a successful PR creation response.
func (c *Creator) handleSuccessfulPR(response map[string]interface{}, sourceBranch string) error {
	// Extract the PR URL and display it
	if htmlURL, ok := response["html_url"].(string); ok {
		fmt.Println("✅ Done")
		fmt.Printf("Pull request created: %s\n", htmlURL)

		// Process the PR number for additional operations
		if number, ok := response["number"].(float64); ok {
			prNumber := int(number)

			// Add labels if specified
			if len(c.config.PRLabels) > 0 {
				if err := c.addLabelsToIssue(prNumber); err != nil {
					return err
				}
			}

			// Close the PR if specified
			if c.config.PRClosed {
				if err := c.closePullRequest(prNumber); err != nil {
					return err
				}
			}
		}
	} else {
		fmt.Println("⚠️  Failed to create PR")
		fmt.Printf("Response: %v\n", response)
		return fmt.Errorf("failed to get PR URL from response")
	}

	// Delete the source branch if auto-branch and delete-source-branch are enabled
	if c.config.DeleteSourceBranch && c.config.AutoBranch {
		branchMgr := NewBranchManager(c.config)
		if err := branchMgr.DeleteSourceBranch(sourceBranch); err != nil {
			return err
		}
	}

	return nil
}

// handleExistingPR processes the case when a PR already exists.
func (c *Creator) handleExistingPR() error {
	fmt.Println("⚠️  Pull request already exists")

	// Find the existing PR
	prs, err := c.findExistingPRs()
	if err != nil {
		return err
	}

	// If we found existing PRs, process them
	if len(prs) > 0 {
		if number, ok := prs[0]["number"].(float64); ok {
			prNumber := int(number)
			fmt.Printf("Found existing PR #%d\n", prNumber)

			// Process the existing PR (labels, closing)
			return c.processExistingPR(prNumber)
		}
	}

	return nil
}

// findExistingPRs searches for existing PRs with the same head and base.
func (c *Creator) findExistingPRs() ([]map[string]interface{}, error) {
	searchCmd := exec.Command("curl", "-s",
		"-H", fmt.Sprintf("Authorization: Bearer %s", c.config.GitHubToken),
		"-H", "Accept: application/vnd.github+json",
		"-H", "X-GitHub-Api-Version: 2022-11-28",
		fmt.Sprintf("https://api.github.com/repos/%s/pulls?head=%s&base=%s",
			os.Getenv("GITHUB_REPOSITORY"),
			c.config.PRBranch,
			c.config.PRBase))

	searchOutput, err := searchCmd.CombinedOutput()
	if err != nil {
		return nil, errors.New("search existing PRs", err)
	}

	var prs []map[string]interface{}
	if err := json.Unmarshal(searchOutput, &prs); err != nil {
		return nil, errors.New("parse existing PR search response", err)
	}

	return prs, nil
}

// processExistingPR applies operations like adding labels or closing to an existing PR.
func (c *Creator) processExistingPR(prNumber int) error {
	// Add labels if specified
	if len(c.config.PRLabels) > 0 {
		if err := c.addLabelsToIssue(prNumber); err != nil {
			return err
		}
	}

	// Close the PR if specified
	if c.config.PRClosed {
		if err := c.closePullRequest(prNumber); err != nil {
			return err
		}
	}

	return nil
}

// addLabelsToIssue adds labels to an issue/PR.
func (c *Creator) addLabelsToIssue(prNumber int) error {
	// Skip if in dry run mode
	if c.config.PRDryRun {
		fmt.Printf("  • [DRY RUN] Would add labels %v to PR #%d... ✅ Skipped\n", c.config.PRLabels, prNumber)
		return nil
	}

	fmt.Printf("  • Adding labels to PR #%d... ", prNumber)

	// Prepare the labels data
	labelsData := map[string]interface{}{
		"labels": c.config.PRLabels,
	}
	jsonLabelsData, err := json.Marshal(labelsData)
	if err != nil {
		return errors.New("marshal labels data", err)
	}

	// Execute the API call
	labelsCurlCmd := exec.Command("curl", "-s", "-X", "POST",
		"-H", fmt.Sprintf("Authorization: Bearer %s", c.config.GitHubToken),
		"-H", "Accept: application/vnd.github+json",
		"-H", "Content-Type: application/json",
		fmt.Sprintf("https://api.github.com/repos/%s/issues/%d/labels",
			os.Getenv("GITHUB_REPOSITORY"), prNumber),
		"-d", string(jsonLabelsData))

	labelsOutput, err := labelsCurlCmd.CombinedOutput()
	if err != nil {
		fmt.Println("❌ Failed")
		fmt.Printf("Error: %v\n", err)
		fmt.Printf("Response: %s\n", string(labelsOutput))
		return errors.NewAPIError("add labels", string(labelsOutput))
	}

	fmt.Println("✅ Done")
	return nil
}

// closePullRequest closes a pull request.
func (c *Creator) closePullRequest(prNumber int) error {
	// Skip if in dry run mode
	if c.config.PRDryRun {
		fmt.Printf("  • [DRY RUN] Would close pull request #%d... ✅ Skipped\n", prNumber)
		return nil
	}

	fmt.Printf("  • Closing pull request #%d... ", prNumber)

	// Prepare the close data
	closeData := map[string]string{
		"state": "closed",
	}
	jsonCloseData, err := json.Marshal(closeData)
	if err != nil {
		return errors.New("marshal close data", err)
	}

	// Execute the API call
	closeCurlCmd := exec.Command("curl", "-s", "-X", "PATCH",
		"-H", fmt.Sprintf("Authorization: Bearer %s", c.config.GitHubToken),
		"-H", "Accept: application/vnd.github+json",
		"-H", "Content-Type: application/json",
		fmt.Sprintf("https://api.github.com/repos/%s/pulls/%d",
			os.Getenv("GITHUB_REPOSITORY"), prNumber),
		"-d", string(jsonCloseData))

	closeOutput, err := closeCurlCmd.CombinedOutput()
	if err != nil {
		fmt.Println("❌ Failed")
		fmt.Printf("Error: %v\n", err)
		fmt.Printf("Response: %s\n", string(closeOutput))
		return errors.NewAPIError("close PR", string(closeOutput))
	}

	fmt.Println("✅ Done")
	return nil
}
