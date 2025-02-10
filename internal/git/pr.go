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

		// 현재 브랜치의 변경사항을 새 브랜치로 복사
		fmt.Printf("  • Creating new branch %s from %s... ", sourceBranch, config.Branch)
		// 먼저 현재 브랜치의 최신 상태를 가져옴
		fetchCmd := exec.Command("git", "fetch", "origin", config.Branch)
		if err := fetchCmd.Run(); err != nil {
			fmt.Println("❌ Failed to fetch")
			return fmt.Errorf("failed to fetch branch: %v", err)
		}

		// 새 브랜치 생성
		createBranch := exec.Command("git", "branch", sourceBranch, fmt.Sprintf("origin/%s", config.Branch))
		if err := createBranch.Run(); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to create branch: %v", err)
		}

		// 새 브랜치로 체크아웃
		checkoutCmd := exec.Command("git", "checkout", sourceBranch)
		if err := checkoutCmd.Run(); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to checkout branch: %v", err)
		}
		fmt.Println("✅ Done")

		// 새 브랜치 푸시
		fmt.Printf("  • Pushing new branch with changes... ")
		pushCmd := exec.Command("git", "push", "-u", "origin", sourceBranch)
		if err := pushCmd.Run(); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to push branch: %v", err)
		}
		fmt.Println("✅ Done")
	} else {
		sourceBranch = config.Branch
	}

	// PR URL 생성 및 출력
	fmt.Printf("\n✅ Branch '%s' has been created and pushed.\n", sourceBranch)
	prURL := fmt.Sprintf("https://github.com/%s/compare/%s...%s",
		os.Getenv("GITHUB_REPOSITORY"),
		config.PRBase,
		sourceBranch)
	fmt.Printf("✅ You can create a pull request by visiting:\n   %s\n", prURL)

	// GitHub CLI로 PR 생성
	fmt.Printf("  • Creating pull request from %s to %s... ", sourceBranch, config.PRBase)
	prCmd := exec.Command("gh", "pr", "create",
		"--title", config.PRTitle,
		"--body", fmt.Sprintf("Created by Go Git Commit Action\nSource: %s\nTarget: %s", sourceBranch, config.PRBase),
		"--base", config.PRBase,
		"--head", sourceBranch)

	if err := prCmd.Run(); err != nil {
		fmt.Println("⚠️  Manual PR creation required")
	} else {
		fmt.Println("✅ Done")
	}

	// 소스 브랜치 삭제 (옵션이 활성화된 경우)
	if config.DeleteSourceBranch {
		fmt.Printf("\n  • Deleting source branch %s... ", sourceBranch)
		deleteCommand := exec.Command("git", "push", "origin", "--delete", sourceBranch)
		if err := deleteCommand.Run(); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to delete source branch: %v", err)
		}
		fmt.Println("✅ Done")
	}

	return nil
}
