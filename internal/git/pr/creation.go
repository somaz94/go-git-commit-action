package pr

import (
	"fmt"
	"os"
	"strings"

	"github.com/somaz94/go-git-commit-action/internal/config"
	"github.com/somaz94/go-git-commit-action/internal/errors"
	"github.com/somaz94/go-git-commit-action/internal/git/shared"
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

// PRResponse is the typed view of a GitHub PR-creation API response, or of the
// internal dry-run mock. It models only the fields the PR path consumes; any
// other API fields are dropped. Errors is left as []any so the diagnostic
// "  - %v" detail print stays byte-identical to the pre-refactor map form.
type PRResponse struct {
	HTMLURL   string // "html_url"; "" when absent (error responses)
	Number    int    // "number"; meaningful only when HasNumber is true
	HasNumber bool   // true when the response carried a numeric "number"
	DryRun    bool   // internal marker; set by the dry-run path, never from the API
	Message   string // "message"; non-empty on API error responses
	Errors    []any  // "errors"; nil when the key is absent (presence == old ",ok")
}

// parsePRResponse decodes the generic client map into a typed PRResponse.
// Presence semantics mirror the pre-refactor ",ok" type assertions: the client
// always yields JSON numbers as float64, and an absent key leaves the zero value.
func parsePRResponse(m map[string]any) PRResponse {
	var r PRResponse
	if v, ok := m["html_url"].(string); ok {
		r.HTMLURL = v
	}
	if v, ok := m["number"].(float64); ok {
		r.Number = int(v)
		r.HasNumber = true
	}
	if v, ok := m["dry_run"].(bool); ok {
		r.DryRun = v
	}
	if v, ok := m["message"].(string); ok {
		r.Message = v
	}
	if v, ok := m["errors"].([]any); ok {
		r.Errors = v
	}
	return r
}

// CreatePullRequest creates a GitHub pull request via API.
func (c *Creator) CreatePullRequest() (PRResponse, error) {
	if c.config.PRDryRun {
		return c.createDryRunPR()
	}
	return c.createActualPR()
}

// createDryRunPR simulates PR creation in dry run mode.
func (c *Creator) createDryRunPR() (PRResponse, error) {
	fmt.Printf("  - [DRY RUN] Would create pull request from %s to %s... ", c.config.PRBranch, c.config.PRBase)
	fmt.Println("Skipped (Dry Run mode)")

	return PRResponse{
		HTMLURL: fmt.Sprintf("https://github.com/%s/compare/%s...%s?dry_run=1",
			c.client.Repo(),
			c.config.PRBase,
			c.config.PRBranch),
		Number:    0,
		HasNumber: true, // preserves pre-refactor pr_number="0" output in dry-run
		DryRun:    true,
	}, nil
}

// createActualPR creates an actual pull request via GitHub API.
func (c *Creator) createActualPR() (PRResponse, error) {
	fmt.Printf("  - Creating pull request from %s to %s... ", c.config.PRBranch, c.config.PRBase)

	prData, err := c.preparePRData()
	if err != nil {
		return PRResponse{}, err
	}

	resp, err := c.client.Post("/pulls", prData)
	if err != nil {
		return PRResponse{}, err
	}

	return parsePRResponse(resp), nil
}

// preparePRData creates the data structure needed for the PR creation API call.
func (c *Creator) preparePRData() (map[string]interface{}, error) {
	runID := os.Getenv("GITHUB_RUN_ID")

	commitSHA, err := shared.CurrentCommitSHA()
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
func (c *Creator) HandlePRResponse(response PRResponse, sourceBranch string) error {
	if response.DryRun {
		return c.handleDryRunResponse(response)
	}

	if response.Message != "" {
		return c.handleErrorResponse(response, response.Message)
	}

	return c.handleSuccessfulPR(response, sourceBranch)
}

// handleDryRunResponse handles the dry run response without making actual changes.
func (c *Creator) handleDryRunResponse(response PRResponse) error {
	fmt.Printf("\n[DRY RUN] Pull request would be created at: %s\n", response.HTMLURL)
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
func (c *Creator) handleErrorResponse(response PRResponse, errMsg string) error {
	fmt.Printf("GitHub API Error: %s\n", errMsg)

	if response.Errors != nil {
		fmt.Println("Error details:")
		for _, err := range response.Errors {
			if errMap, ok := err.(map[string]any); ok {
				fmt.Printf("  - %v\n", errMap)

				if message, ok := errMap["message"].(string); ok &&
					strings.Contains(message, "A pull request already exists") {
					return c.handleExistingPR()
				}
			}
		}
	}

	return errors.NewAPIError("create PR", errMsg)
}

// handleSuccessfulPR processes a successful PR creation response.
func (c *Creator) handleSuccessfulPR(response PRResponse, sourceBranch string) error {
	if response.HTMLURL == "" {
		fmt.Println("[WARN] Failed to create PR")
		fmt.Printf("Response: %v\n", response)
		return errors.NewAPIError("create PR", "failed to get PR URL from response")
	}

	fmt.Println("Done")
	fmt.Printf("Pull request created: %s\n", response.HTMLURL)

	if response.HasNumber {
		if err := c.processExistingPR(response.Number); err != nil {
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

// applyToPR performs a single PR-mutation API call with the standard dry-run
// guard and "  - <progress>... " → "Done" / "FAILED" progress feedback.
// dryRunMsg is the full line printed (and short-circuit returned) in dry-run
// mode; progress is the in-progress label; apiErrOp names the operation for the
// wrapped APIError; call is the client method (Post / Patch) to invoke.
func (c *Creator) applyToPR(
	dryRunMsg, progress, apiErrOp, endpoint string,
	call func(string, interface{}) (map[string]interface{}, error),
	payload interface{},
) error {
	if c.config.PRDryRun {
		fmt.Println(dryRunMsg)
		return nil
	}

	fmt.Printf("  - %s... ", progress)
	if _, err := call(endpoint, payload); err != nil {
		fmt.Println("FAILED")
		return errors.NewAPIErrorFrom(apiErrOp, err)
	}

	fmt.Println("Done")
	return nil
}

// addLabelsToIssue adds labels to an issue/PR.
func (c *Creator) addLabelsToIssue(prNumber int) error {
	return c.applyToPR(
		fmt.Sprintf("  - [DRY RUN] Would add labels %v to PR #%d... Skipped", c.config.PRLabels, prNumber),
		fmt.Sprintf("Adding labels to PR #%d", prNumber),
		"add labels",
		fmt.Sprintf("/issues/%d/labels", prNumber),
		c.client.Post,
		map[string]interface{}{"labels": c.config.PRLabels},
	)
}

// requestReviewers requests reviewers for a pull request.
func (c *Creator) requestReviewers(prNumber int) error {
	return c.applyToPR(
		fmt.Sprintf("  - [DRY RUN] Would request reviewers %v for PR #%d... Skipped", c.config.PRReviewers, prNumber),
		fmt.Sprintf("Requesting reviewers for PR #%d", prNumber),
		"request reviewers",
		fmt.Sprintf("/pulls/%d/requested_reviewers", prNumber),
		c.client.Post,
		map[string]interface{}{"reviewers": c.config.PRReviewers},
	)
}

// addAssignees adds assignees to a pull request.
func (c *Creator) addAssignees(prNumber int) error {
	return c.applyToPR(
		fmt.Sprintf("  - [DRY RUN] Would add assignees %v to PR #%d... Skipped", c.config.PRAssignees, prNumber),
		fmt.Sprintf("Adding assignees to PR #%d", prNumber),
		"add assignees",
		fmt.Sprintf("/issues/%d/assignees", prNumber),
		c.client.Post,
		map[string]interface{}{"assignees": c.config.PRAssignees},
	)
}

// closePullRequest closes a pull request.
func (c *Creator) closePullRequest(prNumber int) error {
	return c.applyToPR(
		fmt.Sprintf("  - [DRY RUN] Would close pull request #%d... Skipped", prNumber),
		fmt.Sprintf("Closing pull request #%d", prNumber),
		"close PR",
		fmt.Sprintf("/pulls/%d", prNumber),
		c.client.Patch,
		map[string]string{"state": "closed"},
	)
}
