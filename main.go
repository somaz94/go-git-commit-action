package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

type GitConfig struct {
	UserEmail     string
	UserName      string
	CommitMessage string
	Branch        string
	RepoPath      string
	FilePattern   string
}

func main() {
	config := GitConfig{
		UserEmail:     os.Getenv("INPUT_USER_EMAIL"),
		UserName:      os.Getenv("INPUT_USER_NAME"),
		CommitMessage: getEnvWithDefault("INPUT_COMMIT_MESSAGE", "Auto commit by Go Git Commit Action"),
		Branch:        getEnvWithDefault("INPUT_BRANCH", "main"),
		RepoPath:      getEnvWithDefault("INPUT_REPOSITORY_PATH", "."),
		FilePattern:   getEnvWithDefault("INPUT_FILE_PATTERN", "."),
	}

	if err := runGitCommit(config); err != nil {
		log.Fatalf("Error executing git commands: %v", err)
	}
}

// 기본값을 처리하는 헬퍼 함수
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func runGitCommit(config GitConfig) error {
	// Debug information
	currentDir, _ := os.Getwd()
	fmt.Printf("Current directory: %s\n", currentDir)
	fmt.Printf("User Email: %s\n", config.UserEmail)
	fmt.Printf("User Name: %s\n", config.UserName)
	fmt.Printf("Commit message: %s\n", config.CommitMessage)
	fmt.Printf("Branch: %s\n", config.Branch)
	fmt.Printf("Repository path: %s\n", config.RepoPath)
	fmt.Printf("File pattern: %s\n", config.FilePattern)

	// List directory contents
	files, _ := os.ReadDir(".")
	fmt.Println("Contents of current directory:")
	for _, file := range files {
		fmt.Printf("- %s\n", file.Name())
	}

	// Change to repository directory if specified
	if config.RepoPath != "." {
		if err := os.Chdir(config.RepoPath); err != nil {
			return fmt.Errorf("failed to change directory: %v", err)
		}

		// Print new directory after change
		newDir, _ := os.Getwd()
		fmt.Printf("New directory after change: %s\n", newDir)
	}

	// Set git configurations first
	commands := []struct {
		name string
		args []string
	}{
		{"git", []string{"config", "--global", "--add", "safe.directory", "/app"}},
		{"git", []string{"config", "--global", "--add", "safe.directory", "/github/workspace"}},
		{"git", []string{"config", "--global", "user.email", config.UserEmail}},
		{"git", []string{"config", "--global", "user.name", config.UserName}},
		// git config 설정 후 확인
		{"git", []string{"config", "--global", "--list"}},
		{"git", []string{"add", config.FilePattern}},
		{"git", []string{"commit", "-m", config.CommitMessage}},
		{"git", []string{"push", "origin", config.Branch}},
	}

	// Execute git commands
	for _, cmd := range commands {
		command := exec.Command(cmd.name, cmd.args...)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr

		if err := command.Run(); err != nil {
			// Skip error if nothing to commit
			if cmd.args[0] == "commit" && err.Error() == "exit status 1" {
				fmt.Println("Nothing to commit, skipping...")
				continue
			}
			return fmt.Errorf("failed to execute %s: %v", cmd.name, err)
		}
	}

	return nil
}
