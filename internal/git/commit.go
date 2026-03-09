package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/somaz94/go-git-commit-action/internal/config"
	"github.com/somaz94/go-git-commit-action/internal/errors"
	"github.com/somaz94/go-git-commit-action/internal/gitcmd"
	"github.com/somaz94/go-git-commit-action/internal/output"
)

const (
	// File and directory permissions
	permDir  = 0755
	permFile = 0644

	// Retry configuration
	retryBaseDelay = time.Second
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
	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := operation(); err != nil {
				lastErr = err
				time.Sleep(retryBaseDelay * time.Duration(i+1))
				continue
			}
			return nil
		}
	}
	return errors.NewWithContext("operation failed after retries", maxRetries, lastErr)
}

// RunGitCommit executes the Git commit operation with the provided configuration.
// It wraps the entire process in a retry mechanism to handle transient failures.
func RunGitCommit(config *config.GitConfig, result *output.Result) error {
	// Create a context with timeout for the entire operation
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Timeout)*time.Second)
	defer cancel()

	// Wrap the entire commit process in retry logic
	return withRetry(ctx, config.RetryCount, func() error {
		return executeGitCommitWorkflow(config, result)
	})
}

// executeGitCommitWorkflow runs all steps of the Git commit process
func executeGitCommitWorkflow(config *config.GitConfig, result *output.Result) error {
	// Validate the configuration
	if err := config.Validate(); err != nil {
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
		fmt.Println("\n[WARN] No changes detected and skip_if_empty is true. Skipping commit process.")
		result.Set(output.KeySkipped, "true")
		result.Set(output.KeyChangedFiles, "0")
		return nil
	}

	result.Set(output.KeySkipped, "false")

	// Count changed files
	changedFiles := countChangedFiles()
	result.Set(output.KeyChangedFiles, fmt.Sprintf("%d", changedFiles))

	// Create a PR or commit directly based on configuration
	if config.CreatePR {
		return handlePullRequestFlow(config, result)
	}

	return commitChanges(config, result)
}

// printDebugInfo outputs debug information about the current environment.
// This includes the working directory and the contents of the directory.
func printDebugInfo() {
	currentDir, err := os.Getwd()
	if err != nil {
		currentDir = "(unknown)"
	}
	fmt.Println("\nStarting Git Commit Action\n" +
		"================================")

	fmt.Println("\nConfiguration:")
	fmt.Printf("  - Working Directory: %s\n", currentDir)

	fmt.Println("\nDirectory Contents:")
	files, err := os.ReadDir(".")
	if err != nil {
		fmt.Printf("  - (failed to read directory: %v)\n", err)
		return
	}
	for _, file := range files {
		fmt.Printf("  - %s\n", file.Name())
	}
}

