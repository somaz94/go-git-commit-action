package git

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/somaz94/go-git-commit-action/internal/config"
)

func CreatePullRequest(config *config.GitConfig) error {
	fmt.Println("\n🔄 Creating Pull Request:")

	var sourceBranch string
	if config.AutoBranch {
		sourceBranch = fmt.Sprintf("update-files-%s", time.Now().Format("20060102-150405"))

		fmt.Printf("  • Creating new branch %s... ", sourceBranch)
		if err := exec.Command("git", "checkout", "-b", sourceBranch).Run(); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to create branch: %v", err)
		}
		fmt.Println("✅ Done")

		commitCommands := []struct {
			name string
			args []string
			desc string
		}{
			{"git", []string{"add", config.FilePattern}, "Adding files"},
			{"git", []string{"commit", "-m", config.CommitMessage}, "Committing changes"},
			{"git", []string{"push", "-u", "origin", sourceBranch}, "Pushing changes"},
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
	} else {
		if config.PRBranch == "" {
			return fmt.Errorf("pr_branch must be specified when auto_branch is false")
		}
		sourceBranch = config.PRBranch

		fmt.Printf("\n📊 Changed files between %s and %s:\n", config.PRBase, sourceBranch)
		diffFiles := exec.Command("git", "diff", fmt.Sprintf("origin/%s..origin/%s", config.PRBase, sourceBranch), "--name-status")
		filesOutput, _ := diffFiles.Output()
		if len(filesOutput) == 0 {
			fmt.Println("No changes detected")
			return fmt.Errorf("no changes to create PR")
		}
		fmt.Printf("%s\n", string(filesOutput))
	}

	fmt.Printf("\n✅ Branch '%s' is ready for PR.\n", sourceBranch)
	prURL := fmt.Sprintf("https://github.com/%s/compare/%s...%s",
		os.Getenv("GITHUB_REPOSITORY"),
		config.PRBase,
		sourceBranch)
	fmt.Printf("✅ You can create a pull request by visiting:\n   %s\n", prURL)

	fmt.Printf("  • Creating pull request from %s to %s... ", sourceBranch, config.PRBase)

	runID := os.Getenv("GITHUB_RUN_ID")
	jsonData := fmt.Sprintf(`{
		"title": "Auto PR: %s to %s (Run ID: %s)",
		"head": "%s",
		"base": "%s",
		"body": "Created by Go Git Commit Action\nSource: %s\nTarget: %s\nGitHub Run ID: %s"
	}`, sourceBranch, config.PRBase, runID, sourceBranch, config.PRBase, sourceBranch, config.PRBase, runID)

	curlCmd := exec.Command("curl", "-s", "-X", "POST",
		"-H", fmt.Sprintf("Authorization: Bearer %s", config.GitHubToken),
		"-H", "Accept: application/vnd.github+json",
		"-H", "Content-Type: application/json",
		fmt.Sprintf("https://api.github.com/repos/%s/pulls", os.Getenv("GITHUB_REPOSITORY")),
		"-d", jsonData)

	output, err := curlCmd.CombinedOutput()
	if err != nil {
		fmt.Println("⚠️  Failed to create PR automatically")
		fmt.Printf("Error executing curl: %v\n", err)
		fmt.Printf("Response: %s\n", string(output))
		fmt.Printf("You can create a pull request manually by visiting:\n   %s\n", prURL)
		return err
	}

	if strings.Contains(string(output), "html_url") {
		fmt.Printf("✅ Done\n")
		var response map[string]interface{}
		if err := json.Unmarshal(output, &response); err == nil {
			if htmlURL, ok := response["html_url"].(string); ok {
				fmt.Printf("Pull request created: %s\n", htmlURL)
			}
		}

		if config.DeleteSourceBranch && config.AutoBranch {
			fmt.Printf("\n  • Deleting source branch %s... ", sourceBranch)
			deleteCommand := exec.Command("git", "push", "origin", "--delete", sourceBranch)
			if err := deleteCommand.Run(); err != nil {
				fmt.Println("❌ Failed")
				return fmt.Errorf("failed to delete source branch: %v", err)
			}
			fmt.Println("✅ Done")
		}
	} else {
		fmt.Printf("⚠️  Failed to create PR\n")
		fmt.Printf("You can create a pull request manually by visiting:\n   %s\n", prURL)
		return fmt.Errorf("failed to create PR")
	}

	fmt.Println("\n✨ Git Commit Action Completed Successfully!\n" +
		"=========================================")

	return nil
}
