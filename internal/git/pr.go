package git

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/somaz94/go-git-commit-action/internal/config"
)

// The GitHubClient is a structure for GitHub API requests.
type GitHubClient struct {
	token      string
	baseURL    string
	repository string
}

// NewGitHubClient creates a new GitHubClient instance.
func NewGitHubClient(token, repository string) *GitHubClient {
	return &GitHubClient{
		token:      token,
		baseURL:    "https://api.github.com",
		repository: repository,
	}
}

// CreatePullRequest creates a pull request using the GitHub API.
func (c *GitHubClient) CreatePullRequest(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	// API request logic
	return nil, nil
}

// AddLabels adds labels to a pull request.
func (c *GitHubClient) AddLabels(ctx context.Context, prNumber int, labels []string) error {
	// Label addition logic
	return nil
}

// ClosePullRequest closes a pull request.
func (c *GitHubClient) ClosePullRequest(ctx context.Context, prNumber int) error {
	// PR closing logic
	return nil
}

// CreatePullRequest is the main function to create a pull request.
func CreatePullRequest(config *config.GitConfig) error {
	fmt.Println("\n🔄 Creating Pull Request:")

	// Prepare the source branch
	sourceBranch, err := prepareSourceBranch(config)
	if err != nil {
		return err
	}

	// Check for changes between branches
	if err := checkBranchDifferences(config); err != nil {
		return err
	}

	// Create a pull request
	prResponse, err := createGitHubPR(config)
	if err != nil {
		return err
	}

	// Process the PR response
	if err := handlePRResponse(config, prResponse, sourceBranch); err != nil {
		return err
	}

	fmt.Println("\n✨ Git Commit Action Completed Successfully!\n" +
		"=========================================")

	return nil
}

// prepareSourceBranch prepares the source branch.
func prepareSourceBranch(config *config.GitConfig) (string, error) {
	var sourceBranch string

	if config.AutoBranch {
		// Create a branch name with a timestamp
		sourceBranch = fmt.Sprintf("update-files-%s", time.Now().Format("20060102-150405"))
		config.PRBranch = sourceBranch

		// Create and switch to a new branch
		fmt.Printf("  • Creating new branch %s... ", sourceBranch)
		if err := exec.Command("git", "checkout", "-b", sourceBranch).Run(); err != nil {
			fmt.Println("❌ Failed")
			return "", fmt.Errorf("failed to create branch: %v", err)
		}
		fmt.Println("✅ Done")

		// Commit changes to the new branch and push
		if err := commitAndPushChanges(config); err != nil {
			return "", err
		}
	} else {
		// When auto_branch=false, checkout the pr_branch
		sourceBranch = config.PRBranch
		fmt.Printf("  • Checking out branch %s... ", sourceBranch)
		if err := exec.Command("git", "checkout", sourceBranch).Run(); err != nil {
			fmt.Println("❌ Failed")
			return "", fmt.Errorf("failed to checkout branch: %v", err)
		}
		fmt.Println("✅ Done")
	}

	return sourceBranch, nil
}

// commitAndPushChanges commits and pushes changes.
func commitAndPushChanges(config *config.GitConfig) error {
	commitCommands := []struct {
		name string
		args []string
		desc string
	}{
		{"git", []string{"add", config.FilePattern}, "Adding files"},
		{"git", []string{"commit", "-m", config.CommitMessage}, "Committing changes"},
		{"git", []string{"push", "-u", "origin", config.PRBranch}, "Pushing changes"},
	}

	for _, cmd := range commitCommands {
		fmt.Printf("  • %s... ", cmd.desc)
		command := exec.Command(cmd.name, cmd.args...)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		if err := command.Run(); err != nil {
			if cmd.args[0] == "commit" && err.Error() == "exit status 1" {
				fmt.Println("⚠️  Nothing to commit, skipping...")
				continue
			}
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to execute %s: %v", cmd.name, err)
		}
		fmt.Println("✅ Done")
	}

	return nil
}