// changeWorkingDirectory changes to the specified repository path if it's not
// the current directory. It reports the new directory after changing.
func changeWorkingDirectory(config *config.GitConfig) error {
	if config.RepoPath != "." {
		if err := os.Chdir(config.RepoPath); err != nil {
			return errors.NewWithPath("change directory", config.RepoPath, err)
		}
		newDir, _ := os.Getwd()
		fmt.Printf("\nChanged to directory: %s\n", newDir)
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

	if err := ExecuteCommandBatch(baseCommands, "\nExecuting Git Commands:"); err != nil {
		return err
	}

	// Setup git credentials for checkout@v6 compatibility
	if err := setupGitCredentials(config); err != nil {
		return err
	}

	// Show final git configuration
	fmt.Printf("  - Checking git configuration... ")
	cmd := exec.Command(gitcmd.CmdGit, gitcmd.ConfigListArgs()...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("FAILED")
		return err
	}
	fmt.Println("Done")

	return nil
}

// setupGitCredentials configures git credential helper for checkout@v6 compatibility.
// Since checkout@v6 stores credentials in $RUNNER_TEMP which is not accessible in Docker containers,
// we need to configure the remote URL with the token directly.
func setupGitCredentials(config *config.GitConfig) error {
	fmt.Printf("  - Configuring git credentials... ")

	// Get GitHub token from environment or config
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" && config.GitHubToken != "" {
		token = config.GitHubToken
	}

	if token == "" {
		fmt.Println("[WARN] No token found, skipping")
		return nil
	}

	// Get the repository URL from git remote
	cmd := exec.Command(gitcmd.CmdGit, gitcmd.ConfigGetArgs("remote.origin.url")...)
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("[WARN] Could not get remote URL, skipping")
		return nil
	}

	remoteURL := strings.TrimSpace(string(output))

	// Only process GitHub URLs
	if !strings.Contains(remoteURL, "github.com") {
		fmt.Println("[WARN] Not a GitHub repository, skipping")
		return nil
	}

	// Replace https:// with https://x-access-token:TOKEN@
	// This works for both checkout@v4 and checkout@v6
	var newURL string
	if strings.HasPrefix(remoteURL, "https://github.com/") {
		newURL = strings.Replace(remoteURL, "https://github.com/", fmt.Sprintf("https://x-access-token:%s@github.com/", token), 1)
	} else {
		fmt.Println("[WARN] Unsupported URL format, skipping")
		return nil
	}

	// Update the remote URL
	setURLCmd := exec.Command(gitcmd.CmdGit, gitcmd.RemoteSetURLArgs(gitcmd.RefOrigin, newURL)...)
	setURLCmd.Stderr = os.Stderr
	if err := setURLCmd.Run(); err != nil {
		fmt.Println("FAILED")
		return errors.New("set remote URL", err)
	}

	fmt.Println("Done")
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
	fmt.Printf("\n[WARN] Branch '%s' not found, creating it...\n", config.Branch)
	createCommands := []Command{
		{gitcmd.CmdGit, gitcmd.CheckoutNewBranchArgs(config.Branch), "Creating new branch"},
		{gitcmd.CmdGit, gitcmd.PushUpstreamArgs(gitcmd.RefOrigin, config.Branch), "Pushing new branch"},
	}

	return ExecuteCommandBatch(createCommands, "")
}

// checkoutRemoteBranch checks out an existing remote branch while handling
// local changes properly through backup, stash, and restore.
func checkoutRemoteBranch(config *config.GitConfig) error {
	fmt.Printf("\n[WARN] Checking out existing remote branch '%s'...\n", config.Branch)

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
		return "", errors.New("get git status", err)
	}
	return string(output), nil
}

// backupChanges creates backups of modified files that need to be preserved
// during branch switching.
func backupChanges(config *config.GitConfig, statusOutput string) ([]FileBackup, error) {
	fmt.Printf("  - Backing up changes... ")

	var backups []FileBackup

	for _, line := range strings.Split(statusOutput, "\n") {
		if len(line) < 4 {
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
			fmt.Println("FAILED")
			return nil, errors.NewWithPath("read file for backup", relPath, err)
		}

		backups = append(backups, FileBackup{path: relPath, content: content})
	}

	fmt.Println("Done")
	return backups, nil
}

// stashChanges safely stashes any local changes to avoid conflicts.
func stashChanges() error {
	fmt.Printf("  - Stashing changes... ")
	stashCmd := exec.Command(gitcmd.CmdGit, gitcmd.StashPushArgs()...)
	stashCmd.Stdout = os.Stdout
	stashCmd.Stderr = os.Stderr

	if err := stashCmd.Run(); err != nil {
		fmt.Println("FAILED")
		return errors.New("stash changes", err)
	}

	fmt.Println("Done")
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
	fmt.Printf("  - Restoring changes... ")

	for _, backup := range backups {
		// Create parent directories if they don't exist
		dir := filepath.Dir(backup.path)
		if dir != "." {
			if err := os.MkdirAll(dir, permDir); err != nil {
				fmt.Println("FAILED")
				return errors.NewWithPath("create directory", dir, err)
			}
		}

		// Write the file content
		if err := os.WriteFile(backup.path, backup.content, permFile); err != nil {
			fmt.Println("FAILED")
			return errors.NewWithPath("restore file", backup.path, err)
		}
	}

	fmt.Println("Done")
	return nil
}

