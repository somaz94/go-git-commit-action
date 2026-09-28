package git

import (
	"context"
	stderrors "errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/somaz94/go-git-commit-action/internal/config"
	"github.com/somaz94/go-git-commit-action/internal/errors"
	"github.com/somaz94/go-git-commit-action/internal/git/shared"
	"github.com/somaz94/go-git-commit-action/internal/gitcmd"
	"github.com/somaz94/go-git-commit-action/internal/output"
)

const (
	// File and directory permissions
	permDir  = 0755
	permFile = 0644
)

// retryBaseDelay is the linear backoff unit between attempts.
var retryBaseDelay = time.Second

// FileBackup is a struct for file backups.
type FileBackup struct {
	path    string
	content []byte
}

// withRetry provides retry logic for operations that might fail transiently.
// It executes the given operation until it succeeds or maxRetries attempts
// (at least one) have failed, backing off linearly between attempts.
// An ErrPushAfterCommit or errPRStage is returned at once, without a rerun.
func withRetry(ctx context.Context, maxRetries int, operation func() error) error {
	attempts := max(maxRetries, 1)
	var lastErr error
	for i := range attempts {
		if i > 0 {
			// Honor context cancellation during backoff instead of
			// blocking for the full linear delay.
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(retryBaseDelay * time.Duration(i)):
			}
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		err := operation()
		if err == nil {
			return nil
		}
		// Both were already retried in place, past a commit: a rerun
		// finds a clean tree and skips, or cuts a second auto branch.
		if stderrors.Is(err, shared.ErrPushAfterCommit) || stderrors.Is(err, errPRStage) {
			return err
		}
		lastErr = err
	}
	return errors.NewWithContext("operation failed after retries", attempts, lastErr)
}

// RunGitCommit executes the Git commit operation with the provided configuration.
// It wraps the entire process in a retry mechanism to handle transient failures.
func RunGitCommit(ctx context.Context, config *config.GitConfig, result *output.Result) error {
	return RunGitCommitWithRunner(ctx, gitcmd.NewExecRunner(), config, result)
}

// RunGitCommitWithRunner is RunGitCommit with an explicit command Runner.
// Tests use it to drive the full workflow against a fake instead of a real
// repository; production callers should use RunGitCommit.
func RunGitCommitWithRunner(ctx context.Context, r gitcmd.Runner, config *config.GitConfig, result *output.Result) error {
	originalDir, err := os.Getwd()
	if err != nil {
		return errors.New("get working directory", err)
	}

	// Derive the timeout context from the caller's context so that an
	// upstream SIGINT/SIGTERM cancellation aborts an in-flight commit.
	ctx, cancel := context.WithTimeout(ctx, time.Duration(config.Timeout)*time.Second)
	defer cancel()

	return withRetry(ctx, config.RetryCount, func() error {
		// Restore original working directory before each attempt
		// to prevent relative path issues (e.g., chdir("test") applied twice)
		if err := os.Chdir(originalDir); err != nil {
			return errors.NewWithPath("restore working directory", originalDir, err)
		}
		return executeGitCommitWorkflow(ctx, r, config, result)
	})
}

