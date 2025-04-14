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
)

// Helper function for retry logic
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

// RunGitCommit runs the Git commit operation.
func RunGitCommit(config *config.GitConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Timeout)*time.Second)
	defer cancel()

	// Wrap the existing code in retry logic
	return withRetry(ctx, config.RetryCount, func() error {
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
		if empty, err := checkIfEmpty(config); err != nil {
			return err
		} else if empty {
			fmt.Println("\n⚠️  No changes detected and skip_if_empty is true. Skipping commit process.")
			return nil
		}

		// Create a PR or commit directly
		if config.CreatePR {
			return handlePullRequestFlow(config)
		} else {
			return commitChanges(config)
		}
	})
}

// validateConfig validates the required configuration.
func validateConfig(config *config.GitConfig) error {
	if config.CreatePR {
		if !config.AutoBranch && config.PRBranch == "" {
			return fmt.Errorf("pr_branch must be specified when auto_branch is false and create_pr is true")
		}
		if config.PRBase == "" {
			return fmt.Errorf("pr_base must be specified when create_pr is true")
		}
		if config.GitHubToken == "" {
			return fmt.Errorf("github_token must be specified when create_pr is true")
		}
	}
	return nil
}

// printDebugInfo prints debug information.
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

// changeWorkingDirectory changes the working directory.
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

// setupGitConfig performs Git basic configuration.
func setupGitConfig(config *config.GitConfig) error {
	baseCommands := []struct {
		name string
		args []string
		desc string
	}{
		{"git", []string{"config", "--global", "--add", "safe.directory", "/app"}, "Setting safe directory (/app)"},
		{"git", []string{"config", "--global", "--add", "safe.directory", "/github/workspace"}, "Setting safe directory (/github/workspace)"},
		{"git", []string{"config", "--global", "user.email", config.UserEmail}, "Configuring user email"},
		{"git", []string{"config", "--global", "user.name", config.UserName}, "Configuring user name"},
		{"git", []string{"config", "--global", "--list"}, "Checking git configuration"},
	}

	fmt.Println("\n⚙️  Executing Git Commands:")
	for _, cmd := range baseCommands {
		fmt.Printf("  • %s... ", cmd.desc)
		command := exec.Command(cmd.name, cmd.args...)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr

		if err := command.Run(); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to execute %s: %v", cmd.name, err)
		}
		fmt.Println("✅ Done")
	}
	return nil
}

// handleBranch handles branch-related operations.
func handleBranch(config *config.GitConfig) error {
	// Check the local branch
	checkLocalBranch := exec.Command("git", "rev-parse", "--verify", config.Branch)
	// Check the remote branch
	checkRemoteBranch := exec.Command("git", "ls-remote", "--heads", "origin", config.Branch)

	if checkLocalBranch.Run() != nil && checkRemoteBranch.Run() != nil {
		// If both local and remote branches are missing, create a new one
		return createNewBranch(config)
	} else if checkLocalBranch.Run() != nil {
		// If the remote branch exists but the local one does not, checkout the remote branch
		return checkoutRemoteBranch(config)
	}
	return nil
}

// createNewBranch creates a new branch.
func createNewBranch(config *config.GitConfig) error {
	fmt.Printf("\n⚠️  Branch '%s' not found, creating it...\n", config.Branch)
	createCommands := []struct {
		name string
		args []string
		desc string
	}{
		{"git", []string{"checkout", "-b", config.Branch}, "Creating new branch"},
		{"git", []string{"push", "-u", "origin", config.Branch}, "Pushing new branch"},
	}

	for _, cmd := range createCommands {
		fmt.Printf("  • %s... ", cmd.desc)
		command := exec.Command(cmd.name, cmd.args...)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr

		if err := command.Run(); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to execute %s: %v", cmd.name, err)
		}
		fmt.Println("✅ Done")
	}
	return nil
}

// checkoutRemoteBranch checks out the remote branch.
func checkoutRemoteBranch(config *config.GitConfig) error {
	fmt.Printf("\n⚠️  Checking out existing remote branch '%s'...\n", config.Branch)

	// Check for modified files
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusOutput, err := statusCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get modified files: %v", err)
	}

	// Backup changes
	backups, err := backupChanges(config, string(statusOutput))
	if err != nil {
		return err
	}

	// Stash changes
	if err := stashChanges(); err != nil {
		return err
	}

	// Check out the remote branch
	if err := fetchAndCheckout(config); err != nil {
		return err
	}

	// Restore changes
	return restoreChanges(backups)
}

// FileBackup is a struct for file backups.
type FileBackup struct {
	path    string
	content []byte
}

// backupChanges backs up changed files.
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

		// Backup only if the file is not deleted
		if status != " D" && status != "D " {
			content, err := os.ReadFile(relPath)
			if err != nil {
				fmt.Println("❌ Failed")
				return nil, fmt.Errorf("failed to read file %s: %v", relPath, err)
			}
			backups = append(backups, FileBackup{path: relPath, content: content})
		}
	}
	fmt.Println("✅ Done")
	return backups, nil
}

