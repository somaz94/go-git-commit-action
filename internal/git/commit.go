package git

import (
	"fmt"
	"os"
	"os/exec"

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

	// 기본 git 설정 실행
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

	// 브랜치 존재 여부 확인
	checkLocalBranch := exec.Command("git", "rev-parse", "--verify", config.Branch)
	checkRemoteBranch := exec.Command("git", "ls-remote", "--heads", "origin", config.Branch)

	if checkLocalBranch.Run() != nil && checkRemoteBranch.Run() != nil {
		// 로컬과 리모트 모두에 브랜치가 없는 경우에만 새로 생성
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
		// 리모트에는 있지만 로컬에는 없는 경우
		fmt.Printf("\n⚠️  Checking out existing remote branch '%s'...\n", config.Branch)

		// 현재 변경사항을 임시 저장
		fmt.Printf("  • Saving current changes... ")
		stashCommand := exec.Command("git", "stash", "save", "temporary")
		stashCommand.Stdout = os.Stdout
		stashCommand.Stderr = os.Stderr
		if err := stashCommand.Run(); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to stash changes: %v", err)
		}
		fmt.Println("✅ Done")

		// 리모트 브랜치 체크아웃
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

		// 임시 저장한 변경사항 적용
		fmt.Printf("  • Applying saved changes... ")
		popCommand := exec.Command("git", "stash", "pop")
		popCommand.Stdout = os.Stdout
		popCommand.Stderr = os.Stderr
		if err := popCommand.Run(); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to apply stashed changes: %v", err)
		}
		fmt.Println("✅ Done")
	}

	// 변경사항 커밋 및 푸시
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
