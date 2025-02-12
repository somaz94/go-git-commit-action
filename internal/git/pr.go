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

// CreatePullRequest PR 생성 메인 함수
func CreatePullRequest(config *config.GitConfig) error {
	fmt.Println("\n🔄 Creating Pull Request:")

	sourceBranch, err := prepareBranch(config)
	if err != nil {
		return err
	}

	if err := commitAndPushChanges(config, sourceBranch); err != nil {
		return err
	}

	if err := createGitHubPR(config, sourceBranch); err != nil {
		return err
	}

	return handlePRCleanup(config, sourceBranch)
}

// prepareBranch 브랜치 준비
func prepareBranch(config *config.GitConfig) (string, error) {
	if config.AutoBranch {
		return createAutoBranch()
	}
	return validateExistingBranch(config)
}

// createAutoBranch 자동 브랜치 생성
func createAutoBranch() (string, error) {
	sourceBranch := fmt.Sprintf("update-files-%s", time.Now().Format("20060102-150405"))
	fmt.Printf("  • Creating new branch %s... ", sourceBranch)

	if err := exec.Command("git", "checkout", "-b", sourceBranch).Run(); err != nil {
		fmt.Println("❌ Failed")
		return "", fmt.Errorf("failed to create branch: %v", err)
	}
	fmt.Println("✅ Done")

	return sourceBranch, nil
}

// validateExistingBranch 기존 브랜치 검증
func validateExistingBranch(config *config.GitConfig) (string, error) {
	if config.PRBranch == "" {
		return "", fmt.Errorf("pr_branch must be specified when auto_branch is false")
	}
	return config.PRBranch, nil
}

// commitAndPushChanges 변경사항 커밋 및 푸시
func commitAndPushChanges(config *config.GitConfig, sourceBranch string) error {
	if !config.AutoBranch {
		return verifyChanges(config, sourceBranch)
	}

	commitCommands := []struct {
		name string
		args []string
		desc string
	}{
		{"git", []string{"add", config.FilePattern}, "Adding files"},
		{"git", []string{"commit", "-m", config.CommitMessage}, "Committing changes"},
		{"git", []string{"push", "-u", "origin", sourceBranch}, "Pushing changes"},
	}

	return executeGitCommands(commitCommands)
}

// createGitHubPR GitHub PR 생성
func createGitHubPR(config *config.GitConfig, sourceBranch string) error {
	runID := os.Getenv("GITHUB_RUN_ID")
	prURL := fmt.Sprintf("https://github.com/%s/compare/%s...%s",
		os.Getenv("GITHUB_REPOSITORY"),
		config.PRBase,
		sourceBranch)

	jsonData := generatePRJSON(config, sourceBranch, runID)
	return submitPRToGitHub(config, jsonData, prURL)
}

// handlePRCleanup PR 생성 후 정리 작업
func handlePRCleanup(config *config.GitConfig, sourceBranch string) error {
	if config.DeleteSourceBranch && config.AutoBranch {
		fmt.Printf("\n  • Deleting source branch %s... ", sourceBranch)
		if err := exec.Command("git", "push", "origin", "--delete", sourceBranch).Run(); err != nil {
			fmt.Println("❌ Failed")
			return fmt.Errorf("failed to delete source branch: %v", err)
		}
		fmt.Println("✅ Done")
	}

	fmt.Println("\n✨ Git Commit Action Completed Successfully!\n" +
		"=========================================")

	return nil
}

// Helper functions
func verifyChanges(config *config.GitConfig, sourceBranch string) error {
	fmt.Printf("\n📊 Changed files between %s and %s:\n", config.PRBase, sourceBranch)
	diffFiles := exec.Command("git", "diff", fmt.Sprintf("origin/%s..origin/%s", config.PRBase, sourceBranch), "--name-status")
	filesOutput, _ := diffFiles.Output()

	if len(filesOutput) == 0 {
		fmt.Println("No changes detected")
		return fmt.Errorf("no changes to create PR")
	}
	fmt.Printf("%s\n", string(filesOutput))
	return nil
}

func generatePRJSON(config *config.GitConfig, sourceBranch, runID string) string {
	return fmt.Sprintf(`{
		"title": "Auto PR: %s to %s (Run ID: %s)",
		"head": "%s",
		"base": "%s",
		"body": "Created by Go Git Commit Action\nSource: %s\nTarget: %s\nGitHub Run ID: %s"
	}`, sourceBranch, config.PRBase, runID, sourceBranch, config.PRBase, sourceBranch, config.PRBase, runID)
}

func submitPRToGitHub(config *config.GitConfig, jsonData, prURL string) error {
	curlCmd := exec.Command("curl", "-s", "-X", "POST",
		"-H", fmt.Sprintf("Authorization: Bearer %s", config.GitHubToken),
		"-H", "Accept: application/vnd.github+json",
		"-H", "Content-Type: application/json",
		fmt.Sprintf("https://api.github.com/repos/%s/pulls", os.Getenv("GITHUB_REPOSITORY")),
		"-d", jsonData)

	output, err := curlCmd.CombinedOutput()
	return handlePRResponse(output, err, prURL)
}

func handlePRResponse(output []byte, err error, prURL string) error {
	if err != nil {
		fmt.Println("⚠️  Failed to create PR automatically")
		fmt.Printf("Error executing curl: %v\n", err)
		fmt.Printf("Response: %s\n", string(output))
		fmt.Printf("You can create a pull request manually by visiting:\n   %s\n", prURL)
		return err
	}

	if strings.Contains(string(output), "html_url") {
		fmt.Printf("✅ Done\n")
		var response map[string]interface{}
		if err := json.Unmarshal(output, &response); err == nil {
			if htmlURL, ok := response["html_url"].(string); ok {
				fmt.Printf("Pull request created: %s\n", htmlURL)
			}
		}
		return nil
	}

	fmt.Printf("⚠️  Failed to create PR\n")
	fmt.Printf("You can create a pull request manually by visiting:\n   %s\n", prURL)
	return fmt.Errorf("failed to create PR")
}