// stashChanges stashes changes.
func stashChanges() error {
	fmt.Printf("  • Stashing changes... ")
	stashCmd := exec.Command("git", "stash", "push", "-u")
	stashCmd.Stdout = os.Stdout
	stashCmd.Stderr = os.Stderr
	if err := stashCmd.Run(); err != nil {
		fmt.Println("❌ Failed")
		return fmt.Errorf("failed to stash changes: %v", err)
	}
	fmt.Println("✅ Done")
	return nil
}

// fetchAndCheckout fetches and checks out the remote branch.
func fetchAndCheckout(config *config.GitConfig) error {
	checkoutCommands := []struct {
		name string
		args []string
		desc string
	}{
		{"git", []string{"fetch", "origin", config.Branch}, "Fetching remote branch"},
		{"git", []string{"checkout", config.Branch}, "Checking out branch"},
		{"git", []string{"reset", "--hard", fmt.Sprintf("origin/%s", config.Branch)}, "Resetting to remote state"},
	}

	for _, cmd := range checkoutCommands {
		fmt.Printf("  • %s... ", cmd.desc)
		command := exec.Command(cmd.name, cmd.args...)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		if err := command.Run(); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to execute %s: %v", cmd.name, err)
		}
		fmt.Println("✅ Done")
	}
	return nil
}

// restoreChanges restores backed up changes.
func restoreChanges(backups []FileBackup) error {
	fmt.Printf("  • Restoring changes... ")
	for _, backup := range backups {
		// Create the directory if it doesn't exist
		dir := filepath.Dir(backup.path)
		if dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Println("❌ Failed")
				return fmt.Errorf("failed to create directory %s: %v", dir, err)
			}
		}

		if err := os.WriteFile(backup.path, backup.content, 0644); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to restore file %s: %v", backup.path, err)
		}
	}
	fmt.Println("✅ Done")
	return nil
}

// checkIfEmpty checks if there are any changes.
func checkIfEmpty(config *config.GitConfig) (bool, error) {
	// Check for local changes in the working directory
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusOutput, err := statusCmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check git status: %v", err)
	}

	// Check for differences between the branch
	diffCmd := exec.Command("git", "diff", fmt.Sprintf("origin/%s...%s", config.PRBase, config.PRBranch), "--name-only")
	diffOutput, err := diffCmd.Output()
	if err != nil {
		// If an error occurs (e.g., new branch), consider it non-empty
		diffOutput = []byte("new-branch")
	}

	isEmpty := len(statusOutput) == 0 && len(diffOutput) == 0

	// Print debug information
	fmt.Printf("\n📊 Change Detection:\n")
	fmt.Printf("  • Local changes: %v\n", len(statusOutput) > 0)
	fmt.Printf("  • Branch differences: %v\n", len(diffOutput) > 0)
	if len(statusOutput) > 0 {
		fmt.Printf("  • Local changes details:\n%s\n", string(statusOutput))
	}
	if len(diffOutput) > 0 {
		fmt.Printf("  • Branch differences details:\n%s\n", string(diffOutput))
	}

	return isEmpty && config.SkipIfEmpty, nil
}

// handlePullRequestFlow handles the PR creation flow.
func handlePullRequestFlow(config *config.GitConfig) error {
	if config.AutoBranch {
		// If AutoBranch is true, the PR creation function will create a new branch and commit
		if err := CreatePullRequest(config); err != nil {
			return fmt.Errorf("failed to create pull request: %v", err)
		}
	} else {
		// If AutoBranch is false, commit the changes first and then create a PR
		if err := commitChanges(config); err != nil {
			return err
		}

		// Commit and then create a PR (use pr_branch and pr_base)
		if err := CreatePullRequest(config); err != nil {
			return fmt.Errorf("failed to create pull request: %v", err)
		}
	}
	return nil
}

// commitChanges commits and pushes the changes.
func commitChanges(config *config.GitConfig) error {
	// Handle git add with file pattern that may contain spaces
	fmt.Printf("  • Adding files... ")

	// Check if file pattern contains spaces, indicating multiple files/patterns
	if strings.Contains(config.FilePattern, " ") {
		// Split the pattern and add each file/pattern individually
		patterns := strings.Fields(config.FilePattern)
		for _, pattern := range patterns {
			addCmd := exec.Command("git", "add", pattern)
			addCmd.Stdout = os.Stdout
			addCmd.Stderr = os.Stderr

			if err := addCmd.Run(); err != nil {
				fmt.Println("❌ Failed")
				return fmt.Errorf("failed to add pattern %s: %v", pattern, err)
			}
		}
		fmt.Println("✅ Done")
	} else {
		// Single file pattern case - proceed as before
		addCmd := exec.Command("git", "add", config.FilePattern)
		addCmd.Stdout = os.Stdout
		addCmd.Stderr = os.Stderr

		if err := addCmd.Run(); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to add files: %v", err)
		}
		fmt.Println("✅ Done")
	}

	// Continue with commit and push commands
	commitPushCommands := []struct {
		name string
		args []string
		desc string
	}{
		{"git", []string{"commit", "-m", config.CommitMessage}, "Committing changes"},
		{"git", []string{"push", "origin", config.Branch}, "Pushing to remote"},
	}

	for _, cmd := range commitPushCommands {
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
