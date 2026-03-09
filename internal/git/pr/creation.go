package pr

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/somaz94/go-git-commit-action/internal/config"
	"github.com/somaz94/go-git-commit-action/internal/errors"
	"github.com/somaz94/go-git-commit-action/internal/gitcmd"
	"github.com/somaz94/go-git-commit-action/internal/github"
)

// Creator handles pull request creation and management.
type Creator struct {
	config *config.GitConfig
	client *github.Client
}

// NewCreator creates a new Creator instance.
func NewCreator(cfg *config.GitConfig) *Creator {
	return &Creator{
		config: cfg,
		client: github.NewClient(cfg.GitHubToken),
	}
}

// CreatePullRequest creates a GitHub pull request via API.
func (c *Creator) CreatePullRequest() (map[string]interface{}, error) {
	if c.config.PRDryRun {
		return c.createDryRunPR()
	}
	return c.createActualPR()
}

// createDryRunPR simulates PR creation in dry run mode.
func (c *Creator) createDryRunPR() (map[string]interface{}, error) {
	fmt.Printf("  - [DRY RUN] Would create pull request from %s to %s... ", c.config.PRBranch, c.config.PRBase)
	fmt.Println("Skipped (Dry Run mode)")

	mockResponse := map[string]interface{}{
		"html_url": fmt.Sprintf("https://github.com/%s/compare/%s...%s?dry_run=1",
			c.client.Repo(),
			c.config.PRBase,
			c.config.PRBranch),
		"number":  float64(0),
		"dry_run": true,
	}

	return mockResponse, nil
}

// createActualPR creates an actual pull request via GitHub API.
func (c *Creator) createActualPR() (map[string]interface{}, error) {
	fmt.Printf("  - Creating pull request from %s to %s... ", c.config.PRBranch, c.config.PRBase)

	prData, err := c.preparePRData()
	if err != nil {
		return nil, err
	}

	return c.client.Post("/pulls", prData)
}

