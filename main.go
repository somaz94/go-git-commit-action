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
		RepoPath:      os.Getenv("INPUT_REPOSITORY_PATH"),
		FilePattern:   os.Getenv("INPUT_FILE_PATTERN"),
	}

	if err := runGitCommit(config); err != nil {
		log.Fatalf("Error executing git commands: %v", err)
	}
}

func runGitCommit(config GitConfig) error {
	// Change to repository directory if specified
	if config.RepoPath != "." {
		if err := os.Chdir(config.RepoPath); err != nil {
			return fmt.Errorf("failed to change directory: %v", err)
		}
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
