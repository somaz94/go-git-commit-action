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
	"github.com/somaz94/go-git-commit-action/internal/gitcmd"
)

// CommandDef defines a command to be executed
type PRCommandDef struct {
	name string
	args []string
	desc string
}

// GitHubClient handles GitHub API interactions.
type GitHubClient struct {
	token      string
	baseURL    string
	repository string
}

// NewGitHubClient creates a new GitHubClient instance with the specified token and repository.
func NewGitHubClient(token, repository string) *GitHubClient {
	return &GitHubClient{
		token:      token,
		baseURL:    "https://api.github.com",
		repository: repository,
	}
}

// CreatePullRequest creates a pull request using the GitHub API.
// This is a placeholder for actual GitHub API implementation.
func (c *GitHubClient) CreatePullRequest(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	// API request logic would be implemented here
	return nil, nil
}

// AddLabels adds labels to a pull request.
// This is a placeholder for actual GitHub API implementation.
func (c *GitHubClient) AddLabels(ctx context.Context, prNumber int, labels []string) error {
	// Label addition logic would be implemented here
	return nil
}

// ClosePullRequest closes a pull request.
// This is a placeholder for actual GitHub API implementation.
func (c *GitHubClient) ClosePullRequest(ctx context.Context, prNumber int) error {
	// PR closing logic would be implemented here
	return nil
}

// CreatePullRequest is the main function to create a GitHub pull request.
// It handles the entire flow of preparing branches, creating the PR,
// and processing post-creation tasks like adding labels or closing the PR.
func CreatePullRequest(config *config.GitConfig) error {
	fmt.Println("\n🔄 Creating Pull Request:")

	// Step 1: Prepare the source branch
	sourceBranch, err := prepareSourceBranch(config)
	if err != nil {
		return err
	}

	// Step 2: Check for differences between branches
	if err := checkBranchDifferences(config); err != nil {
		return err
	}

	// Step 3: Create the actual pull request via GitHub API
	prResponse, err := createGitHubPR(config)
	if err != nil {
		return err
	}

	// Step 4: Process the PR response (labels, closing, etc.)
	if err := handlePRResponse(config, prResponse, sourceBranch); err != nil {
		return err
	}

	fmt.Println("\n✨ Git Commit Action Completed Successfully!\n" +
		"=========================================")

	return nil
}

// prepareSourceBranch sets up the branch that will be used as the source for the PR.
// If auto_branch is enabled, it creates a new branch with a timestamp.
// Otherwise, it uses the specified PR branch.
func prepareSourceBranch(config *config.GitConfig) (string, error) {
	if config.AutoBranch {
		return createAutoBranch(config)
	}

	// Use the specified branch when auto_branch is disabled
	return checkoutExistingBranch(config)
}

// createAutoBranch creates a new branch with a timestamp and commits changes to it.
func createAutoBranch(config *config.GitConfig) (string, error) {
	// Create a branch name with a timestamp
	sourceBranch := fmt.Sprintf("update-files-%s", time.Now().Format("20060102-150405"))
	config.PRBranch = sourceBranch

	// Create and switch to a new branch
	fmt.Printf("  • Creating new branch %s... ", sourceBranch)
	if err := exec.Command(gitcmd.CmdGit, gitcmd.CheckoutNewBranchArgs(sourceBranch)...).Run(); err != nil {
		fmt.Println("❌ Failed")
		return "", fmt.Errorf("failed to create branch: %v", err)
	}
	fmt.Println("✅ Done")

	// Commit changes to the new branch and push
	if err := commitAndPushChanges(config); err != nil {
		return "", err
	}

	return sourceBranch, nil
}

// checkoutExistingBranch checks out the specified PR branch.
func checkoutExistingBranch(config *config.GitConfig) (string, error) {
	sourceBranch := config.PRBranch
	fmt.Printf("  • Checking out branch %s... ", sourceBranch)
	if err := exec.Command(gitcmd.CmdGit, gitcmd.CheckoutArgs(sourceBranch)...).Run(); err != nil {
		fmt.Println("❌ Failed")
		return "", fmt.Errorf("failed to checkout branch: %v", err)
	}
	fmt.Println("✅ Done")

	return sourceBranch, nil
}

// commitAndPushChanges stages, commits, and pushes the specified files to the PR branch.
func commitAndPushChanges(config *config.GitConfig) error {
	// Stage the files
	if err := stagePRFiles(config.FilePattern); err != nil {
		return err
	}

	// Commit and push the changes
	return commitAndPushToBranch(config)
}