// checkIfEmpty determines if there are any changes to commit.
// It checks both working directory changes and differences between branches.
func checkIfEmpty(config *config.GitConfig) (bool, error) {
	// Get local working directory changes
	statusCmd := exec.Command(gitcmd.CmdGit, gitcmd.StatusPorcelainArgs()...)
	statusOutput, err := statusCmd.Output()
	if err != nil {
		return false, errors.New("check git status", err)
	}

	// Check for differences between branches
	diffCmd := exec.Command(gitcmd.CmdGit, gitcmd.DiffNameOnlyArgs(
		fmt.Sprintf("origin/%s", config.PRBase),
		config.PRBranch,
	)...)
	diffOutput, err := diffCmd.Output()
	if err != nil {
		// Likely a new branch or missing remote ref; log and treat as non-empty to proceed
		fmt.Printf("  - [WARN] Branch diff failed (proceeding anyway): %v\n", err)
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
	fmt.Printf("\nChange Detection:\n")
	fmt.Printf("  - Local changes: %v\n", hasLocalChanges)
	fmt.Printf("  - Branch differences: %v\n", hasBranchDifferences)

	if hasLocalChanges {
		fmt.Printf("  - Local changes details:\n%s\n", string(statusOutput))
	}

	if hasBranchDifferences {
		fmt.Printf("  - Branch differences details:\n%s\n", string(diffOutput))
	}
}

// countChangedFiles counts the number of changed files in the working directory.
func countChangedFiles() int {
	statusCmd := exec.Command(gitcmd.CmdGit, gitcmd.StatusPorcelainArgs()...)
	statusOutput, err := statusCmd.Output()
	if err != nil {
		return 0
	}
	count := 0
	for _, line := range strings.Split(string(statusOutput), "\n") {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}

// handlePullRequestFlow manages the creation of pull requests
// based on the auto_branch configuration.
func handlePullRequestFlow(config *config.GitConfig, result *output.Result) error {
	if config.AutoBranch {
		// Auto branch creation and PR creation in one step
		if err := CreatePullRequest(config, result); err != nil {
			return errors.New("create pull request with auto branch", err)
		}
	} else {
		// First commit changes to the specified branch
		if err := commitChanges(config, result); err != nil {
			return err
		}

		// Then create a PR from that branch
		if err := CreatePullRequest(config, result); err != nil {
			return errors.New("create pull request", err)
		}
	}
	return nil
}

// commitChanges stages, commits, and pushes the specified files.
func commitChanges(config *config.GitConfig, result *output.Result) error {
	// Stage files first
	if err := StageFiles(config.FilePattern); err != nil {
		return err
	}

	// Perform commit and push
	if err := performCommitAndPush(config); err != nil {
		return err
	}

	// Capture commit SHA for output
	commitSHA, err := getCommitSHA()
	if err == nil {
		result.Set(output.KeyCommitSHA, commitSHA)
	}

	return nil
}

// getCommitSHA retrieves the current HEAD commit SHA.
func getCommitSHA() (string, error) {
	cmd := exec.Command(gitcmd.CmdGit, gitcmd.RevParseArgs("HEAD")...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// performCommitAndPush commits the staged changes and pushes them to the remote.
func performCommitAndPush(config *config.GitConfig) error {
	commitPushCommands := []Command{
		{gitcmd.CmdGit, gitcmd.CommitArgs(config.CommitMessage), "Committing changes"},
		{gitcmd.CmdGit, gitcmd.PushArgs(gitcmd.RefOrigin, config.Branch), "Pushing to remote"},
	}

	return ExecuteCommandBatch(commitPushCommands, "")
}