// preparePRData creates the data structure needed for the PR creation API call.
func (c *Creator) preparePRData() (map[string]interface{}, error) {
	runID := os.Getenv("GITHUB_RUN_ID")

	commitSHA, err := getCurrentCommitSHA()
	if err != nil {
		return nil, err
	}

	title, body := c.generatePRTitleAndBody(runID, commitSHA)

	data := map[string]interface{}{
		"title": title,
		"head":  c.config.PRBranch,
		"base":  c.config.PRBase,
		"body":  body,
	}

	if c.config.PRDraft {
		data["draft"] = true
	}

	return data, nil
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

// HandlePRResponse processes the PR creation response and performs follow-up actions.
func (c *Creator) HandlePRResponse(response map[string]interface{}, sourceBranch string) error {
	if dryRun, ok := response["dry_run"].(bool); ok && dryRun {
		return c.handleDryRunResponse(response)
	}

	if errMsg, ok := response["message"].(string); ok {
		return c.handleErrorResponse(response, errMsg)
	}

	return c.handleSuccessfulPR(response, sourceBranch)
}

// handleDryRunResponse handles the dry run response without making actual changes.
func (c *Creator) handleDryRunResponse(response map[string]interface{}) error {
	fmt.Printf("\n[DRY RUN] Pull request would be created at: %s\n", response["html_url"])
	fmt.Printf("No actual PR was created (dry run mode)\n")

	fmt.Printf("\nPR details that would be submitted:\n")
	fmt.Printf("  - Title: %s\n", c.config.PRTitle)
	fmt.Printf("  - Source branch: %s\n", c.config.PRBranch)
	fmt.Printf("  - Target branch: %s\n", c.config.PRBase)

	if len(c.config.PRLabels) > 0 {
		fmt.Printf("  - Labels: %s\n", strings.Join(c.config.PRLabels, ", "))
	}

	if c.config.PRDraft {
		fmt.Printf("  - Draft PR: Yes\n")
	}

	if len(c.config.PRReviewers) > 0 {
		fmt.Printf("  - Reviewers: %s\n", strings.Join(c.config.PRReviewers, ", "))
	}

	if len(c.config.PRAssignees) > 0 {
		fmt.Printf("  - Assignees: %s\n", strings.Join(c.config.PRAssignees, ", "))
	}

	if c.config.PRClosed {
		fmt.Printf("  - Would be closed immediately: Yes\n")
	}

	if c.config.DeleteSourceBranch {
		if c.config.AutoBranch {
			fmt.Printf("  - Source branch would be deleted: Yes (auto-generated branch)\n")
		} else {
			fmt.Printf("  - Warning: delete_source_branch is set but requires auto_branch:true to take effect\n")
		}
	}

	return nil
}

// handleErrorResponse processes error responses from the GitHub API.
func (c *Creator) handleErrorResponse(response map[string]interface{}, errMsg string) error {
	fmt.Printf("GitHub API Error: %s\n", errMsg)

	if apiErrors, ok := response["errors"].([]interface{}); ok {
		fmt.Println("Error details:")
		for _, err := range apiErrors {
			if errMap, ok := err.(map[string]interface{}); ok {
				fmt.Printf("  - %v\n", errMap)

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
	htmlURL, ok := response["html_url"].(string)
	if !ok {
		fmt.Println("[WARN] Failed to create PR")
		fmt.Printf("Response: %v\n", response)
		return fmt.Errorf("failed to get PR URL from response")
	}

	fmt.Println("Done")
	fmt.Printf("Pull request created: %s\n", htmlURL)

	if number, ok := response["number"].(float64); ok {
		if err := c.processExistingPR(int(number)); err != nil {
			return err
		}
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
	fmt.Println("[WARN] Pull request already exists")

	endpoint := fmt.Sprintf("/pulls?head=%s&base=%s", c.config.PRBranch, c.config.PRBase)
	prs, err := c.client.GetArray(endpoint)
	if err != nil {
		return err
	}

	if len(prs) > 0 {
		if number, ok := prs[0]["number"].(float64); ok {
			prNumber := int(number)
			fmt.Printf("Found existing PR #%d\n", prNumber)
			return c.processExistingPR(prNumber)
		}
	}

	return nil
}

// processExistingPR applies operations like adding labels, reviewers, assignees, or closing to an existing PR.
func (c *Creator) processExistingPR(prNumber int) error {
	if len(c.config.PRLabels) > 0 {
		if err := c.addLabelsToIssue(prNumber); err != nil {
			return err
		}
	}

	if len(c.config.PRReviewers) > 0 {
		if err := c.requestReviewers(prNumber); err != nil {
			return err
		}
	}

	if len(c.config.PRAssignees) > 0 {
		if err := c.addAssignees(prNumber); err != nil {
			return err
		}
	}

	if c.config.PRClosed {
		if err := c.closePullRequest(prNumber); err != nil {
			return err
		}
	}

	return nil
}

// addLabelsToIssue adds labels to an issue/PR.
func (c *Creator) addLabelsToIssue(prNumber int) error {
	if c.config.PRDryRun {
		fmt.Printf("  - [DRY RUN] Would add labels %v to PR #%d... Skipped\n", c.config.PRLabels, prNumber)
		return nil
	}

	fmt.Printf("  - Adding labels to PR #%d... ", prNumber)

	endpoint := fmt.Sprintf("/issues/%d/labels", prNumber)
	labelsData := map[string]interface{}{
		"labels": c.config.PRLabels,
	}

	if _, err := c.client.Post(endpoint, labelsData); err != nil {
		fmt.Println("FAILED")
		return errors.NewAPIError("add labels", err.Error())
	}

	fmt.Println("Done")
	return nil
}

// requestReviewers requests reviewers for a pull request.
func (c *Creator) requestReviewers(prNumber int) error {
	if c.config.PRDryRun {
		fmt.Printf("  - [DRY RUN] Would request reviewers %v for PR #%d... Skipped\n", c.config.PRReviewers, prNumber)
		return nil
	}

	fmt.Printf("  - Requesting reviewers for PR #%d... ", prNumber)

	endpoint := fmt.Sprintf("/pulls/%d/requested_reviewers", prNumber)
	data := map[string]interface{}{
		"reviewers": c.config.PRReviewers,
	}

	if _, err := c.client.Post(endpoint, data); err != nil {
		fmt.Println("FAILED")
		return errors.NewAPIError("request reviewers", err.Error())
	}

	fmt.Println("Done")
	return nil
}

// addAssignees adds assignees to a pull request.
func (c *Creator) addAssignees(prNumber int) error {
	if c.config.PRDryRun {
		fmt.Printf("  - [DRY RUN] Would add assignees %v to PR #%d... Skipped\n", c.config.PRAssignees, prNumber)
		return nil
	}

	fmt.Printf("  - Adding assignees to PR #%d... ", prNumber)

	endpoint := fmt.Sprintf("/issues/%d/assignees", prNumber)
	data := map[string]interface{}{
		"assignees": c.config.PRAssignees,
	}

	if _, err := c.client.Post(endpoint, data); err != nil {
		fmt.Println("FAILED")
		return errors.NewAPIError("add assignees", err.Error())
	}

	fmt.Println("Done")
	return nil
}

// closePullRequest closes a pull request.
func (c *Creator) closePullRequest(prNumber int) error {
	if c.config.PRDryRun {
		fmt.Printf("  - [DRY RUN] Would close pull request #%d... Skipped\n", prNumber)
		return nil
	}

	fmt.Printf("  - Closing pull request #%d... ", prNumber)

	endpoint := fmt.Sprintf("/pulls/%d", prNumber)
	closeData := map[string]string{
		"state": "closed",
	}

	if _, err := c.client.Patch(endpoint, closeData); err != nil {
		fmt.Println("FAILED")
		return errors.NewAPIError("close PR", err.Error())
	}

	fmt.Println("Done")
	return nil
}
