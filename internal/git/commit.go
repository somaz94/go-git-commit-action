package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/somaz94/go-git-commit-action/internal/config"
)

// 재시도 로직을 위한 헬퍼 함수
func withRetry(ctx context.Context, maxRetries int, operation func() error) error {
	var lastErr error
	for i := range make([]struct{}, maxRetries) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := operation(); err != nil {
				lastErr = err
				time.Sleep(time.Second * time.Duration(i+1))
				continue
			}
			return nil
		}
	}
	return fmt.Errorf("operation failed after %d retries: %v", maxRetries, lastErr)
}

// RunGitCommit은 Git 커밋 작업을 실행합니다.
func RunGitCommit(config *config.GitConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Timeout)*time.Second)
	defer cancel()

	// 기존 코드를 재시도 로직으로 래핑
	return withRetry(ctx, config.RetryCount, func() error {
		// 설정 검증
		if err := validateConfig(config); err != nil {
			return err
		}

		// 디버그 정보 출력
		printDebugInfo()

		// 작업 디렉토리 변경
		if err := changeWorkingDirectory(config); err != nil {
			return err
		}

		// Git 기본 설정
		if err := setupGitConfig(config); err != nil {
			return err
		}

		// 브랜치 처리
		if err := handleBranch(config); err != nil {
			return err
		}

		// 변경사항 확인
		if empty, err := checkIfEmpty(config); err != nil {
			return err
		} else if empty {
			fmt.Println("\n⚠️  No changes detected and skip_if_empty is true. Skipping commit process.")
			return nil
		}

		// PR 생성 또는 직접 커밋
		if config.CreatePR {
			return handlePullRequestFlow(config)
		} else {
			return commitChanges(config)
		}
	})
}

// validateConfig는 필수 설정을 검증합니다.
func validateConfig(config *config.GitConfig) error {
	if config.CreatePR {
		if !config.AutoBranch && config.PRBranch == "" {
			return fmt.Errorf("pr_branch must be specified when auto_branch is false and create_pr is true")
		}
		if config.PRBase == "" {
			return fmt.Errorf("pr_base must be specified when create_pr is true")
		}
		if config.GitHubToken == "" {
			return fmt.Errorf("github_token must be specified when create_pr is true")
		}
	}
	return nil
}

// printDebugInfo는 디버그 정보를 출력합니다.
func printDebugInfo() {
	currentDir, _ := os.Getwd()
	fmt.Println("\n🚀 Starting Git Commit Action\n" +
		"================================")

	fmt.Println("\n📋 Configuration:")
	fmt.Printf("  • Working Directory: %s\n", currentDir)

	fmt.Println("\n📁 Directory Contents:")
	files, _ := os.ReadDir(".")
	for _, file := range files {
		fmt.Printf("  • %s\n", file.Name())
	}
}

// changeWorkingDirectory는 작업 디렉토리를 변경합니다.
func changeWorkingDirectory(config *config.GitConfig) error {
	if config.RepoPath != "." {
		if err := os.Chdir(config.RepoPath); err != nil {
			return fmt.Errorf("❌ Failed to change directory: %v", err)
		}
		newDir, _ := os.Getwd()
		fmt.Printf("\n📂 Changed to directory: %s\n", newDir)
	}
	return nil
}