// stagePRFiles adds the specified files to the Git staging area.
// It handles multiple file patterns separated by spaces.
func stagePRFiles(filePattern string) error {
	fmt.Printf("  • Adding files... ")

	// Handle multiple patterns separated by spaces
	if strings.Contains(filePattern, " ") {
		patterns := strings.Fields(filePattern)
		for _, pattern := range patterns {
			if err := executePRGitAdd(pattern); err != nil {
				fmt.Println("❌ Failed")
				return fmt.Errorf("failed to add pattern %s: %v", pattern, err)
			}
		}
	} else {
		// Single pattern case
		if err := executePRGitAdd(filePattern); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to add files: %v", err)
		}
	}

	fmt.Println("✅ Done")
	return nil
}

// executePRGitAdd executes the git add command for a specific pattern.
func executePRGitAdd(pattern string) error {
	addCmd := exec.Command(gitcmd.CmdGit, gitcmd.AddArgs(pattern)...)
	addCmd.Stdout = os.Stdout
	addCmd.Stderr = os.Stderr
	return addCmd.Run()
}

// commitAndPushToBranch commits the staged changes and pushes them to the remote branch.
func commitAndPushToBranch(config *config.GitConfig) error {
	commitPushCommands := []PRCommandDef{
		{gitcmd.CmdGit, gitcmd.CommitArgs(config.CommitMessage), "Committing changes"},
		{gitcmd.CmdGit, gitcmd.PushUpstreamArgs(gitcmd.RefOrigin, config.PRBranch), "Pushing changes"},
	}

	return executePRCommandBatch(commitPushCommands, "")
}

