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
		// 자동 브랜치 생성
		sourceBranch = fmt.Sprintf("update-files-%s", time.Now().Format("20060102-150405"))

		// 브랜치 생성 및 변경사항 적용
		fmt.Printf("  • Fetching latest changes... ")
		if err := exec.Command("git", "fetch", "--all").Run(); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to fetch branch: %v", err)
		}
		fmt.Println("✅ Done")

		// test 브랜치 체크아웃
		fmt.Printf("  • Checking out source branch %s... ", config.Branch)
		if err := exec.Command("git", "checkout", config.Branch).Run(); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to checkout source branch: %v", err)
		}
		fmt.Println("✅ Done")

		// test 브랜치 최신화
		fmt.Printf("  • Pulling latest changes... ")
		if err := exec.Command("git", "pull", "origin", config.Branch).Run(); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to pull latest changes: %v", err)
		}
		fmt.Println("✅ Done")

		// 새 브랜치 생성
		fmt.Printf("  • Creating new branch %s... ", sourceBranch)
		if err := exec.Command("git", "checkout", "-b", sourceBranch).Run(); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to create branch: %v", err)
		}
		fmt.Println("✅ Done")

		// 새 브랜치 푸시
		fmt.Printf("  • Pushing new branch... ")
		if err := exec.Command("git", "push", "-u", "origin", sourceBranch).Run(); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to push branch: %v", err)
		}
		fmt.Println("✅ Done")

		// 잠시 대기 (원격 저장소 반영 대기)
		time.Sleep(2 * time.Second)

		// 새 브랜치와 PRBase 간의 변경사항 확인
		fmt.Printf("\n📊 Changed files between %s and %s:\n", config.PRBase, sourceBranch)
		diffFiles := exec.Command("git", "diff", fmt.Sprintf("origin/%s..origin/%s", config.PRBase, sourceBranch), "--name-status")
		filesOutput, _ := diffFiles.Output()
		if len(filesOutput) == 0 {
			// 변경사항이 없으면 브랜치 삭제하고 종료
			if config.DeleteSourceBranch {
				fmt.Printf("\n  • Deleting source branch %s... ", sourceBranch)
				deleteCommand := exec.Command("git", "push", "origin", "--delete", sourceBranch)
				deleteCommand.Run()
				fmt.Println("✅ Done")
			}
			fmt.Println("No changes detected")
			return fmt.Errorf("no changes to create PR")
		}
		fmt.Printf("%s\n", string(filesOutput))

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

	// JSON 데이터 준비
	jsonData := fmt.Sprintf(`{
		"title": "Auto PR: %s to %s (Run ID: %s)",
		"head": "%s",
		"base": "%s",
		"body": "Created by Go Git Commit Action\nSource: %s\nTarget: %s\nGitHub Run ID: %s"
	}`, sourceBranch, config.PRBase, runID, sourceBranch, config.PRBase, sourceBranch, config.PRBase, runID)

	// GitHub API를 통해 PR 생성
	curlCmd := exec.Command("curl", "-s", "-X", "POST",
		"-H", fmt.Sprintf("Authorization: Bearer %s", os.Getenv("GITHUB_TOKEN")),
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
	} else {
		// API 응답이 성공적인지 확인
		if strings.Contains(string(output), "html_url") {
			fmt.Printf("✅ Done\n")
			// API 응답에서 PR URL 추출
			var response map[string]interface{}
			if err := json.Unmarshal(output, &response); err == nil {
				if htmlURL, ok := response["html_url"].(string); ok {
					fmt.Printf("Pull request created: %s\n", htmlURL)
				}
			}
		} else {
			fmt.Printf("⚠️  Failed to create PR\n")
			fmt.Printf("You can create a pull request manually by visiting:\n   %s\n", prURL)
		}
	}

	// 소스 브랜치 삭제는 PR 생성 성공 후에만 수행
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
