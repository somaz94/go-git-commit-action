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
		CommitMessage: os.Getenv("INPUT_COMMIT_MESSAGE"),
		Branch:        os.Getenv("INPUT_BRANCH"),
		RepoPath:      getEnvWithDefault("INPUT_REPOSITORY_PATH", "."),
		FilePattern:   os.Getenv("INPUT_FILE_PATTERN"),
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
	// Confirm Git Config
	gitConfig := exec.Command("git", "config", "--global", "user.email")
	gitConfig.Stdout = os.Stdout
	gitConfig.Stderr = os.Stderr
	if err := gitConfig.Run(); err != nil {
		return fmt.Errorf("failed to execute git config: %v", err)
	}
	fmt.Printf("Commit message: %s\n", config.CommitMessage)
	fmt.Printf("Branch: %s\n", config.Branch)
	fmt.Printf("Current directory: %s\n", currentDir)
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

	// Set git configurations
	commands := []struct {
		name string
		args []string
	}{
		{"git", []string{"config", "--global", "user.email", config.UserEmail}},
		{"git", []string{"config", "--global", "user.name", config.UserName}},
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
