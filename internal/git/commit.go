package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/somaz94/go-git-commit-action/internal/config"
)

// Git 환경 설정
func SetupGitEnvironment(config *config.GitConfig) error {
	currentDir, _ := os.Getwd()
	fmt.Println("\n🚀 Starting Git Commit Action\n" +
		"================================")

	// Configuration Info 출력
	printConfiguration(config, currentDir)

	// 디렉토리 변경
	if err := changeDirectory(config); err != nil {
		return err
	}

	// Git 기본 설정
	return setupGitConfig(config)
}

func printConfiguration(config *config.GitConfig, currentDir string) {
	fmt.Println("\n📋 Configuration:")
	fmt.Printf("  • Working Directory: %s\n", currentDir)
	fmt.Printf("  • User Email: %s\n", config.UserEmail)
	fmt.Printf("  • User Name: %s\n", config.UserName)
	fmt.Printf("  • Commit Message: %s\n", config.CommitMessage)
	fmt.Printf("  • Target Branch: %s\n", config.Branch)
	fmt.Printf("  • Repository Path: %s\n", config.RepoPath)
	fmt.Printf("  • File Pattern: %s\n", config.FilePattern)

	fmt.Println("\n📁 Directory Contents:")
	files, _ := os.ReadDir(".")
	for _, file := range files {
		fmt.Printf("  • %s\n", file.Name())
	}
}

func changeDirectory(config *config.GitConfig) error {
	if config.RepoPath != "." {
		if err := os.Chdir(config.RepoPath); err != nil {
			return fmt.Errorf("❌ Failed to change directory: %v", err)
		}
		newDir, _ := os.Getwd()
		fmt.Printf("\n📂 Changed to directory: %s\n", newDir)
	}
	return nil
}

func setupGitConfig(config *config.GitConfig) error {
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

	return executeCommands(baseCommands)
}

// 실제 커밋 실행
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
		// Create new branch
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
		// Checkout existing remote branch
		fmt.Printf("\n⚠️  Checking out existing remote branch '%s'...\n", config.Branch)

		// Backup current changes
		if err := backupAndStashChanges(); err != nil {
			return err
		}

		checkoutCommands := []struct {
			name string
			args []string
			desc string
		}{
			{"git", []string{"fetch", "origin", config.Branch}, "Fetching remote branch"},
			{"git", []string{"checkout", "-b", config.Branch, fmt.Sprintf("origin/%s", config.Branch)}, "Checking out branch"},
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
		if err := restoreChanges(); err != nil {
			return err
		}
	}

	// Commit changes
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

	return nil
}

func executeCommands(commands []struct {
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

func backupAndStashChanges() error {
	// 현재 변경사항이 있는지 확인
	statusCmd := exec.Command("git", "status", "--porcelain")
	output, _ := statusCmd.Output()

	if len(output) > 0 {
		fmt.Println("  • Stashing current changes... ")
		stashCmd := exec.Command("git", "stash", "push", "-u")
		stashCmd.Stdout = os.Stdout
		stashCmd.Stderr = os.Stderr

		if err := stashCmd.Run(); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to stash changes: %v", err)
		}
		fmt.Println("✅ Done")
	}

	return nil
}

func restoreChanges() error {
	// stash list 확인
	listCmd := exec.Command("git", "stash", "list")
	output, _ := listCmd.Output()

	if len(output) > 0 {
		fmt.Println("  • Restoring stashed changes... ")

		// stash apply 사용 (pop 대신)
		applyCmd := exec.Command("git", "stash", "apply")
		applyOutput, err := applyCmd.CombinedOutput()
		if err != nil {
			// 충돌이 발생한 경우
			if strings.Contains(string(applyOutput), "CONFLICT") {
				fmt.Println("⚠️  Conflicts detected, discarding stashed changes")

				// 변경사항 초기화
				resetCmd := exec.Command("git", "reset", "--hard")
				if resetErr := resetCmd.Run(); resetErr != nil {
					fmt.Println("❌ Failed to reset changes")
					return fmt.Errorf("failed to reset after conflict: %v", resetErr)
				}

				// stash 드롭
				dropCmd := exec.Command("git", "stash", "drop")
				if dropErr := dropCmd.Run(); dropErr != nil {
					fmt.Println("⚠️  Failed to drop stash, but continuing...")
				}

				fmt.Println("✅ Cleaned up conflicts")
				return nil
			}

			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to restore changes: %v", err)
		}

		// 성공적으로 적용된 경우 stash 드롭
		dropCmd := exec.Command("git", "stash", "drop")
		if err := dropCmd.Run(); err != nil {
			fmt.Println("⚠️  Failed to drop stash, but continuing...")
		}

		fmt.Println("✅ Done")
	}

	return nil
}