// executePRCommandBatch runs a batch of commands with consistent output formatting.
func executePRCommandBatch(commands []PRCommandDef, headerMessage string) error {
	if headerMessage != "" {
		fmt.Println(headerMessage)
	}

	for _, cmd := range commands {
		fmt.Printf("  • %s... ", cmd.desc)
		command := exec.Command(cmd.name, cmd.args...)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr

		if err := command.Run(); err != nil {
			// Special handling for "nothing to commit" case
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
// It also shows the potential PR URL for manual creation if the API fails.
func checkBranchDifferences(config *config.GitConfig) error {
	fmt.Printf("\n📊 Changed files between %s and %s:\n", config.PRBase, config.PRBranch)

	// Fetch the latest from both branches
	if err := fetchBranches(config); err != nil {
		return err
	}

	// Display the changed files
	return displayChangedFiles(config)
}

// fetchBranches fetches the latest from both the base and source branches.
func fetchBranches(config *config.GitConfig) error {
	fetchCommands := []PRCommandDef{
		{gitcmd.CmdGit, gitcmd.FetchArgs(gitcmd.RefOrigin, config.PRBase), "Fetching base branch"},
		{gitcmd.CmdGit, gitcmd.FetchArgs(gitcmd.RefOrigin, config.PRBranch), "Fetching source branch"},
	}

	for _, cmd := range fetchCommands {
		if err := exec.Command(cmd.name, cmd.args...).Run(); err != nil {
			return fmt.Errorf("%s: %v", cmd.desc, err)
		}
	}

	return nil
}

// displayChangedFiles shows the changed files between branches and validates if changes exist.
func displayChangedFiles(config *config.GitConfig) error {
	// Check the changed files
	diffFiles := exec.Command(gitcmd.CmdGit, gitcmd.DiffNameStatusArgs(
		fmt.Sprintf("origin/%s", config.PRBase),
		fmt.Sprintf("origin/%s", config.PRBranch),
	)...)
	filesOutput, _ := diffFiles.Output()

	if len(filesOutput) == 0 {
		fmt.Println("No changes detected")
		if config.SkipIfEmpty {
			return nil
		}
		return fmt.Errorf("no changes to create PR")
	}

	fmt.Printf("%s\n", string(filesOutput))

	// Display the PR URL for manual creation if needed
	fmt.Printf("\n✅ Branch '%s' is ready for PR.\n", config.PRBranch)
	prURL := fmt.Sprintf("https://github.com/%s/compare/%s...%s",
		os.Getenv("GITHUB_REPOSITORY"),
		config.PRBase,
		config.PRBranch)
	fmt.Printf("✅ You can create a pull request by visiting:\n   %s\n", prURL)

	return nil
}

// createGitHubPR creates a pull request using the GitHub API.
// It handles both the actual API call and dry run mode.
func createGitHubPR(config *config.GitConfig) (map[string]interface{}, error) {
	// Handle dry run mode separately
	if config.PRDryRun {
		return createDryRunPR(config)
	}

	// Create a real PR via GitHub API
	return createActualPR(config)
}

// createDryRunPR simulates PR creation in dry run mode.
func createDryRunPR(config *config.GitConfig) (map[string]interface{}, error) {
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

// createActualPR creates an actual pull request via GitHub API.
func createActualPR(config *config.GitConfig) (map[string]interface{}, error) {
	fmt.Printf("  • Creating pull request from %s to %s... ", config.PRBranch, config.PRBase)

	// Prepare the PR data
	prData, err := preparePRData(config)
	if err != nil {
		return nil, err
	}

	// Call the GitHub API
	return callGitHubAPI(config, prData)
}

// preparePRData creates the data structure needed for the PR creation API call.
func preparePRData(config *config.GitConfig) (map[string]interface{}, error) {
	// Get the GitHub Run ID for reference
	runID := os.Getenv("GITHUB_RUN_ID")

	// Get the current commit SHA
	commitSHA, err := getCurrentCommitSHA()
	if err != nil {
		return nil, err
	}

	// Generate title and body if not provided
	title, body := generatePRTitleAndBody(config, runID, commitSHA)

	// Create the PR data
	prData := map[string]interface{}{
		"title": title,
		"head":  config.PRBranch,
		"base":  config.PRBase,
		"body":  body,
	}

	return prData, nil
}

// getCurrentCommitSHA retrieves the current commit SHA.
func getCurrentCommitSHA() (string, error) {
	commitCmd := exec.Command(gitcmd.CmdGit, gitcmd.RevParseArgs("HEAD")...)
	commitSHA, err := commitCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit SHA: %v", err)
	}
	return strings.TrimSpace(string(commitSHA)), nil
}

// generatePRTitleAndBody creates default PR title and body if not specified.
func generatePRTitleAndBody(config *config.GitConfig, runID string, commitSHA string) (string, string) {
	title := config.PRTitle
	if title == "" {
		title = fmt.Sprintf("Auto PR: %s to %s (Run ID: %s)", config.PRBranch, config.PRBase, runID)
	}

	body := config.PRBody
	if body == "" {
		body = fmt.Sprintf("Created by Go Git Commit Action\nSource: %s\nTarget: %s\nCommit: %s\nGitHub Run ID: %s",
			config.PRBranch, config.PRBase, commitSHA, runID)
	}

	return title, body
}

// callGitHubAPI makes the actual GitHub API call to create a PR.
func callGitHubAPI(config *config.GitConfig, prData map[string]interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(prData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal PR data: %v", err)
	}

	// Create a pull request using the GitHub API via curl
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

	// Parse the response JSON
	var response map[string]interface{}
	if err := json.Unmarshal(output, &response); err != nil {
		fmt.Printf("Raw response: %s\n", string(output))
		return nil, fmt.Errorf("failed to parse PR response: %v", err)
	}

	return response, nil
}

// handlePRResponse processes the PR creation response and performs follow-up actions.
// This includes handling dry run responses, adding labels, closing PRs, and deleting branches.
func handlePRResponse(config *config.GitConfig, response map[string]interface{}, sourceBranch string) error {
	// Handle dry run responses separately
	if dryRun, ok := response["dry_run"].(bool); ok && dryRun {
		return handleDryRunResponse(config, response)
	}

	// Check for error messages in the response
	if errMsg, ok := response["message"].(string); ok {
		return handleErrorResponse(config, response, errMsg)
	}

	// Handle successful PR creation
	return handleSuccessfulPR(config, response, sourceBranch)
}

// handleDryRunResponse handles the dry run response without making actual changes.
func handleDryRunResponse(config *config.GitConfig, response map[string]interface{}) error {
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

	if config.DeleteSourceBranch {
		if config.AutoBranch {
			fmt.Printf("  • Source branch would be deleted: Yes (auto-generated branch)\n")
		} else {
			fmt.Printf("  • Warning: delete_source_branch is set but requires auto_branch:true to take effect\n")
		}
	}

	return nil
}

// handleErrorResponse processes error responses from the GitHub API.
func handleErrorResponse(config *config.GitConfig, response map[string]interface{}, errMsg string) error {
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
					return handleExistingPR(config)
				}
			}
		}
	}

	return fmt.Errorf("GitHub API error: %s", errMsg)
}