// setupGitConfig는 Git 기본 설정을 수행합니다.
func setupGitConfig(config *config.GitConfig) error {
	baseCommands := []struct {
		name string
		args []string
		desc string
	}{
		{"git", []string{"config", "--global", "--add", "safe.directory", "/app"}, "Setting safe directory (/app)"},
		{"git", []string{"config", "--global", "--add", "safe.directory", "/github/workspace"}, "Setting safe directory (/github/workspace)"},
		{"git", []string{"config", "--global", "user.email", config.UserEmail}, "Configuring user email"},
		{"git", []string{"config", "--global", "user.name", config.UserName}, "Configuring user name"},
		{"git", []string{"config", "--global", "--list"}, "Checking git configuration"},
	}

	fmt.Println("\n⚙️  Executing Git Commands:")
	for _, cmd := range baseCommands {
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
	return nil
}

// handleBranch는 브랜치 관련 작업을 처리합니다.
func handleBranch(config *config.GitConfig) error {
	// 로컬 브랜치 확인
	checkLocalBranch := exec.Command("git", "rev-parse", "--verify", config.Branch)
	// 원격 브랜치 확인
	checkRemoteBranch := exec.Command("git", "ls-remote", "--heads", "origin", config.Branch)

	if checkLocalBranch.Run() != nil && checkRemoteBranch.Run() != nil {
		// 로컬과 원격 모두에 브랜치가 없는 경우 새로 생성
		return createNewBranch(config)
	} else if checkLocalBranch.Run() != nil {
		// 원격에는 있지만 로컬에는 없는 경우 체크아웃
		return checkoutRemoteBranch(config)
	}
	return nil
}

// createNewBranch는 새 브랜치를 생성합니다.
func createNewBranch(config *config.GitConfig) error {
	fmt.Printf("\n⚠️  Branch '%s' not found, creating it...\n", config.Branch)
	createCommands := []struct {
		name string
		args []string
		desc string
	}{
		{"git", []string{"checkout", "-b", config.Branch}, "Creating new branch"},
		{"git", []string{"push", "-u", "origin", config.Branch}, "Pushing new branch"},
	}

	for _, cmd := range createCommands {
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
	return nil
}

// checkoutRemoteBranch는 원격 브랜치를 체크아웃합니다.
func checkoutRemoteBranch(config *config.GitConfig) error {
	fmt.Printf("\n⚠️  Checking out existing remote branch '%s'...\n", config.Branch)

	// 수정된 파일 확인
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusOutput, err := statusCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get modified files: %v", err)
	}

	// 변경사항 백업
	backups, err := backupChanges(config, string(statusOutput))
	if err != nil {
		return err
	}

	// 변경사항 스태시
	if err := stashChanges(); err != nil {
		return err
	}

	// 원격 브랜치 체크아웃
	if err := fetchAndCheckout(config); err != nil {
		return err
	}

	// 변경사항 복원
	return restoreChanges(backups)
}

// FileBackup은 파일 백업을 위한 구조체입니다.
type FileBackup struct {
	path    string
	content []byte
}

// backupChanges는 변경된 파일을 백업합니다.
func backupChanges(config *config.GitConfig, statusOutput string) ([]FileBackup, error) {
	fmt.Printf("  • Backing up changes... ")

	var backups []FileBackup

	for _, line := range strings.Split(statusOutput, "\n") {
		if line == "" {
			continue
		}

		// 상태 코드와 파일 경로 분리
		status := line[:2]
		fullPath := strings.TrimSpace(line[3:])

		// config.RepoPath 기준으로 상대 경로 계산
		relPath := fullPath
		if config.RepoPath != "." {
			relPath = strings.TrimPrefix(fullPath, config.RepoPath+"/")
		}

		fmt.Printf("\n    - Found modified file: %s (status: %s)", relPath, status)

		// 삭제되지 않은 경우만 백업
		if status != " D" && status != "D " {
			content, err := os.ReadFile(relPath)
			if err != nil {
				fmt.Println("❌ Failed")
				return nil, fmt.Errorf("failed to read file %s: %v", relPath, err)
			}
			backups = append(backups, FileBackup{path: relPath, content: content})
		}
	}
	fmt.Println("✅ Done")
	return backups, nil
}

// stashChanges는 변경사항을 스태시합니다.
func stashChanges() error {
	fmt.Printf("  • Stashing changes... ")
	stashCmd := exec.Command("git", "stash", "push", "-u")
	stashCmd.Stdout = os.Stdout
	stashCmd.Stderr = os.Stderr
	if err := stashCmd.Run(); err != nil {
		fmt.Println("❌ Failed")
		return fmt.Errorf("failed to stash changes: %v", err)
	}
	fmt.Println("✅ Done")
	return nil
}

// fetchAndCheckout은 원격 브랜치를 가져와 체크아웃합니다.
func fetchAndCheckout(config *config.GitConfig) error {
	checkoutCommands := []struct {
		name string
		args []string
		desc string
	}{
		{"git", []string{"fetch", "origin", config.Branch}, "Fetching remote branch"},
		{"git", []string{"checkout", config.Branch}, "Checking out branch"},
		{"git", []string{"reset", "--hard", fmt.Sprintf("origin/%s", config.Branch)}, "Resetting to remote state"},
	}

	for _, cmd := range checkoutCommands {
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
	return nil
}

// restoreChanges는 백업된 변경사항을 복원합니다.
func restoreChanges(backups []FileBackup) error {
	fmt.Printf("  • Restoring changes... ")
	for _, backup := range backups {
		// 필요한 경우 디렉토리 생성
		dir := filepath.Dir(backup.path)
		if dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Println("❌ Failed")
				return fmt.Errorf("failed to create directory %s: %v", dir, err)
			}
		}

		if err := os.WriteFile(backup.path, backup.content, 0644); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to restore file %s: %v", backup.path, err)
		}
	}
	fmt.Println("✅ Done")
	return nil
}

