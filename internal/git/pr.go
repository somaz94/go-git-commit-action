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
		// Generate timestamped branch name
		sourceBranch = fmt.Sprintf("update-files-%s", time.Now().Format("20060102-150405"))
		config.PRBranch = sourceBranch

		// Create and switch to new branch
		fmt.Printf("  • Creating new branch %s... ", sourceBranch)
		if err := exec.Command("git", "checkout", "-b", sourceBranch).Run(); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to create branch: %v", err)
		}
		fmt.Println("✅ Done")

		// Commit and push changes
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
	}

	// Check for changes between branches
	fmt.Printf("\n📊 Changed files between %s and %s:\n", config.PRBase, config.PRBranch)

	// Fetch the latest changes
	fetchCmd := exec.Command("git", "fetch", "origin", config.PRBranch)
	if err := fetchCmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch branch: %v", err)
	}

	diffFiles := exec.Command("git", "diff", fmt.Sprintf("origin/%s..origin/%s", config.PRBase, config.PRBranch), "--name-status")
	filesOutput, _ := diffFiles.Output()

	if len(filesOutput) == 0 {
		fmt.Println("No changes detected")
		if config.SkipIfEmpty {
			return nil
		}
		return fmt.Errorf("no changes to create PR")
	}
	fmt.Printf("%s\n", string(filesOutput))

	// PR URL 생성 및 출력
	fmt.Printf("\n✅ Branch '%s' is ready for PR.\n", config.PRBranch)
	prURL := fmt.Sprintf("https://github.com/%s/compare/%s...%s",
		os.Getenv("GITHUB_REPOSITORY"),
		config.PRBase,
		config.PRBranch)
	fmt.Printf("✅ You can create a pull request by visiting:\n   %s\n", prURL)

	// Create PR
	fmt.Printf("  • Creating pull request from %s to %s... ", config.PRBranch, config.PRBase)

	// GitHub Run ID 가져오기
	runID := os.Getenv("GITHUB_RUN_ID")

	// Get current commit SHA
	commitCmd := exec.Command("git", "rev-parse", "HEAD")
	commitSHA, err := commitCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get commit SHA: %v", err)
	}
	commitID := strings.TrimSpace(string(commitSHA))

	// PR 제목 설정
	title := config.PRTitle
	if title == "" {
		title = fmt.Sprintf("Auto PR: %s to %s (Run ID: %s)", config.PRBranch, config.PRBase, runID)
	}

	// PR body 설정
	body := config.PRBody
	if body == "" {
		body = fmt.Sprintf("Created by Go Git Commit Action\nSource: %s\nTarget: %s\nCommit: %s\nGitHub Run ID: %s",
			config.PRBranch, config.PRBase, commitID, runID)
	}

	// Create PR request
	prData := map[string]interface{}{
		"title": title,
		"head":  config.PRBranch,
		"base":  config.PRBase,
		"body":  body,
	}

	jsonData, err := json.Marshal(prData)
	if err != nil {
		return fmt.Errorf("failed to marshal PR data: %v", err)
	}

	// Create PR using GitHub API
	curlCmd := exec.Command("curl", "-s", "-X", "POST",
		"-H", fmt.Sprintf("Authorization: Bearer %s", config.GitHubToken),
		"-H", "Accept: application/vnd.github+json",
		"-H", "Content-Type: application/json",
		fmt.Sprintf("https://api.github.com/repos/%s/pulls", os.Getenv("GITHUB_REPOSITORY")),
		"-d", string(jsonData))

	output, err := curlCmd.CombinedOutput()
	if err != nil {
		fmt.Println("⚠️  Failed to create PR automatically")
		fmt.Printf("Error: %v\n", err)
		fmt.Printf("Response: %s\n", string(output))
		return fmt.Errorf("failed to create PR: %v - %s", err, string(output))
	}

	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(output, &response); err != nil {
		fmt.Printf("Failed to parse response: %s\n", string(output))
		return fmt.Errorf("failed to parse PR response: %v", err)
	}

	// PR이 성공적으로 생성되었는지 확인
	if response["message"] != nil {
		fmt.Printf("GitHub API Error: %s\n", response["message"])
		return fmt.Errorf("GitHub API Error: %s", response["message"])
	}

	// PR 번호 추출
	prNumber, ok := response["number"].(float64)
	if !ok {
		return fmt.Errorf("failed to get PR number from response")
	}

	fmt.Println("✅ Done")
	fmt.Printf("Pull request created: %s\n", response["html_url"].(string))

	// 라벨 추가
	if config.PRLabels != "" {
		fmt.Printf("  • Adding labels to PR #%d... ", int(prNumber))
		labels := strings.Split(config.PRLabels, ",")
		for i := range labels {
			labels[i] = strings.TrimSpace(labels[i])
		}

		labelsData := map[string]interface{}{
			"labels": labels,
		}
		jsonLabelsData, _ := json.Marshal(labelsData)

		labelsCurlCmd := exec.Command("curl", "-s", "-X", "POST",
			"-H", fmt.Sprintf("Authorization: Bearer %s", config.GitHubToken),
			"-H", "Accept: application/vnd.github+json",
			"-H", "Content-Type: application/json",
			"-H", "X-GitHub-Api-Version: 2022-11-28",
			fmt.Sprintf("https://api.github.com/repos/%s/issues/%d/labels",
				os.Getenv("GITHUB_REPOSITORY"), int(prNumber)),
			"-d", string(jsonLabelsData))

		labelsOutput, err := labelsCurlCmd.CombinedOutput()
		if err != nil {
			fmt.Println("❌ Failed")
			fmt.Printf("Error adding labels: %v\n", err)
			fmt.Printf("Response: %s\n", string(labelsOutput))
		} else {
			fmt.Println("✅ Done")
		}
	}

	// PR 닫기
	if config.PRClosed {
		fmt.Printf("  • Closing pull request #%d... ", int(prNumber))
		closeData := map[string]string{
			"state": "closed",
		}
		jsonCloseData, _ := json.Marshal(closeData)

		closeCurlCmd := exec.Command("curl", "-s", "-X", "PATCH",
			"-H", fmt.Sprintf("Authorization: Bearer %s", config.GitHubToken),
			"-H", "Accept: application/vnd.github+json",
			"-H", "Content-Type: application/json",
			"-H", "X-GitHub-Api-Version: 2022-11-28",
			fmt.Sprintf("https://api.github.com/repos/%s/pulls/%d",
				os.Getenv("GITHUB_REPOSITORY"), int(prNumber)),
			"-d", string(jsonCloseData))

		closeOutput, err := closeCurlCmd.CombinedOutput()
		if err != nil {
			fmt.Println("❌ Failed")
			fmt.Printf("Error closing PR: %v\n", err)
			fmt.Printf("Response: %s\n", string(closeOutput))
		} else {
			fmt.Println("✅ Done")
		}
	}

	// Delete source branch if requested
	if config.DeleteSourceBranch && config.AutoBranch {
		fmt.Printf("\n  • Deleting source branch %s... ", sourceBranch)
		deleteCommand := exec.Command("git", "push", "origin", "--delete", sourceBranch)
		if err := deleteCommand.Run(); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to delete source branch: %v", err)
		}
		fmt.Println("✅ Done")
	}

	fmt.Println("\n✨ Git Commit Action Completed Successfully!\n" +
		"=========================================")

	return nil
}
