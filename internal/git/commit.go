package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/somaz94/go-git-commit-action/internal/config"
)

func RunGitCommit(config *config.GitConfig) error {
	// Debug information
	currentDir, _ := os.Getwd()
	fmt.Println("\n🚀 Starting Git Commit Action\n" +
		"================================")

	// Configuration Info
	fmt.Println("\n📋 Configuration:")
	fmt.Printf("  • Working Directory: %s\n", currentDir)
	fmt.Printf("  • User Email: %s\n", config.UserEmail)
	fmt.Printf("  • User Name: %s\n", config.UserName)
	fmt.Printf("  • Commit Message: %s\n", config.CommitMessage)
	fmt.Printf("  • Target Branch: %s\n", config.Branch)
	fmt.Printf("  • Repository Path: %s\n", config.RepoPath)
	fmt.Printf("  • File Pattern: %s\n", config.FilePattern)

	// Directory Contents
	fmt.Println("\n📁 Directory Contents:")
	files, _ := os.ReadDir(".")
	for _, file := range files {
		fmt.Printf("  • %s\n", file.Name())
	}

	// Change Directory
	if config.RepoPath != "." {
		if err := os.Chdir(config.RepoPath); err != nil {
			return fmt.Errorf("❌ Failed to change directory: %v", err)
		}
		newDir, _ := os.Getwd()
		fmt.Printf("\n📂 Changed to directory: %s\n", newDir)
	}

	// Git Operations
	fmt.Println("\n⚙️  Executing Git Commands:")
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

	// Git Configuration
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

	// Branch Existence Check
	checkLocalBranch := exec.Command("git", "rev-parse", "--verify", config.Branch)
	checkRemoteBranch := exec.Command("git", "ls-remote", "--heads", "origin", config.Branch)

	if checkLocalBranch.Run() != nil && checkRemoteBranch.Run() != nil {
		// New only if there are no branches in both local and remote
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
	} else if checkLocalBranch.Run() != nil {
		// Remote exists but local does not exist
		fmt.Printf("\n⚠️  Checking out existing remote branch '%s'...\n", config.Branch)

		// Get modified files
		fmt.Printf("  • Checking modified files... ")
		statusCmd := exec.Command("git", "status", "--porcelain")
		statusOutput, err := statusCmd.Output()
		if err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to get modified files: %v", err)
		}
		fmt.Println("✅ Done")

		// Backup modified files
		type FileBackup struct {
			path    string
			content []byte
		}
		var backups []FileBackup

		fmt.Printf("  • Backing up changes... ")
		for _, line := range strings.Split(string(statusOutput), "\n") {
			if line == "" {
				continue
			}
			// Split status code and file path
			status := line[:2]
			fullPath := strings.TrimSpace(line[3:])

			// Calculate relative path based on config.RepoPath
			relPath := fullPath
			if config.RepoPath != "." {
				relPath = strings.TrimPrefix(fullPath, config.RepoPath+"/")
			}

			fmt.Printf("\n    - Found modified file: %s (status: %s)", relPath, status)

			// Backup only if not deleted
			if status != " D" && status != "D " {
				content, err := os.ReadFile(relPath)
				if err != nil {
					fmt.Println("❌ Failed")
					return fmt.Errorf("failed to read file %s: %v", relPath, err)
				}
				backups = append(backups, FileBackup{path: relPath, content: content})
			}
		}
		fmt.Println("✅ Done")

		// Stash changes
		fmt.Printf("  • Stashing changes... ")
		stashCmd := exec.Command("git", "stash", "push", "-u")
		stashCmd.Stdout = os.Stdout
		stashCmd.Stderr = os.Stderr
		if err := stashCmd.Run(); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to stash changes: %v", err)
		}
		fmt.Println("✅ Done")

		// Checkout remote branch
		checkoutCommands := []struct {
			name string
			args []string
			desc string
		}{
			{"git", []string{"fetch", "origin", config.Branch}, "Fetching remote branch"},
			{"git", []string{"checkout", config.Branch}, "Checking out branch"}, // -b 옵션 제거
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

		// Restore changes
		fmt.Printf("  • Restoring changes... ")
		for _, backup := range backups {
			// Create directory if necessary
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
	}

	// Different actions depending on whether PR is generated or not
	if config.CreatePR {
		if err := CreatePullRequest(config); err != nil {
			return fmt.Errorf("failed to create pull request: %v", err)
		}
	} else {
		// Commit and push if no PR is created
		commitCommands := []struct {
			name string
			args []string
			desc string
		}{
			{"git", []string{"add", config.FilePattern}, "Adding files"},
			{"git", []string{"commit", "-m", config.CommitMessage}, "Committing changes"},
			{"git", []string{"push", "origin", config.Branch}, "Pushing to remote"},
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
	}

	return nil
}
