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
	TagName       string
	TagMessage    string
	DeleteTag     bool
	TagReference  string
}

func main() {
	config := GitConfig{
		UserEmail:     os.Getenv("INPUT_USER_EMAIL"),
		UserName:      os.Getenv("INPUT_USER_NAME"),
		CommitMessage: getEnvWithDefault("INPUT_COMMIT_MESSAGE", "Auto commit by Go Git Commit Action"),
		Branch:        getEnvWithDefault("INPUT_BRANCH", "main"),
		RepoPath:      getEnvWithDefault("INPUT_REPOSITORY_PATH", "."),
		FilePattern:   getEnvWithDefault("INPUT_FILE_PATTERN", "."),
		TagName:       os.Getenv("INPUT_TAG_NAME"),
		TagMessage:    os.Getenv("INPUT_TAG_MESSAGE"),
		DeleteTag:     os.Getenv("INPUT_DELETE_TAG") == "true",
		TagReference:  os.Getenv("INPUT_TAG_REFERENCE"),
	}

	if err := runGitCommit(config); err != nil {
		log.Fatalf("Error executing git commands: %v", err)
	}

	if config.TagName != "" {
		if err := handleGitTag(config); err != nil {
			log.Fatalf("Error handling git tag: %v", err)
		}
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
	commands := []struct {
		name string
		args []string
		desc string
	}{
		{"git", []string{"config", "--global", "--add", "safe.directory", "/app"}, "Setting safe directory (/app)"},
		{"git", []string{"config", "--global", "--add", "safe.directory", "/github/workspace"}, "Setting safe directory (/github/workspace)"},
		{"git", []string{"config", "--global", "user.email", config.UserEmail}, "Configuring user email"},
		{"git", []string{"config", "--global", "user.name", config.UserName}, "Configuring user name"},
		{"git", []string{"config", "--global", "--list"}, "Checking git configuration"},
		{"git", []string{"add", config.FilePattern}, "Adding files"},
		{"git", []string{"commit", "-m", config.CommitMessage}, "Committing changes"},
		{"git", []string{"push", "origin", config.Branch}, "Pushing to remote"},
	}

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

	fmt.Println("\n✨ Git Commit Action Completed Successfully!\n" +
		"=========================================")
	return nil
}

func handleGitTag(config GitConfig) error {
	fmt.Println("\n🏷️  Handling Git Tag:")

	if config.DeleteTag {
		// Delete tag
		commands := []struct {
			name string
			args []string
			desc string
		}{
			{"git", []string{"tag", "-d", config.TagName}, "Deleting local tag"},
			{"git", []string{"push", "origin", ":refs/tags/" + config.TagName}, "Deleting remote tag"},
		}

		for _, cmd := range commands {
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
	} else {
		// Create tag
		var tagArgs []string
		if config.TagMessage != "" {
			if config.TagReference != "" {
				tagArgs = []string{"tag", "-f", "-a", config.TagName, config.TagReference, "-m", config.TagMessage}
			} else {
				tagArgs = []string{"tag", "-f", "-a", config.TagName, "-m", config.TagMessage}
			}
		} else {
			if config.TagReference != "" {
				tagArgs = []string{"tag", "-f", config.TagName, config.TagReference}
			} else {
				tagArgs = []string{"tag", "-f", config.TagName}
			}
		}

		// 설명 메시지 생성
		desc := "Creating local tag " + config.TagName
		if config.TagReference != "" {
			desc += fmt.Sprintf(" pointing to %s", config.TagReference)
		}

		commands := []struct {
			name string
			args []string
			desc string
		}{
			{"git", tagArgs, desc},
			{"git", []string{"push", "origin", config.TagName}, "Pushing tag to remote"},
		}

		for _, cmd := range commands {
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
	}

	return nil
}
