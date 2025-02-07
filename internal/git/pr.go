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

	var sourceBranch string
	if config.AutoBranch {
		// 자동 브랜치 생성
		sourceBranch = fmt.Sprintf("update-files-%s", time.Now().Format("20060102-150405"))

		commands := []struct {
			name string
			args []string
			desc string
		}{
			{"git", []string{"checkout", "-b", sourceBranch}, "Creating new branch"},
			{"git", []string{"push", "-u", "origin", sourceBranch}, "Pushing new branch"},
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
		// 사용자가 지정한 브랜치 사용
		sourceBranch = config.Branch
	}

	// GitHub CLI를 사용하여 PR 생성
	fmt.Printf("  • Creating pull request from %s to %s... ", sourceBranch, config.PRBase)
	prCommand := exec.Command("gh", "pr", "create",
		"--base", config.PRBase,
		"--head", sourceBranch,
		"--title", config.PRTitle,
		"--fill") // 자동으로 PR 내용 채우기

	if err := prCommand.Run(); err != nil {
		fmt.Println("❌ Failed")
		return fmt.Errorf("failed to create pull request: %v", err)
	}
	fmt.Println("✅ Done")

	// 소스 브랜치 삭제 (옵션이 활성화된 경우)
	if config.DeleteSourceBranch {
		fmt.Printf("  • Deleting source branch %s... ", sourceBranch)
		deleteCommand := exec.Command("git", "push", "origin", "--delete", sourceBranch)
		if err := deleteCommand.Run(); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to delete source branch: %v", err)
		}
		fmt.Println("✅ Done")
	}

	return nil
}
