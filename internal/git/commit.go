package git

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/somaz94/go-git-commit-action/internal/config"
	"github.com/somaz94/go-git-commit-action/internal/gitcmd"
)

// FileBackup is a struct for file backups.
type FileBackup struct {
	path    string
	content []byte
}

// withRetry provides retry logic for operations that might fail transiently.
// It executes the given operation repeatedly until it succeeds or the maximum
// number of retries is reached. The delay between retries increases linearly.
func withRetry(ctx context.Context, maxRetries int, operation func() error) error {
	var lastErr error
	for i := range make([]struct{}, maxRetries) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := operation(); err != nil {
				lastErr = err
				time.Sleep(time.Second * time.Duration(i+1))
				continue
			}
			return nil
		}
	}
	return fmt.Errorf("operation failed after %d retries: %v", maxRetries, lastErr)
}

// RunGitCommit executes the Git commit operation with the provided configuration.
// It wraps the entire process in a retry mechanism to handle transient failures.
func RunGitCommit(config *config.GitConfig) error {
	// Create a context with timeout for the entire operation
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Timeout)*time.Second)
	defer cancel()

	// Wrap the entire commit process in retry logic
	return withRetry(ctx, config.RetryCount, func() error {
		return executeGitCommitWorkflow(config)
	})
}

// executeGitCommitWorkflow runs all steps of the Git commit process
func executeGitCommitWorkflow(config *config.GitConfig) error {
	// Validate the configuration
	if err := validateConfig(config); err != nil {
		return err
	}

	// Print debug information
	printDebugInfo()

	// Change the working directory
	if err := changeWorkingDirectory(config); err != nil {
		return err
	}

	// Setup Git configuration
	if err := setupGitConfig(config); err != nil {
		return err
	}

	// Handle the branch
	if err := handleBranch(config); err != nil {
		return err
	}

	// Check for changes
	isEmpty, err := checkIfEmpty(config)
	if err != nil {
		return err
	}

	if isEmpty {
		fmt.Println("\n⚠️  No changes detected and skip_if_empty is true. Skipping commit process.")
		return nil
	}

	// Create a PR or commit directly based on configuration
	if config.CreatePR {
		return handlePullRequestFlow(config)
	}

	return commitChanges(config)
}

// validateConfig ensures all required configuration parameters are set.
// It checks that the necessary fields for PR creation are specified when the
// create_pr option is enabled.
func validateConfig(config *config.GitConfig) error {
	if !config.CreatePR {
		return nil
	}

	// Validate PR-specific configuration
	if !config.AutoBranch && config.PRBranch == "" {
		return fmt.Errorf("pr_branch must be specified when auto_branch is false and create_pr is true")
	}

	if config.PRBase == "" {
		return fmt.Errorf("pr_base must be specified when create_pr is true")
	}

	if config.GitHubToken == "" {
		return fmt.Errorf("github_token must be specified when create_pr is true")
	}

	return nil
}

// printDebugInfo outputs debug information about the current environment.
// This includes the working directory and the contents of the directory.
func printDebugInfo() {
	currentDir, _ := os.Getwd()
	fmt.Println("\n🚀 Starting Git Commit Action\n" +
		"================================")

	fmt.Println("\n📋 Configuration:")
	fmt.Printf("  • Working Directory: %s\n", currentDir)

	fmt.Println("\n📁 Directory Contents:")
	files, _ := os.ReadDir(".")
	for _, file := range files {
		fmt.Printf("  • %s\n", file.Name())
	}
}

// changeWorkingDirectory changes to the specified repository path if it's not
// the current directory. It reports the new directory after changing.
func changeWorkingDirectory(config *config.GitConfig) error {
	if config.RepoPath != "." {
		if err := os.Chdir(config.RepoPath); err != nil {
			return fmt.Errorf("❌ Failed to change directory: %v", err)
		}
		newDir, _ := os.Getwd()
		fmt.Printf("\n📂 Changed to directory: %s\n", newDir)
	}
	return nil
}

