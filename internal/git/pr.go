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
	} else {
		// PRBranch가 지정되어 있는지 확인
		if config.PRBranch == "" {
			return fmt.Errorf("pr_branch must be specified when auto_branch is false")
		}
		sourceBranch = config.PRBranch
	}

	// 브랜치 생성 및 변경사항 적용
	fmt.Printf("  • Fetching latest changes... ")
	if err := exec.Command("git", "fetch", "origin", config.Branch).Run(); err != nil {
		fmt.Println("❌ Failed")
		return fmt.Errorf("failed to fetch branch: %v", err)
	}
	fmt.Println("✅ Done")

	// 새 브랜치 생성 (origin/test의 상태에서 시작)
	fmt.Printf("  • Creating new branch %s from origin/%s... ", sourceBranch, config.Branch)
	if err := exec.Command("git", "checkout", "-b", sourceBranch, fmt.Sprintf("origin/%s", config.Branch)).Run(); err != nil {
		fmt.Println("❌ Failed")
		return fmt.Errorf("failed to create new branch: %v", err)
	}
	fmt.Println("✅ Done")

	// test 브랜치의 변경사항을 새 브랜치에 적용
	fmt.Printf("  • Applying changes from test branch... ")
	if err := exec.Command("git", "cherry-pick", fmt.Sprintf("origin/%s..origin/%s", config.PRBase, config.Branch)).Run(); err != nil {
		fmt.Println("❌ Failed")
		return fmt.Errorf("failed to apply changes: %v", err)
	}
	fmt.Println("✅ Done")

	// 새 브랜치 푸시
	fmt.Printf("  • Pushing new branch with changes... ")
	if err := exec.Command("git", "push", "-u", "origin", sourceBranch).Run(); err != nil {
		fmt.Println("❌ Failed")
		return fmt.Errorf("failed to push branch: %v", err)
	}
	fmt.Println("✅ Done")

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
