package git

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/somaz94/go-git-commit-action/internal/config"
)

func CreatePullRequest(config *config.GitConfig) error {
	fmt.Println("\n🔄 Creating Pull Request:")

	// PRBase와 현재 브랜치(Branch)의 차이점 확인 - 파일 목록만
	fmt.Printf("\n📊 Changed files between %s and %s:\n", config.PRBase, config.Branch)
	diffFiles := exec.Command("git", "diff", fmt.Sprintf("origin/%s..origin/%s", config.PRBase, config.Branch), "--name-status")
	filesOutput, _ := diffFiles.Output()
	if len(filesOutput) > 0 {
		fmt.Printf("%s\n", string(filesOutput))
	} else {
		fmt.Println("No changes detected")
		return fmt.Errorf("no changes to create PR")
	}

	var sourceBranch string
	if config.AutoBranch {
		// 자동 브랜치 생성
		sourceBranch = fmt.Sprintf("update-files-%s", time.Now().Format("20060102-150405"))

		// 브랜치 생성 및 변경사항 적용
		fmt.Printf("  • Fetching latest changes... ")
		if err := exec.Command("git", "fetch", "origin", config.Branch).Run(); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to fetch branch: %v", err)
		}
		fmt.Println("✅ Done")

		// config.Branch 브랜치로 체크아웃
		fmt.Printf("  • Checking out %s branch... ", config.Branch)
		if err := exec.Command("git", "checkout", config.Branch).Run(); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to checkout branch: %v", err)
		}
		fmt.Println("✅ Done")

		// config.Branch 브랜치의 최신 상태로 업데이트
		fmt.Printf("  • Updating to latest state... ")
		if err := exec.Command("git", "pull", "origin", config.Branch).Run(); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to pull latest changes: %v", err)
		}
		fmt.Println("✅ Done")

		// 새 브랜치 생성
		fmt.Printf("  • Creating new branch %s... ", sourceBranch)
		if err := exec.Command("git", "checkout", "-b", sourceBranch).Run(); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to create new branch: %v", err)
		}
		fmt.Println("✅ Done")

		// 새 브랜치 푸시
		fmt.Printf("  • Pushing new branch with changes... ")
		if err := exec.Command("git", "push", "-u", "origin", sourceBranch).Run(); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to push branch: %v", err)
		}
		fmt.Println("✅ Done")
	} else {
		// PRBranch가 지정되어 있는지 확인
		if config.PRBranch == "" {
			return fmt.Errorf("pr_branch must be specified when auto_branch is false")
		}
		sourceBranch = config.PRBranch
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

	prTitle := fmt.Sprintf("Auto PR: %s to %s", sourceBranch, config.PRBase)
	prBody := fmt.Sprintf("Created by Go Git Commit Action\nSource: %s\nTarget: %s", sourceBranch, config.PRBase)

	// GitHub API를 통해 PR 생성
	curlCmd := exec.Command("curl", "-s", "-X", "POST",
		"-H", fmt.Sprintf("Authorization: token %s", os.Getenv("GITHUB_TOKEN")),
		"-H", "Accept: application/vnd.github+json",
		fmt.Sprintf("https://api.github.com/repos/%s/pulls", os.Getenv("GITHUB_REPOSITORY")),
		"-d", fmt.Sprintf(`{"title":"%s", "head":"%s", "base":"%s", "body":"%s"}`,
			prTitle, sourceBranch, config.PRBase, prBody))

	if output, err := curlCmd.CombinedOutput(); err != nil {
		fmt.Println("⚠️  Failed to create PR automatically")
		fmt.Printf("Error: %s\n", string(output))
		fmt.Printf("You can create a pull request manually by visiting:\n   %s\n", prURL)
	} else {
		fmt.Printf("✅ Done\n")
		fmt.Printf("PR created successfully\n")
	}

	// 소스 브랜치 삭제 (옵션이 활성화된 경우와 auto_branch가 true인 경우에만)
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