// setupGitConfig configures Git with user information and safety settings.
// It runs a series of git config commands to ensure the proper environment.
func setupGitConfig(config *config.GitConfig) error {
	baseCommands := []Command{
		{gitcmd.CmdGit, gitcmd.ConfigSafeDirArgs(gitcmd.PathApp), "Setting safe directory (/app)"},
		{gitcmd.CmdGit, gitcmd.ConfigSafeDirArgs(gitcmd.PathGitHubWorkspace), "Setting safe directory (/github/workspace)"},
		{gitcmd.CmdGit, gitcmd.ConfigUserEmailArgs(config.UserEmail), "Configuring user email"},
		{gitcmd.CmdGit, gitcmd.ConfigUserNameArgs(config.UserName), "Configuring user name"},
	}

	// Setup git credentials for checkout@v6 compatibility
	if err := setupGitCredentials(config); err != nil {
		return err
	}

	baseCommands = append(baseCommands, Command{gitcmd.CmdGit, gitcmd.ConfigListArgs(), "Checking git configuration"})

	return ExecuteCommandBatch(baseCommands, "\n⚙️  Executing Git Commands:")
}

// setupGitCredentials configures git credential helper for checkout@v6 compatibility.
// Since checkout@v6 stores credentials in $RUNNER_TEMP which is not accessible in Docker containers,
// we need to configure the credential helper directly using GITHUB_TOKEN.
func setupGitCredentials(config *config.GitConfig) error {
	// Get GitHub token from environment or config
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" && config.GitHubToken != "" {
		token = config.GitHubToken
	}

	if token == "" {
		// No token available, skip credential setup
		// This is acceptable for public repos or when credentials are already configured
		return nil
	}

	// Get the repository URL from git remote
	cmd := exec.Command(gitcmd.CmdGit, "config", "--get", "remote.origin.url")
	output, err := cmd.Output()
	if err != nil {
		// If we can't get remote URL, continue without credential setup
		return nil
	}

	remoteURL := strings.TrimSpace(string(output))

	// Extract the repository info from the URL
	// Format: https://github.com/owner/repo.git
	if !strings.Contains(remoteURL, "github.com") {
		return nil
	}

	// Configure git credential helper using extraheader
	// This is compatible with checkout@v6's approach
	auth := fmt.Sprintf("x-access-token:%s", token)
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	authHeader := fmt.Sprintf("AUTHORIZATION: basic %s", encodedAuth)

	configCmd := exec.Command(gitcmd.CmdGit, "config", "--global", "http.https://github.com/.extraheader", authHeader)
	configCmd.Stdout = os.Stdout
	configCmd.Stderr = os.Stderr

	if err := configCmd.Run(); err != nil {
		return fmt.Errorf("failed to configure git credentials: %v", err)
	}

	fmt.Println("  • Configuring git credentials... ✅ Done")
	return nil
}

// handleBranch manages branch-related operations, checking for local and remote
// branch existence and taking appropriate action.
func handleBranch(config *config.GitConfig) error {
	// Check if local branch exists
	localBranchExists := exec.Command(gitcmd.CmdGit, gitcmd.RevParseArgs(config.Branch)...).Run() == nil

	// Check if remote branch exists
	remoteBranchExists := exec.Command(gitcmd.CmdGit, gitcmd.LsRemoteHeadsArgs(gitcmd.RefOrigin, config.Branch)...).Run() == nil

	// Determine the appropriate action based on branch existence
	if !localBranchExists && !remoteBranchExists {
		// Neither local nor remote branch exists, create a new one
		return createNewBranch(config)
	} else if !localBranchExists && remoteBranchExists {
		// Only remote branch exists, check it out
		return checkoutRemoteBranch(config)
	}

	// Local branch already exists and is checked out, nothing to do
	return nil
}

// createNewBranch creates a new branch and pushes it to the remote repository.
func createNewBranch(config *config.GitConfig) error {
	fmt.Printf("\n⚠️  Branch '%s' not found, creating it...\n", config.Branch)
	createCommands := []Command{
		{gitcmd.CmdGit, gitcmd.CheckoutNewBranchArgs(config.Branch), "Creating new branch"},
		{gitcmd.CmdGit, gitcmd.PushUpstreamArgs(gitcmd.RefOrigin, config.Branch), "Pushing new branch"},
	}

	return ExecuteCommandBatch(createCommands, "")
}

