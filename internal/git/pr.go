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
	}

	// 현재 변경사항 확인
	statusCommand := exec.Command("git", "status", "--porcelain")
	statusOutput, _ := statusCommand.Output()
	fmt.Printf("\n📝 Current working tree status:\n%s\n", string(statusOutput))

	var sourceBranch string
	if config.AutoBranch {
		// 자동 브랜치 생성
		sourceBranch = fmt.Sprintf("update-files-%s", time.Now().Format("20060102-150405"))

		// 새 브랜치 생성 시 현재 브랜치의 커밋 내용을 포함
		fmt.Printf("  • Creating new branch %s from %s... ", sourceBranch, config.Branch)
		createBranch := exec.Command("git", "checkout", "-b", sourceBranch, fmt.Sprintf("origin/%s", config.Branch))
		if err := createBranch.Run(); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to create branch: %v", err)
		}
		fmt.Println("✅ Done")

		// 변경사항 스테이징 및 커밋
		fmt.Printf("  • Staging changes... ")
		addCommand := exec.Command("git", "add", config.FilePattern)
		if err := addCommand.Run(); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to stage changes: %v", err)
		}
		fmt.Println("✅ Done")

		// 커밋 생성
		fmt.Printf("  • Creating commit... ")
		commitCmd := exec.Command("git", "commit", "-m", config.CommitMessage)
		if err := commitCmd.Run(); err != nil {
			if err.Error() == "exit status 1" {
				fmt.Println("⚠️  Nothing new to commit")
			} else {
				fmt.Println("❌ Failed")
				return fmt.Errorf("failed to create commit: %v", err)
			}
		} else {
			fmt.Println("✅ Done")
		}
	} else {
		// 사용자가 지정한 브랜치 사용
		sourceBranch = config.Branch
	}

	// 커밋 및 푸시
	commitCommands := []struct {
		name string
		args []string
		desc string
	}{
		{"git", []string{"push", "-u", "origin", sourceBranch}, "Pushing changes"},
	}

	for _, cmd := range commitCommands {
		fmt.Printf("  • %s... ", cmd.desc)
		command := exec.Command(cmd.name, cmd.args...)
		command.Stdout = os.Stderr
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

	// PR URL 생성 및 출력
	fmt.Printf("\n✅ Branch '%s' has been created and pushed.\n", sourceBranch)
	fmt.Printf("✅ You can create a pull request by visiting:\n")
	fmt.Printf("   https://github.com/%s/compare/%s...%s\n",
		os.Getenv("GITHUB_REPOSITORY"),
		config.PRBase,
		sourceBranch)

	// git request-pull 명령어로 PR 생성
	fmt.Printf("  • Creating pull request from %s to %s... ", sourceBranch, config.PRBase)
	prCommand := exec.Command("git", "request-pull", config.PRBase, "origin", sourceBranch)
	if err := prCommand.Run(); err != nil {
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
