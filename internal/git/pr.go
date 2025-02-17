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

		// 새 브랜치 생성 전에 이미 존재하는지 확인
		checkBranch := exec.Command("git", "show-ref", "--verify", fmt.Sprintf("refs/heads/%s", sourceBranch))
		if checkBranch.Run() == nil {
			return fmt.Errorf("branch %s already exists", sourceBranch)
		}

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
	} else {
		// PRBranch가 지정되어 있는지 확인
		if config.PRBranch == "" {
			return fmt.Errorf("pr_branch must be specified when auto_branch is false")
		}
		sourceBranch = config.PRBranch

		// PRBase와 PRBranch 간의 변경사항 확인
		fmt.Printf("\n📊 Changed files between %s and %s:\n", config.PRBase, sourceBranch)
		diffFiles := exec.Command("git", "diff", fmt.Sprintf("origin/%s..origin/%s", config.PRBase, sourceBranch), "--name-status")
		filesOutput, _ := diffFiles.Output()
		if len(filesOutput) == 0 {
			fmt.Println("No changes detected")
			return fmt.Errorf("no changes to create PR")
		}
		fmt.Printf("%s\n", string(filesOutput))
	}

	// PR URL 생성 및 출력
	fmt.Printf("\n✅ Branch '%s' is ready for PR.\n", sourceBranch)
	prURL := fmt.Sprintf("https://github.com/%s/compare/%s...%s",
		os.Getenv("GITHUB_REPOSITORY"),
		config.PRBase,
		sourceBranch)
	fmt.Printf("✅ You can create a pull request by visiting:\n   %s\n", prURL)

	// PR 생성
	fmt.Printf("  • Creating pull request from %s to %s... ", sourceBranch, config.PRBase)

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
		title = fmt.Sprintf("Auto PR: %s to %s (Run ID: %s)", sourceBranch, config.PRBase, runID)
	}

	// PR body 설정
	body := config.PRBody
	if body == "" {
		body = fmt.Sprintf("Created by Go Git Commit Action\nSource: %s\nTarget: %s\nCommit: %s\nGitHub Run ID: %s",
			sourceBranch, config.PRBase, commitID, runID)
	}

	// JSON 데이터 준비
	prData := map[string]interface{}{
		"title": title,
		"head":  sourceBranch,
		"base":  config.PRBase,
		"body":  body,
	}

	jsonData, err := json.Marshal(prData)
	if err != nil {
		return fmt.Errorf("failed to marshal PR data: %v", err)
	}

	// GitHub API를 통해 PR 생성
	curlCmd := exec.Command("curl", "-s", "-X", "POST",
		"-H", fmt.Sprintf("Authorization: Bearer %s", config.GitHubToken),
		"-H", "Accept: application/vnd.github+json",
		"-H", "Content-Type: application/json",
		fmt.Sprintf("https://api.github.com/repos/%s/pulls", os.Getenv("GITHUB_REPOSITORY")),
		"-d", string(jsonData))

	output, err := curlCmd.CombinedOutput()
	if err != nil {
		fmt.Println("⚠️  Failed to create PR automatically")
		fmt.Printf("Error executing curl: %v\n", err)
		fmt.Printf("Response: %s\n", string(output))
		fmt.Printf("You can create a pull request manually by visiting:\n   %s\n", prURL)
	} else {
		// API 응답이 성공적인지 확인
		if strings.Contains(string(output), "html_url") {
			fmt.Printf("✅ Done\n")
			// API 응답에서 PR URL과 번호 추출
			var response map[string]interface{}
			if err := json.Unmarshal(output, &response); err == nil {
				if htmlURL, ok := response["html_url"].(string); ok {
					fmt.Printf("Pull request created: %s\n", htmlURL)

					// PR 번호 추출
					if number, ok := response["number"].(float64); ok {
						prNumber := int(number)

						// 라벨 추가
						if config.PRLabels != "" {
							fmt.Printf("  • Adding labels to PR #%d... ", prNumber)
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
								fmt.Sprintf("https://api.github.com/repos/%s/issues/%d/labels",
									os.Getenv("GITHUB_REPOSITORY"), prNumber),
								"-d", string(jsonLabelsData))

							if labelsOutput, err := labelsCurlCmd.CombinedOutput(); err != nil {
								fmt.Println("❌ Failed")
								fmt.Printf("Error adding labels: %v\n", err)
								fmt.Printf("Response: %s\n", string(labelsOutput))
							} else {
								fmt.Println("✅ Done")
							}
						}

						// PR을 자동으로 close해야 하는 경우
						if config.PRClosed {
							fmt.Printf("  • Closing pull request #%d... ", prNumber)

							// PR close를 위한 API 호출
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
								fmt.Printf("Error closing PR: %v\n", err)
								fmt.Printf("Response: %s\n", string(closeOutput))
							} else {
								fmt.Println("✅ Done")
							}
						}
					}
				}
			}
		} else {
			fmt.Printf("⚠️  Failed to create PR\n")
			fmt.Printf("You can create a pull request manually by visiting:\n   %s\n", prURL)
		}
	}

	// PR 생성 후에만 브랜치 삭제
	if config.DeleteSourceBranch && config.AutoBranch && strings.Contains(string(output), "html_url") {
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