// checkoutRemoteBranch checks out an existing remote branch while handling
// local changes properly through backup, stash, and restore.
func checkoutRemoteBranch(config *config.GitConfig) error {
	fmt.Printf("\n⚠️  Checking out existing remote branch '%s'...\n", config.Branch)

	// Get the current working directory state
	statusOutput, err := getGitStatus()
	if err != nil {
		return err
	}

	// Backup any modified files
	backups, err := backupChanges(config, statusOutput)
	if err != nil {
		return err
	}

	// Stash any changes to avoid conflicts during checkout
	if err := stashChanges(); err != nil {
		return err
	}

	// Fetch and checkout the remote branch
	if err := fetchAndCheckout(config); err != nil {
		return err
	}

	// Restore the backed up changes
	return restoreChanges(backups)
}

// getGitStatus returns the current Git status in porcelain format.
func getGitStatus() (string, error) {
	statusCmd := exec.Command(gitcmd.CmdGit, gitcmd.StatusPorcelainArgs()...)
	output, err := statusCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get modified files: %v", err)
	}
	return string(output), nil
}

// backupChanges creates backups of modified files that need to be preserved
// during branch switching.
func backupChanges(config *config.GitConfig, statusOutput string) ([]FileBackup, error) {
	fmt.Printf("  • Backing up changes... ")

	var backups []FileBackup

	for _, line := range strings.Split(statusOutput, "\n") {
		if line == "" {
			continue
		}

		// Separate the status code and file path
		status := line[:2]
		fullPath := strings.TrimSpace(line[3:])

		// Calculate the relative path based on config.RepoPath
		relPath := fullPath
		if config.RepoPath != "." {
			relPath = strings.TrimPrefix(fullPath, config.RepoPath+"/")
		}

		fmt.Printf("\n    - Found modified file: %s (status: %s)", relPath, status)

		// Skip deleted files since they don't need backup
		if status == " D" || status == "D " {
			continue
		}

		// Read and store file contents
		content, err := os.ReadFile(relPath)
		if err != nil {
			fmt.Println("❌ Failed")
			return nil, fmt.Errorf("failed to read file %s: %v", relPath, err)
		}

		backups = append(backups, FileBackup{path: relPath, content: content})
	}

	fmt.Println("✅ Done")
	return backups, nil
}

// stashChanges safely stashes any local changes to avoid conflicts.
func stashChanges() error {
	fmt.Printf("  • Stashing changes... ")
	stashCmd := exec.Command(gitcmd.CmdGit, gitcmd.StashPushArgs()...)
	stashCmd.Stdout = os.Stdout
	stashCmd.Stderr = os.Stderr

	if err := stashCmd.Run(); err != nil {
		fmt.Println("❌ Failed")
		return fmt.Errorf("failed to stash changes: %v", err)
	}

	fmt.Println("✅ Done")
	return nil
}

// fetchAndCheckout fetches the remote branch and checks it out locally.
func fetchAndCheckout(config *config.GitConfig) error {
	checkoutCommands := []Command{
		{gitcmd.CmdGit, gitcmd.FetchArgs(gitcmd.RefOrigin, config.Branch), "Fetching remote branch"},
		{gitcmd.CmdGit, gitcmd.CheckoutArgs(config.Branch), "Checking out branch"},
		{gitcmd.CmdGit, gitcmd.ResetHardArgs(fmt.Sprintf("origin/%s", config.Branch)), "Resetting to remote state"},
	}

	return ExecuteCommandBatch(checkoutCommands, "")
}

// restoreChanges brings back the backed up files after branch switching.
func restoreChanges(backups []FileBackup) error {
	fmt.Printf("  • Restoring changes... ")

	for _, backup := range backups {
		// Create parent directories if they don't exist
		dir := filepath.Dir(backup.path)
		if dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Println("❌ Failed")
				return fmt.Errorf("failed to create directory %s: %v", dir, err)
			}
		}

		// Write the file content
		if err := os.WriteFile(backup.path, backup.content, 0644); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to restore file %s: %v", backup.path, err)
		}
	}

	fmt.Println("✅ Done")
	return nil
}

