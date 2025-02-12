package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/somaz94/go-git-commit-action/internal/config"
)

type FileBackup struct {
	path    string
	content []byte
}

// setupGitEnvironment Git 환경 설정
func setupGitEnvironment(config *config.GitConfig) error {
	currentDir, _ := os.Getwd()
	fmt.Println("\n🚀 Starting Git Commit Action\n" +
		"================================")

	// Configuration Info
	printConfiguration(config, currentDir)

	// Directory Contents
	printDirectoryContents()

	// Change Directory
	if err := changeWorkingDirectory(config); err != nil {
		return err
	}

	// Git Configuration
	return configureGitSettings(config)
}

// handleBranch 브랜치 처리
func handleBranch(config *config.GitConfig) error {
	// Branch Existence Check
	checkLocalBranch := exec.Command("git", "rev-parse", "--verify", config.Branch)
	checkRemoteBranch := exec.Command("git", "ls-remote", "--heads", "origin", config.Branch)

	if checkLocalBranch.Run() != nil && checkRemoteBranch.Run() != nil {
		return createNewBranch(config)
	} else if checkLocalBranch.Run() != nil {
		return checkoutExistingBranch(config)
	}
	return nil
}

// commitChanges 변경사항 커밋
func commitChanges(config *config.GitConfig) error {
	commitCommands := []struct {
		name string
		args []string
		desc string
	}{
		{"git", []string{"add", config.FilePattern}, "Adding files"},
		{"git", []string{"commit", "-m", config.CommitMessage}, "Committing changes"},
		{"git", []string{"push", "origin", config.Branch}, "Pushing to remote"},
	}

	return executeGitCommands(commitCommands)
}

// RunGitCommit 메인 Git 커밋 함수
func RunGitCommit(config *config.GitConfig) error {
	if err := setupGitEnvironment(config); err != nil {
		return err
	}

	if err := handleBranch(config); err != nil {
		return err
	}

	// 파일 백업
	backups, err := backupFiles(config.FilePattern)
	if err != nil {
		return err
	}

	// 변경사항 커밋 시도
	if err := commitChanges(config); err != nil {
		// 실패시 파일 복원
		if restoreErr := restoreFiles(backups); restoreErr != nil {
			return fmt.Errorf("failed to restore files after commit error: %v (original error: %v)", restoreErr, err)
		}
		return err
	}

	return nil
}

// backupFiles 파일 백업
func backupFiles(pattern string) ([]FileBackup, error) {
	fmt.Println("\n💾 Backing up files...")
	var backups []FileBackup

	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to glob files: %v", err)
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %v", file, err)
		}
		backups = append(backups, FileBackup{
			path:    file,
			content: content,
		})
		fmt.Printf("  • Backed up: %s\n", file)
	}

	return backups, nil
}

// restoreFiles 파일 복원
func restoreFiles(backups []FileBackup) error {
	fmt.Println("\n♻️  Restoring files...")
	for _, backup := range backups {
		dir := filepath.Dir(backup.path)
		if dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %v", dir, err)
			}
		}
		if err := os.WriteFile(backup.path, backup.content, 0644); err != nil {
			return fmt.Errorf("failed to restore file %s: %v", backup.path, err)
		}
		fmt.Printf("  • Restored: %s\n", backup.path)
	}
	return nil
}

// Helper functions
func printConfiguration(config *config.GitConfig, currentDir string) {
	fmt.Println("\n📋 Configuration:")
	fmt.Printf("  • Working Directory: %s\n", currentDir)
	fmt.Printf("  • User Email: %s\n", config.UserEmail)
	fmt.Printf("  • User Name: %s\n", config.UserName)
	fmt.Printf("  • Commit Message: %s\n", config.CommitMessage)
	fmt.Printf("  • Target Branch: %s\n", config.Branch)
	fmt.Printf("  • Repository Path: %s\n", config.RepoPath)
	fmt.Printf("  • File Pattern: %s\n", config.FilePattern)
}

func printDirectoryContents() {
	fmt.Println("\n📁 Directory Contents:")
	files, _ := os.ReadDir(".")
	for _, file := range files {
		fmt.Printf("  • %s\n", file.Name())
	}
}

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

func configureGitSettings(config *config.GitConfig) error {
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

	return executeGitCommands(baseCommands)
}

func executeGitCommands(commands []struct {
	name string
	args []string
	desc string
}) error {
	for _, cmd := range commands {
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

// createNewBranch 새 브랜치 생성
func createNewBranch(config *config.GitConfig) error {
	fmt.Printf("  • Creating new branch %s... ", config.Branch)

	commands := []struct {
		name string
		args []string
		desc string
	}{
		{"git", []string{"checkout", "-b", config.Branch}, "Creating branch"},
		{"git", []string{"push", "-u", "origin", config.Branch}, "Pushing branch to remote"},
	}

	for _, cmd := range commands {
		command := exec.Command(cmd.name, cmd.args...)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr

		if err := command.Run(); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to %s: %v", cmd.desc, err)
		}
	}

	fmt.Println("✅ Done")
	return nil
}

// checkoutExistingBranch 기존 브랜치 체크아웃
func checkoutExistingBranch(config *config.GitConfig) error {
	fmt.Printf("  • Checking out existing branch %s... ", config.Branch)

	commands := []struct {
		name string
		args []string
		desc string
	}{
		{"git", []string{"fetch", "origin", config.Branch}, "Fetching branch"},
		{"git", []string{"checkout", "-b", config.Branch, fmt.Sprintf("origin/%s", config.Branch)}, "Checking out branch"},
	}

	for _, cmd := range commands {
		command := exec.Command(cmd.name, cmd.args...)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr

		if err := command.Run(); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to %s: %v", cmd.desc, err)
		}
	}

	fmt.Println("✅ Done")
	return nil
}