// checkBranchDifferences checks the differences between the PR base branch and the source branch.
func checkBranchDifferences(config *config.GitConfig) error {
	fmt.Printf("\n📊 Changed files between %s and %s:\n", config.PRBase, config.PRBranch)

	// Get the two branches
	if err := fetchBranches(config); err != nil {
		return err
	}

	// Check the changed files
	diffFiles := exec.Command("git", "diff", fmt.Sprintf("origin/%s..origin/%s", config.PRBase, config.PRBranch), "--name-status")
	filesOutput, _ := diffFiles.Output()

	if len(filesOutput) == 0 {
		fmt.Println("No changes detected")
		if config.SkipIfEmpty {
			return nil
		}
		return fmt.Errorf("no changes to create PR")
	}
	fmt.Printf("%s\n", string(filesOutput))

	// Create and print the PR URL
	fmt.Printf("\n✅ Branch '%s' is ready for PR.\n", config.PRBranch)
	prURL := fmt.Sprintf("https://github.com/%s/compare/%s...%s",
		os.Getenv("GITHUB_REPOSITORY"),
		config.PRBase,
		config.PRBranch)
	fmt.Printf("✅ You can create a pull request by visiting:\n   %s\n", prURL)

	return nil
}

// fetchBranches fetches the base and source branches.
func fetchBranches(config *config.GitConfig) error {
	fetchBaseCmd := exec.Command("git", "fetch", "origin", config.PRBase)
	if err := fetchBaseCmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch base branch: %v", err)
	}

	fetchBranchCmd := exec.Command("git", "fetch", "origin", config.PRBranch)
	if err := fetchBranchCmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch source branch: %v", err)
	}

	return nil
}

// createGitHubPR creates a pull request using the GitHub API.
func createGitHubPR(config *config.GitConfig) (map[string]interface{}, error) {
	// Check for dry run mode
	if config.PRDryRun {
		fmt.Printf("  • [DRY RUN] Would create pull request from %s to %s... ", config.PRBranch, config.PRBase)
		fmt.Println("✅ Skipped (Dry Run mode)")

		// Create a mock response with the PR URL
		mockResponse := map[string]interface{}{
			"html_url": fmt.Sprintf("https://github.com/%s/compare/%s...%s?dry_run=1",
				os.Getenv("GITHUB_REPOSITORY"),
				config.PRBase,
				config.PRBranch),
			"number":  float64(0),
			"dry_run": true,
		}

		return mockResponse, nil
	}

	fmt.Printf("  • Creating pull request from %s to %s... ", config.PRBranch, config.PRBase)

	// Prepare the PR data
	prData, err := preparePRData(config)
	if err != nil {
		return nil, err
	}

	// Call the GitHub API
	return callGitHubAPI(config, prData)
}

// preparePRData prepares the data needed to create a pull request.
func preparePRData(config *config.GitConfig) (map[string]interface{}, error) {
	// Get the GitHub Run ID
	runID := os.Getenv("GITHUB_RUN_ID")

	// Get the current commit SHA
	commitCmd := exec.Command("git", "rev-parse", "HEAD")
	commitSHA, err := commitCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get commit SHA: %v", err)
	}
	commitID := strings.TrimSpace(string(commitSHA))

	// Set the PR title
	title := config.PRTitle
	if title == "" {
		title = fmt.Sprintf("Auto PR: %s to %s (Run ID: %s)", config.PRBranch, config.PRBase, runID)
	}

	// Set the PR body
	body := config.PRBody
	if body == "" {
		body = fmt.Sprintf("Created by Go Git Commit Action\nSource: %s\nTarget: %s\nCommit: %s\nGitHub Run ID: %s",
			config.PRBranch, config.PRBase, commitID, runID)
	}

	// PR request data
	prData := map[string]interface{}{
		"title": title,
		"head":  config.PRBranch,
		"base":  config.PRBase,
		"body":  body,
	}

	return prData, nil
}