// executeGitCommitWorkflow runs all steps of the Git commit process
func executeGitCommitWorkflow(ctx context.Context, r gitcmd.Runner, config *config.GitConfig, result *output.Result) error {
	if err := config.Validate(); err != nil {
		return err
	}

	if config.Debug {
		printDebugInfo()
	}

	if err := changeWorkingDirectory(config); err != nil {
		return err
	}

	if err := setupGitConfig(r, config); err != nil {
		return err
	}

	if err := handleBranch(r, config); err != nil {
		return err
	}

	isEmpty, err := checkIfEmpty(r, config)
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

	changedFiles := countChangedFiles(r)
	result.Set(output.KeyChangedFiles, fmt.Sprintf("%d", changedFiles))

	if config.CreatePR {
		return handlePullRequestFlow(ctx, r, config, result)
	}

	return commitChanges(ctx, r, config, result)
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
func setupGitConfig(r gitcmd.Runner, config *config.GitConfig) error {
	baseCommands := []Command{
		{gitcmd.CmdGit, gitcmd.ConfigSafeDirArgs(gitcmd.PathApp), "Setting safe directory (/app)"},
		{gitcmd.CmdGit, gitcmd.ConfigSafeDirArgs(gitcmd.PathGitHubWorkspace), "Setting safe directory (/github/workspace)"},
		{gitcmd.CmdGit, gitcmd.ConfigUserEmailArgs(config.UserEmail), "Configuring user email"},
		{gitcmd.CmdGit, gitcmd.ConfigUserNameArgs(config.UserName), "Configuring user name"},
	}

	if err := ExecuteCommandBatch(r, baseCommands, "\nExecuting Git Commands:"); err != nil {
		return err
	}

	if err := setupGitCredentials(r, config); err != nil {
		return err
	}

	if err := shared.RunStep(r, "Checking git configuration", gitcmd.CmdGit, gitcmd.ConfigListArgs()...); err != nil {
		return err
	}

	return nil
}

// setupGitCredentials embeds the token in origin's URL for checkout@v6 compatibility.
// Since checkout@v6 stores credentials in $RUNNER_TEMP which is not accessible in Docker containers,
// we need to configure the remote URL with the token directly.
func setupGitCredentials(r gitcmd.Runner, config *config.GitConfig) error {
	fmt.Printf("  - Configuring git credentials... ")

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" && config.GitHubToken != "" {
		token = config.GitHubToken
	}

	if token == "" {
		fmt.Println("[WARN] No token found, skipping")
		return nil
	}

	output, err := r.Output(gitcmd.CmdGit, gitcmd.ConfigGetArgs("remote.origin.url")...)
	if err != nil {
		fmt.Println("[WARN] Could not get remote URL, skipping")
		return nil
	}

	remoteURL := strings.TrimSpace(string(output))

	if !strings.Contains(remoteURL, "github.com") {
		fmt.Println("[WARN] Not a GitHub repository, skipping")
		return nil
	}

	// x-access-token URL auth works under both checkout@v4 and checkout@v6.
	var newURL string
	if strings.HasPrefix(remoteURL, "https://github.com/") {
		newURL = strings.Replace(remoteURL, "https://github.com/", fmt.Sprintf("https://x-access-token:%s@github.com/", token), 1)
	} else {
		fmt.Println("[WARN] Unsupported URL format, skipping")
		return nil
	}

	if err := r.Run(gitcmd.CmdGit, gitcmd.RemoteSetURLArgs(gitcmd.RefOrigin, newURL)...); err != nil {
		fmt.Println("FAILED")
		return errors.New("set remote URL", err)
	}

	fmt.Println("Done")
	return nil
}

// handleBranch manages branch-related operations, checking for local and remote
// branch existence and taking appropriate action.
func handleBranch(r gitcmd.Runner, config *config.GitConfig) error {
	// These are existence probes, so Output is used rather than Run: only the
	// exit status matters and the command's own output must stay off the log.
	_, localErr := r.Output(gitcmd.CmdGit, gitcmd.RevParseArgs(config.Branch)...)
	localBranchExists := localErr == nil

	// "git ls-remote --heads" exits 0 with empty output when nothing matches, so
	// the exit status alone would report every branch as existing whenever the
	// remote is merely reachable. The listing itself is the answer.
	remoteRefs, remoteErr := r.Output(gitcmd.CmdGit, gitcmd.LsRemoteHeadsArgs(gitcmd.RefOrigin, config.Branch)...)
	remoteBranchExists := remoteErr == nil && len(strings.TrimSpace(string(remoteRefs))) > 0

	if !localBranchExists && !remoteBranchExists {
		return createNewBranch(r, config)
	} else if !localBranchExists && remoteBranchExists {
		return checkoutRemoteBranch(r, config)
	}

	// Local branch exists: assumed to be the current checkout; not switched here.
	return nil
}

// createNewBranch creates a new branch and pushes it to the remote repository.
func createNewBranch(r gitcmd.Runner, config *config.GitConfig) error {
	fmt.Printf("\n[WARN] Branch '%s' not found, creating it...\n", config.Branch)
	createCommands := []Command{
		{gitcmd.CmdGit, gitcmd.CheckoutNewBranchArgs(config.Branch), "Creating new branch"},
		{gitcmd.CmdGit, gitcmd.PushUpstreamArgs(gitcmd.RefOrigin, config.Branch), "Pushing new branch"},
	}

	return ExecuteCommandBatch(r, createCommands, "")
}

// checkoutRemoteBranch checks out an existing remote branch while handling
// local changes properly through backup, stash, and restore.
func checkoutRemoteBranch(r gitcmd.Runner, config *config.GitConfig) error {
	fmt.Printf("\n[WARN] Checking out existing remote branch '%s'...\n", config.Branch)

	statusOutput, err := getGitStatus(r)
	if err != nil {
		return err
	}

	backups, err := backupChanges(config, statusOutput)
	if err != nil {
		return err
	}

	// The stash only clears the tree for checkout and is never popped;
	// restoreChanges replays the in-memory backup instead.
	if err := stashChanges(r); err != nil {
		return err
	}

	if err := fetchAndCheckout(r, config); err != nil {
		return err
	}

	return restoreChanges(backups)
}

