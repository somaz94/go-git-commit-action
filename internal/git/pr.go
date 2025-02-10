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
		}

		for _, cmd := range commands {
			fmt.Printf("  • %s... ", cmd.desc)
			command := exec.Command(cmd.name, cmd.args...)
			command.Stdout = os.Stderr
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

	// 파일 추가 및 커밋
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