// callGitHubAPI calls the GitHub API to create a pull request.
func callGitHubAPI(config *config.GitConfig, prData map[string]interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(prData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal PR data: %v", err)
	}

	// Create a pull request using the GitHub API
	curlCmd := exec.Command("curl", "-s", "-X", "POST",
		"-H", fmt.Sprintf("Authorization: Bearer %s", config.GitHubToken),
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

	// Parse the response
	var response map[string]interface{}
	if err := json.Unmarshal(output, &response); err != nil {
		fmt.Printf("Raw response: %s\n", string(output))
		return nil, fmt.Errorf("failed to parse PR response: %v", err)
	}

	return response, nil
}

// handlePRResponse handles the PR creation response.
func handlePRResponse(config *config.GitConfig, response map[string]interface{}, sourceBranch string) error {
	// Check if this is a dry run response
	if dryRun, ok := response["dry_run"].(bool); ok && dryRun {
		fmt.Printf("\n🔍 [DRY RUN] Pull request would be created at: %s\n", response["html_url"])
		fmt.Printf("👉 No actual PR was created (dry run mode)\n")

		// Print mock PR details
		fmt.Printf("\n📋 PR details that would be submitted:\n")
		fmt.Printf("  • Title: %s\n", config.PRTitle)
		fmt.Printf("  • Source branch: %s\n", config.PRBranch)
		fmt.Printf("  • Target branch: %s\n", config.PRBase)
		if len(config.PRLabels) > 0 {
			fmt.Printf("  • Labels: %s\n", strings.Join(config.PRLabels, ", "))
		}
		if config.PRClosed {
			fmt.Printf("  • Would be closed immediately: Yes\n")
		}
		if config.DeleteSourceBranch && config.AutoBranch {
			fmt.Printf("  • Source branch would be deleted: Yes\n")
		}

		return nil
	}

	// Check for error message
	if errMsg, ok := response["message"].(string); ok {
		fmt.Printf("GitHub API Error: %s\n", errMsg)
		if errors, ok := response["errors"].([]interface{}); ok {
			fmt.Println("Error details:")
			for _, err := range errors {
				if errMap, ok := err.(map[string]interface{}); ok {
					fmt.Printf("  • %v\n", errMap)
					// Handle the case where the PR already exists
					if errMap["message"].(string) == "A pull request already exists for somaz94:test." {
						return handleExistingPR(config)
					}
				}
			}
		}
		return fmt.Errorf("GitHub API error: %s", errMsg)
	}

	// Check for PR URL
	if htmlURL, ok := response["html_url"].(string); ok {
		fmt.Println("✅ Done")
		fmt.Printf("Pull request created: %s\n", htmlURL)

		// Handle the PR number
		if number, ok := response["number"].(float64); ok {
			prNumber := int(number)

			// Add labels
			if len(config.PRLabels) > 0 {
				if err := addLabelsToIssue(config, prNumber); err != nil {
					return err
				}
			}

			// Close the PR
			if config.PRClosed {
				if err := closePullRequest(config, prNumber); err != nil {
					return err
				}
			}
		}
	} else {
		fmt.Println("⚠️  Failed to create PR")
		fmt.Printf("Response: %v\n", response)
		return fmt.Errorf("failed to get PR URL from response")
	}

	// Delete the source branch
	if config.DeleteSourceBranch && config.AutoBranch {
		if err := deleteSourceBranch(config, sourceBranch); err != nil {
			return err
		}
	}

	return nil
}

// handleExistingPR handles an existing pull request.
func handleExistingPR(config *config.GitConfig) error {
	fmt.Println("⚠️  Pull request already exists")

	// Find the existing PR
	searchCmd := exec.Command("curl", "-s",
		"-H", fmt.Sprintf("Authorization: Bearer %s", config.GitHubToken),
		"-H", "Accept: application/vnd.github+json",
		"-H", "X-GitHub-Api-Version: 2022-11-28",
		fmt.Sprintf("https://api.github.com/repos/%s/pulls?head=%s&base=%s",
			os.Getenv("GITHUB_REPOSITORY"),
			config.PRBranch,
			config.PRBase))

	searchOutput, _ := searchCmd.CombinedOutput()
	var prs []map[string]interface{}
	if err := json.Unmarshal(searchOutput, &prs); err == nil && len(prs) > 0 {
		if number, ok := prs[0]["number"].(float64); ok {
			prNumber := int(number)
			fmt.Printf("Found existing PR #%d\n", prNumber)

			// Add labels
			if len(config.PRLabels) > 0 {
				if err := addLabelsToIssue(config, prNumber); err != nil {
					return err
				}
			}

			// Close the PR
			if config.PRClosed {
				if err := closePullRequest(config, prNumber); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// addLabelsToIssue adds labels to an issue/PR.
func addLabelsToIssue(config *config.GitConfig, prNumber int) error {
	// Skip if in dry run mode
	if config.PRDryRun {
		fmt.Printf("  • [DRY RUN] Would add labels %v to PR #%d... ✅ Skipped\n", config.PRLabels, prNumber)
		return nil
	}

	fmt.Printf("  • Adding labels to PR #%d... ", prNumber)
	labelsData := map[string]interface{}{
		"labels": config.PRLabels,
	}
	jsonLabelsData, _ := json.Marshal(labelsData)

	labelsCurlCmd := exec.Command("curl", "-s", "-X", "POST",
		"-H", fmt.Sprintf("Authorization: Bearer %s", config.GitHubToken),
		"-H", "Accept: application/vnd.github+json",
		"-H", "Content-Type: application/json",
		fmt.Sprintf("https://api.github.com/repos/%s/issues/%d/labels",
			os.Getenv("GITHUB_REPOSITORY"), prNumber),
		"-d", string(jsonLabelsData))

	if labelsOutput, err := labelsCurlCmd.CombinedOutput(); err != nil {
		fmt.Println("❌ Failed")
		fmt.Printf("Error: %v\n", err)
		fmt.Printf("Response: %s\n", string(labelsOutput))
		return fmt.Errorf("failed to add labels: %v", err)
	} else {
		fmt.Println("✅ Done")
	}

	return nil
}

// closePullRequest closes a pull request.
func closePullRequest(config *config.GitConfig, prNumber int) error {
	// Skip if in dry run mode
	if config.PRDryRun {
		fmt.Printf("  • [DRY RUN] Would close pull request #%d... ✅ Skipped\n", prNumber)
		return nil
	}

	fmt.Printf("  • Closing pull request #%d... ", prNumber)
	closeData := map[string]string{
		"state": "closed",
	}
	jsonCloseData, _ := json.Marshal(closeData)

	closeCurlCmd := exec.Command("curl", "-s", "-X", "PATCH",
		"-H", fmt.Sprintf("Authorization: Bearer %s", config.GitHubToken),
		"-H", "Accept: application/vnd.github+json",
		"-H", "Content-Type: application/json",
		fmt.Sprintf("https://api.github.com/repos/%s/pulls/%d",
			os.Getenv("GITHUB_REPOSITORY"), prNumber),
		"-d", string(jsonCloseData))

	if closeOutput, err := closeCurlCmd.CombinedOutput(); err != nil {
		fmt.Println("❌ Failed")
		fmt.Printf("Error: %v\n", err)
		fmt.Printf("Response: %s\n", string(closeOutput))
		return fmt.Errorf("failed to close PR: %v", err)
	} else {
		fmt.Println("✅ Done")
	}

	return nil
}

// deleteSourceBranch deletes the source branch.
func deleteSourceBranch(config *config.GitConfig, sourceBranch string) error {
	// Skip if in dry run mode
	if config.PRDryRun {
		fmt.Printf("\n  • [DRY RUN] Would delete source branch %s... ✅ Skipped\n", sourceBranch)
		return nil
	}

	fmt.Printf("\n  • Deleting source branch %s... ", sourceBranch)
	deleteCommand := exec.Command("git", "push", "origin", "--delete", sourceBranch)
	if err := deleteCommand.Run(); err != nil {
		fmt.Println("❌ Failed")
		return fmt.Errorf("failed to delete source branch: %v", err)
	}
	fmt.Println("✅ Done")

	return nil
}
