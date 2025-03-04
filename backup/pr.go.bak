package git

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/somaz94/go-git-commit-action/internal/config"
)

// 추가: GitHub API 요청을 위한 구조체
type GitHubClient struct {
	token      string
	baseURL    string
	repository string
}

func NewGitHubClient(token, repository string) *GitHubClient {
	return &GitHubClient{
		token:      token,
		baseURL:    "https://api.github.com",
		repository: repository,
	}
}

// GitHub API 요청 메서드 추가
func (c *GitHubClient) CreatePullRequest(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	// API 요청 로직
	return nil, nil
}

func (c *GitHubClient) AddLabels(ctx context.Context, prNumber int, labels []string) error {
	// 라벨 추가 로직
	return nil
}

func (c *GitHubClient) ClosePullRequest(ctx context.Context, prNumber int) error {
	// PR 닫기 로직
	return nil
}

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

		// Commit and push changes to the new branch
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
		// auto_branch=false일 때는 pr_branch로 체크아웃
		sourceBranch = config.PRBranch
		fmt.Printf("  • Checking out branch %s... ", sourceBranch)
		if err := exec.Command("git", "checkout", sourceBranch).Run(); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to checkout branch: %v", err)
		}
		fmt.Println("✅ Done")
	}

	// Check for changes between pr_base and pr_branch
	fmt.Printf("\n📊 Changed files between %s and %s:\n", config.PRBase, config.PRBranch)

	// Fetch both branches
	fetchBaseCmd := exec.Command("git", "fetch", "origin", config.PRBase)
	if err := fetchBaseCmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch base branch: %v", err)
	}

	fetchBranchCmd := exec.Command("git", "fetch", "origin", config.PRBranch)
	if err := fetchBranchCmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch source branch: %v", err)
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

	// Create PR URL and print it
	fmt.Printf("\n✅ Branch '%s' is ready for PR.\n", config.PRBranch)
	prURL := fmt.Sprintf("https://github.com/%s/compare/%s...%s",
		os.Getenv("GITHUB_REPOSITORY"),
		config.PRBase,
		config.PRBranch)
	fmt.Printf("✅ You can create a pull request by visiting:\n   %s\n", prURL)

	// Create PR
	fmt.Printf("  • Creating pull request from %s to %s... ", config.PRBranch, config.PRBase)

	// Get GitHub Run ID
	runID := os.Getenv("GITHUB_RUN_ID")

	// Get current commit SHA
	commitCmd := exec.Command("git", "rev-parse", "HEAD")
	commitSHA, err := commitCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get commit SHA: %v", err)
	}
	commitID := strings.TrimSpace(string(commitSHA))

	// Set PR title
	title := config.PRTitle
	if title == "" {
		title = fmt.Sprintf("Auto PR: %s to %s (Run ID: %s)", config.PRBranch, config.PRBase, runID)
	}

	// Set PR body
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
		return fmt.Errorf("failed to execute curl command: %v", err)
	}

	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(output, &response); err != nil {
		fmt.Printf("Raw response: %s\n", string(output))
		return fmt.Errorf("failed to parse PR response: %v", err)
	}

	// Check for error in response
	if errMsg, ok := response["message"].(string); ok {
		fmt.Printf("GitHub API Error: %s\n", errMsg)
		if errors, ok := response["errors"].([]interface{}); ok {
			fmt.Println("Error details:")
			for _, err := range errors {
				if errMap, ok := err.(map[string]interface{}); ok {
					fmt.Printf("  • %v\n", errMap)
					// PR이 이미 존재하는 경우
					if errMap["message"].(string) == "A pull request already exists for somaz94:test." {
						fmt.Println("⚠️  Pull request already exists")
						// 기존 PR 찾기
						searchCmd := exec.Command("curl", "-s",
							"-H", fmt.Sprintf("Authorization: Bearer %s", config.GitHubToken),
							"-H", "Accept: application/vnd.github+json",
							"-H", "X-GitHub-Api-Version: 2022-11-28",
							fmt.Sprintf("https://api.github.com/repos/%s/pulls?head=%s&base=%s",
								os.Getenv("GITHUB_REPOSITORY"),
								config.PRBranch,
								config.PRBase))

						searchOutput, _ := searchCmd.CombinedOutput()
						var prs []map[string]interface{}
						if err := json.Unmarshal(searchOutput, &prs); err == nil && len(prs) > 0 {
							if number, ok := prs[0]["number"].(float64); ok {
								prNumber := int(number)
								fmt.Printf("Found existing PR #%d\n", prNumber)

								// 라벨 추가
								if len(config.PRLabels) > 0 {
									fmt.Printf("  • Adding labels to PR #%d... ", prNumber)
									labelsData := map[string]interface{}{
										"labels": config.PRLabels,
									}
									jsonLabelsData, _ := json.Marshal(labelsData)

									labelsCurlCmd := exec.Command("curl", "-s", "-X", "POST",
										"-H", fmt.Sprintf("Authorization: Bearer %s", config.GitHubToken),
										"-H", "Accept: application/vnd.github+json",
										"-H", "Content-Type: application/json",
										fmt.Sprintf("https://api.github.com/repos/%s/issues/%d/labels",
											os.Getenv("GITHUB_REPOSITORY"), prNumber),
										"-d", string(jsonLabelsData))

									if labelsOutput, err := labelsCurlCmd.CombinedOutput(); err != nil {
										fmt.Println("❌ Failed")
										fmt.Printf("Error: %v\n", err)
										fmt.Printf("Response: %s\n", string(labelsOutput))
									} else {
										fmt.Println("✅ Done")
									}
								}

								// PR 닫기
								if config.PRClosed {
									fmt.Printf("  • Closing pull request #%d... ", prNumber)
									closeData := map[string]string{
										"state": "closed",
									}
									jsonCloseData, _ := json.Marshal(closeData)

									closeCurlCmd := exec.Command("curl", "-s", "-X", "PATCH",
										"-H", fmt.Sprintf("Authorization: Bearer %s", config.GitHubToken),
										"-H", "Accept: application/vnd.github+json",
										"-H", "Content-Type: application/json",
										fmt.Sprintf("https://api.github.com/repos/%s/pulls/%d",
											os.Getenv("GITHUB_REPOSITORY"), prNumber),
										"-d", string(jsonCloseData))

									if closeOutput, err := closeCurlCmd.CombinedOutput(); err != nil {
										fmt.Println("❌ Failed")
										fmt.Printf("Error: %v\n", err)
										fmt.Printf("Response: %s\n", string(closeOutput))
									} else {
										fmt.Println("✅ Done")
									}
								}

								return nil
							}
						}
					}
				}
			}
		}
		return fmt.Errorf("GitHub API error: %s", errMsg)
	}

	if htmlURL, ok := response["html_url"].(string); ok {
		fmt.Println("✅ Done")
		fmt.Printf("Pull request created: %s\n", htmlURL)

		// Handle PR number
		if number, ok := response["number"].(float64); ok {
			prNumber := int(number)

			// Add labels if specified
			if len(config.PRLabels) > 0 {
				fmt.Printf("  • Adding labels to PR #%d... ", prNumber)
				labelsData := map[string]interface{}{
					"labels": config.PRLabels,
				}
				jsonLabelsData, _ := json.Marshal(labelsData)

				labelsCurlCmd := exec.Command("curl", "-s", "-X", "POST",
					"-H", fmt.Sprintf("Authorization: Bearer %s", config.GitHubToken),
					"-H", "Accept: application/vnd.github+json",
					"-H", "Content-Type: application/json",
					fmt.Sprintf("https://api.github.com/repos/%s/issues/%d/labels",
						os.Getenv("GITHUB_REPOSITORY"), prNumber),
					"-d", string(jsonLabelsData))

				if labelsOutput, err := labelsCurlCmd.CombinedOutput(); err != nil {
					fmt.Println("❌ Failed")
					fmt.Printf("Error: %v\n", err)
					fmt.Printf("Response: %s\n", string(labelsOutput))
				} else {
					fmt.Println("✅ Done")
				}
			}

			// Close PR if requested
			if config.PRClosed {
				fmt.Printf("  • Closing pull request #%d... ", prNumber)
				closeData := map[string]string{
					"state": "closed",
				}
				jsonCloseData, _ := json.Marshal(closeData)

				closeCurlCmd := exec.Command("curl", "-s", "-X", "PATCH",
					"-H", fmt.Sprintf("Authorization: Bearer %s", config.GitHubToken),
					"-H", "Accept: application/vnd.github+json",
					"-H", "Content-Type: application/json",
					fmt.Sprintf("https://api.github.com/repos/%s/pulls/%d",
						os.Getenv("GITHUB_REPOSITORY"), prNumber),
					"-d", string(jsonCloseData))

				if closeOutput, err := closeCurlCmd.CombinedOutput(); err != nil {
					fmt.Println("❌ Failed")
					fmt.Printf("Error: %v\n", err)
					fmt.Printf("Response: %s\n", string(closeOutput))
				} else {
					fmt.Println("✅ Done")
				}
			}
		}
	} else {
		fmt.Println("⚠️  Failed to create PR")
		fmt.Printf("Response: %s\n", string(output))
		return fmt.Errorf("failed to get PR URL from response")
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