// checkIfEmpty determines if there are any changes to commit.
// It checks both working directory changes and differences between branches.
func checkIfEmpty(config *config.GitConfig) (bool, error) {
	// Get local working directory changes
	statusCmd := exec.Command(gitcmd.CmdGit, gitcmd.StatusPorcelainArgs()...)
	statusOutput, err := statusCmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check git status: %v", err)
	}

	// Check for differences between branches
	diffCmd := exec.Command(gitcmd.CmdGit, gitcmd.DiffNameOnlyArgs(
		fmt.Sprintf("origin/%s", config.PRBase),
		config.PRBranch,
	)...)
	diffOutput, err := diffCmd.Output()
	if err != nil {
		// If error (likely new branch), consider it non-empty to proceed
		diffOutput = []byte("new-branch")
	}

	// Determine if there are no changes to commit
	hasLocalChanges := len(statusOutput) > 0
	hasBranchDifferences := len(diffOutput) > 0
	isEmpty := !hasLocalChanges && !hasBranchDifferences

	// Print debug information for better visibility
	printChangeDetectionInfo(statusOutput, diffOutput, hasLocalChanges, hasBranchDifferences)

	// Return true only if empty and config says to skip empty commits
	return isEmpty && config.SkipIfEmpty, nil
}

// printChangeDetectionInfo outputs information about detected changes.
func printChangeDetectionInfo(statusOutput, diffOutput []byte, hasLocalChanges, hasBranchDifferences bool) {
	fmt.Printf("\n📊 Change Detection:\n")
	fmt.Printf("  • Local changes: %v\n", hasLocalChanges)
	fmt.Printf("  • Branch differences: %v\n", hasBranchDifferences)

	if hasLocalChanges {
		fmt.Printf("  • Local changes details:\n%s\n", string(statusOutput))
	}

	if hasBranchDifferences {
		fmt.Printf("  • Branch differences details:\n%s\n", string(diffOutput))
	}
}

// handlePullRequestFlow manages the creation of pull requests
// based on the auto_branch configuration.
func handlePullRequestFlow(config *config.GitConfig) error {
	if config.AutoBranch {
		// Auto branch creation and PR creation in one step
		if err := CreatePullRequest(config); err != nil {
			return fmt.Errorf("failed to create pull request: %v", err)
		}
	} else {
		// First commit changes to the specified branch
		if err := commitChanges(config); err != nil {
			return err
		}

		// Then create a PR from that branch
		if err := CreatePullRequest(config); err != nil {
			return fmt.Errorf("failed to create pull request: %v", err)
		}
	}
	return nil
}

// commitChanges stages, commits, and pushes the specified files.
func commitChanges(config *config.GitConfig) error {
	// Stage files first
	if err := stageFiles(config.FilePattern); err != nil {
		return err
	}

	// Perform commit and push
	return performCommitAndPush(config)
}

// stageFiles adds the specified files to the Git staging area.
// It handles multiple file patterns separated by spaces.
func stageFiles(filePattern string) error {
	fmt.Printf("  • Adding files... ")

	// Handle multiple patterns separated by spaces
	if strings.Contains(filePattern, " ") {
		patterns := strings.Fields(filePattern)
		for _, pattern := range patterns {
			if err := executeGitAdd(pattern); err != nil {
				fmt.Println("❌ Failed")
				return fmt.Errorf("failed to add pattern %s: %v", pattern, err)
			}
		}
	} else {
		// Single pattern case
		if err := executeGitAdd(filePattern); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to add files: %v", err)
		}
	}

	fmt.Println("✅ Done")
	return nil
}

// executeGitAdd executes the git add command for a specific pattern.
func executeGitAdd(pattern string) error {
	addCmd := exec.Command(gitcmd.CmdGit, gitcmd.AddArgs(pattern)...)
	addCmd.Stdout = os.Stdout
	addCmd.Stderr = os.Stderr
	return addCmd.Run()
}

// performCommitAndPush commits the staged changes and pushes them to the remote.
func performCommitAndPush(config *config.GitConfig) error {
	commitPushCommands := []Command{
		{gitcmd.CmdGit, gitcmd.CommitArgs(config.CommitMessage), "Committing changes"},
		{gitcmd.CmdGit, gitcmd.PushArgs(gitcmd.RefOrigin, config.Branch), "Pushing to remote"},
	}

	return ExecuteCommandBatch(commitPushCommands, "")
}