// getGitStatus returns the current Git status in porcelain format.
func getGitStatus(r gitcmd.Runner) (string, error) {
	output, err := r.Output(gitcmd.CmdGit, gitcmd.StatusPorcelainArgs()...)
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

		status := line[:2]
		fullPath := strings.TrimSpace(line[3:])

		// Porcelain paths are repo-root-relative, but cwd is already RepoPath.
		relPath := fullPath
		if config.RepoPath != "." {
			relPath = strings.TrimPrefix(fullPath, config.RepoPath+"/")
		}

		fmt.Printf("\n    - Found modified file: %s (status: %s)", relPath, status)

		// Skip deleted files since they don't need backup
		if status == " D" || status == "D " {
			continue
		}

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
func stashChanges(r gitcmd.Runner) error {
	if err := shared.RunStep(r, "Stashing changes", gitcmd.CmdGit, gitcmd.StashPushArgs()...); err != nil {
		return errors.New("stash changes", err)
	}

	return nil
}

// fetchAndCheckout fetches the remote branch and checks it out locally.
func fetchAndCheckout(r gitcmd.Runner, config *config.GitConfig) error {
	checkoutCommands := []Command{
		{gitcmd.CmdGit, gitcmd.FetchArgs(gitcmd.RefOrigin, config.Branch), "Fetching remote branch"},
		{gitcmd.CmdGit, gitcmd.CheckoutArgs(config.Branch), "Checking out branch"},
		{gitcmd.CmdGit, gitcmd.ResetHardArgs(fmt.Sprintf("origin/%s", config.Branch)), "Resetting to remote state"},
	}

	return ExecuteCommandBatch(r, checkoutCommands, "")
}

// restoreChanges brings back the backed up files after branch switching.
func restoreChanges(backups []FileBackup) error {
	fmt.Printf("  - Restoring changes... ")

	for _, backup := range backups {
		dir := filepath.Dir(backup.path)
		if dir != "." {
			if err := os.MkdirAll(dir, permDir); err != nil {
				fmt.Println("FAILED")
				return errors.NewWithPath("create directory", dir, err)
			}
		}

		if err := os.WriteFile(backup.path, backup.content, permFile); err != nil {
			fmt.Println("FAILED")
			return errors.NewWithPath("restore file", backup.path, err)
		}
	}

	fmt.Println("Done")
	return nil
}

// checkIfEmpty determines if there are any local changes to commit.
// The skip decision is based solely on local working directory changes (git status).
// Branch differences are logged for informational purposes but do not affect the skip logic,
// since existing branch differences are about PR content, not about new uncommitted work.
func checkIfEmpty(r gitcmd.Runner, config *config.GitConfig) (bool, error) {
	statusOutput, err := r.Output(gitcmd.CmdGit, gitcmd.StatusPorcelainArgs()...)
	if err != nil {
		return false, errors.New("check git status", err)
	}

	hasLocalChanges := len(statusOutput) > 0

	var hasBranchDifferences bool
	diffOutput, err := r.Output(gitcmd.CmdGit, gitcmd.DiffNameOnlyArgs(
		fmt.Sprintf("origin/%s", config.PRBase),
		config.PRBranch,
	)...)
	if err != nil {
		fmt.Printf("  - [WARN] Branch diff failed (proceeding anyway): %v\n", err)
		hasBranchDifferences = false
	} else {
		hasBranchDifferences = len(diffOutput) > 0
	}

	printChangeDetectionInfo(statusOutput, diffOutput, hasLocalChanges, hasBranchDifferences)

	return !hasLocalChanges && config.SkipIfEmpty, nil
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
func countChangedFiles(r gitcmd.Runner) int {
	statusOutput, err := r.Output(gitcmd.CmdGit, gitcmd.StatusPorcelainArgs()...)
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
func handlePullRequestFlow(ctx context.Context, r gitcmd.Runner, config *config.GitConfig, result *output.Result) error {
	if config.AutoBranch {
		if err := CreatePullRequest(ctx, r, config, result); err != nil {
			return errors.New("create pull request with auto branch", err)
		}
	} else {
		// In dry run mode, skip actual commit/push since we only simulate PR creation
		if !config.PRDryRun {
			if err := commitChanges(ctx, r, config, result); err != nil {
				return err
			}
		}

		if err := CreatePullRequest(ctx, r, config, result); err != nil {
			return errors.New("create pull request", err)
		}
	}
	return nil
}

// commitChanges stages, commits, and pushes the specified files.
func commitChanges(ctx context.Context, r gitcmd.Runner, config *config.GitConfig, result *output.Result) error {
	if err := StageFiles(r, config.FilePattern); err != nil {
		return err
	}

	// Existing tracked branch (no upstream flag); an empty commit skips the push instead of failing.
	if err := shared.CommitAndPush(ctx, r, config.CommitMessage, config.Branch, shared.CommitPushOptions{
		TolerateNothingToCommit: true,
		PushAttempts:            config.RetryCount,
	}); err != nil {
		return err
	}

	// Capture commit SHA for output
	commitSHA, err := shared.CurrentCommitSHA(r)
	if err == nil {
		result.Set(output.KeyCommitSHA, commitSHA)
	}

	return nil
}