// handleSuccessfulPR processes a successful PR creation response.
func handleSuccessfulPR(config *config.GitConfig, response map[string]interface{}, sourceBranch string) error {
	// Extract the PR URL and display it
	if htmlURL, ok := response["html_url"].(string); ok {
		fmt.Println("✅ Done")
		fmt.Printf("Pull request created: %s\n", htmlURL)

		// Process the PR number for additional operations
		if number, ok := response["number"].(float64); ok {
			prNumber := int(number)

			// Add labels if specified
			if len(config.PRLabels) > 0 {
				if err := addLabelsToIssue(config, prNumber); err != nil {
					return err
				}
			}

			// Close the PR if specified
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

	// Delete the source branch if auto-branch and delete-source-branch are enabled
	if config.DeleteSourceBranch && config.AutoBranch {
		if err := deleteSourceBranch(config, sourceBranch); err != nil {
			return err
		}
	}

	return nil
}

// handleExistingPR processes the case when a PR already exists.
func handleExistingPR(config *config.GitConfig) error {
	fmt.Println("⚠️  Pull request already exists")

	// Find the existing PR
	prs, err := findExistingPRs(config)
	if err != nil {
		return err
	}

	// If we found existing PRs, process them
	if len(prs) > 0 {
		if number, ok := prs[0]["number"].(float64); ok {
			prNumber := int(number)
			fmt.Printf("Found existing PR #%d\n", prNumber)

			// Process the existing PR (labels, closing)
			return processExistingPR(config, prNumber)
		}
	}

	return nil
}

// findExistingPRs searches for existing PRs with the same head and base.
func findExistingPRs(config *config.GitConfig) ([]map[string]interface{}, error) {
	searchCmd := exec.Command("curl", "-s",
		"-H", fmt.Sprintf("Authorization: Bearer %s", config.GitHubToken),
		"-H", "Accept: application/vnd.github+json",
		"-H", "X-GitHub-Api-Version: 2022-11-28",
		fmt.Sprintf("https://api.github.com/repos/%s/pulls?head=%s&base=%s",
			os.Getenv("GITHUB_REPOSITORY"),
			config.PRBranch,
			config.PRBase))

	searchOutput, err := searchCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to search for existing PRs: %v", err)
	}

	var prs []map[string]interface{}
	if err := json.Unmarshal(searchOutput, &prs); err != nil {
		return nil, fmt.Errorf("failed to parse existing PR search response: %v", err)
	}

	return prs, nil
}

// processExistingPR applies operations like adding labels or closing to an existing PR.
func processExistingPR(config *config.GitConfig, prNumber int) error {
	// Add labels if specified
	if len(config.PRLabels) > 0 {
		if err := addLabelsToIssue(config, prNumber); err != nil {
			return err
		}
	}

	// Close the PR if specified
	if config.PRClosed {
		if err := closePullRequest(config, prNumber); err != nil {
			return err
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

	// Prepare the labels data
	labelsData := map[string]interface{}{
		"labels": config.PRLabels,
	}
	jsonLabelsData, err := json.Marshal(labelsData)
	if err != nil {
		return fmt.Errorf("failed to marshal labels data: %v", err)
	}

	// Execute the API call
	labelsCurlCmd := exec.Command("curl", "-s", "-X", "POST",
		"-H", fmt.Sprintf("Authorization: Bearer %s", config.GitHubToken),
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
		return fmt.Errorf("failed to add labels: %v", err)
	}

	fmt.Println("✅ Done")
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

	// Prepare the close data
	closeData := map[string]string{
		"state": "closed",
	}
	jsonCloseData, err := json.Marshal(closeData)
	if err != nil {
		return fmt.Errorf("failed to marshal close data: %v", err)
	}

	// Execute the API call
	closeCurlCmd := exec.Command("curl", "-s", "-X", "PATCH",
		"-H", fmt.Sprintf("Authorization: Bearer %s", config.GitHubToken),
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
		return fmt.Errorf("failed to close PR: %v", err)
	}

	fmt.Println("✅ Done")
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
	deleteCommand := exec.Command(gitcmd.CmdGit, gitcmd.SubCmdPush, gitcmd.RefOrigin, "--delete", sourceBranch)
	if err := deleteCommand.Run(); err != nil {
		fmt.Println("❌ Failed")
		return fmt.Errorf("failed to delete source branch: %v", err)
	}

	fmt.Println("✅ Done")
	return nil
}