// checkIfEmpty는 변경사항이 있는지 확인합니다.
func checkIfEmpty(config *config.GitConfig) (bool, error) {
	// 1. 작업 디렉토리의 로컬 변경사항 확인
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusOutput, err := statusCmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check git status: %v", err)
	}

	// 2. 브랜치 간 차이점 확인
	diffCmd := exec.Command("git", "diff", fmt.Sprintf("origin/%s...%s", config.PRBase, config.PRBranch), "--name-only")
	diffOutput, err := diffCmd.Output()
	if err != nil {
		// 오류 발생 시(예: 새 브랜치), 비어있지 않은 것으로 간주
		diffOutput = []byte("new-branch")
	}

	isEmpty := len(statusOutput) == 0 && len(diffOutput) == 0

	// 디버그 정보 출력
	fmt.Printf("\n📊 Change Detection:\n")
	fmt.Printf("  • Local changes: %v\n", len(statusOutput) > 0)
	fmt.Printf("  • Branch differences: %v\n", len(diffOutput) > 0)
	if len(statusOutput) > 0 {
		fmt.Printf("  • Local changes details:\n%s\n", string(statusOutput))
	}
	if len(diffOutput) > 0 {
		fmt.Printf("  • Branch differences details:\n%s\n", string(diffOutput))
	}

	return isEmpty && config.SkipIfEmpty, nil
}

// handlePullRequestFlow는 PR 생성 흐름을 처리합니다.
func handlePullRequestFlow(config *config.GitConfig) error {
	if config.AutoBranch {
		// AutoBranch가 true인 경우, PR 생성 함수가 새 브랜치를 생성하고 커밋
		if err := CreatePullRequest(config); err != nil {
			return fmt.Errorf("failed to create pull request: %v", err)
		}
	} else {
		// AutoBranch가 false인 경우, 먼저 지정된 브랜치에 커밋
		if err := commitChanges(config); err != nil {
			return err
		}

		// 커밋 후 PR 생성 (pr_branch와 pr_base 사용)
		if err := CreatePullRequest(config); err != nil {
			return fmt.Errorf("failed to create pull request: %v", err)
		}
	}
	return nil
}

// commitChanges는 변경사항을 커밋하고 푸시합니다.
func commitChanges(config *config.GitConfig) error {
	commitCommands := []struct {
		name string
		args []string
		desc string
	}{
		{"git", []string{"add", config.FilePattern}, "Adding files"},
		{"git", []string{"commit", "-m", config.CommitMessage}, "Committing changes"},
		{"git", []string{"push", "origin", config.Branch}, "Pushing to remote"},
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
	return nil
}
