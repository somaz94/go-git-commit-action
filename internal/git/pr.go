package git

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/somaz94/go-git-commit-action/internal/config"
)

// GitHubClient는 GitHub API 요청을 위한 구조체입니다.
type GitHubClient struct {
	token      string
	baseURL    string
	repository string
}

// NewGitHubClient는 새로운 GitHubClient 인스턴스를 생성합니다.
func NewGitHubClient(token, repository string) *GitHubClient {
	return &GitHubClient{
		token:      token,
		baseURL:    "https://api.github.com",
		repository: repository,
	}
}

// CreatePullRequest는 GitHub API를 사용하여 PR을 생성합니다.
func (c *GitHubClient) CreatePullRequest(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	// API 요청 로직
	return nil, nil
}

// AddLabels는 PR에 라벨을 추가합니다.
func (c *GitHubClient) AddLabels(ctx context.Context, prNumber int, labels []string) error {
	// 라벨 추가 로직
	return nil
}

// ClosePullRequest는 PR을 닫습니다.
func (c *GitHubClient) ClosePullRequest(ctx context.Context, prNumber int) error {
	// PR 닫기 로직
	return nil
}

// CreatePullRequest는 PR을 생성하는 메인 함수입니다.
func CreatePullRequest(config *config.GitConfig) error {
	fmt.Println("\n🔄 Creating Pull Request:")

	// 소스 브랜치 준비
	sourceBranch, err := prepareSourceBranch(config)
	if err != nil {
		return err
	}

	// 브랜치 간 변경사항 확인
	if err := checkBranchDifferences(config); err != nil {
		return err
	}

	// PR 생성
	prResponse, err := createGitHubPR(config)
	if err != nil {
		return err
	}

	// PR 응답 처리
	if err := handlePRResponse(config, prResponse, sourceBranch); err != nil {
		return err
	}

	fmt.Println("\n✨ Git Commit Action Completed Successfully!\n" +
		"=========================================")

	return nil
}

// prepareSourceBranch는 소스 브랜치를 준비합니다.
func prepareSourceBranch(config *config.GitConfig) (string, error) {
	var sourceBranch string

	if config.AutoBranch {
		// 타임스탬프가 포함된 브랜치 이름 생성
		sourceBranch = fmt.Sprintf("update-files-%s", time.Now().Format("20060102-150405"))
		config.PRBranch = sourceBranch

		// 새 브랜치 생성 및 전환
		fmt.Printf("  • Creating new branch %s... ", sourceBranch)
		if err := exec.Command("git", "checkout", "-b", sourceBranch).Run(); err != nil {
			fmt.Println("❌ Failed")
			return "", fmt.Errorf("failed to create branch: %v", err)
		}
		fmt.Println("✅ Done")

		// 변경사항을 새 브랜치에 커밋하고 푸시
		if err := commitAndPushChanges(config); err != nil {
			return "", err
		}
	} else {
		// auto_branch=false일 때는 pr_branch로 체크아웃
		sourceBranch = config.PRBranch
		fmt.Printf("  • Checking out branch %s... ", sourceBranch)
		if err := exec.Command("git", "checkout", sourceBranch).Run(); err != nil {
			fmt.Println("❌ Failed")
			return "", fmt.Errorf("failed to checkout branch: %v", err)
		}
		fmt.Println("✅ Done")
	}

	return sourceBranch, nil
}

// commitAndPushChanges는 변경사항을 커밋하고 푸시합니다.
func commitAndPushChanges(config *config.GitConfig) error {
	commitCommands := []struct {
		name string
		args []string
		desc string
	}{
		{"git", []string{"add", config.FilePattern}, "Adding files"},
		{"git", []string{"commit", "-m", config.CommitMessage}, "Committing changes"},
		{"git", []string{"push", "-u", "origin", config.PRBranch}, "Pushing changes"},
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

// checkBranchDifferences는 PR 기본 브랜치와 소스 브랜치 간의 차이를 확인합니다.
func checkBranchDifferences(config *config.GitConfig) error {
	fmt.Printf("\n📊 Changed files between %s and %s:\n", config.PRBase, config.PRBranch)

	// 두 브랜치 가져오기
	if err := fetchBranches(config); err != nil {
		return err
	}

	// 변경된 파일 확인
	diffFiles := exec.Command("git", "diff", fmt.Sprintf("origin/%s..origin/%s", config.PRBase, config.PRBranch), "--name-status")
	filesOutput, _ := diffFiles.Output()

	if len(filesOutput) == 0 {
		fmt.Println("No changes detected")
		if config.SkipIfEmpty {
			return nil
		}
		return fmt.Errorf("no changes to create PR")
	}
	fmt.Printf("%s\n", string(filesOutput))

	// PR URL 생성 및 출력
	fmt.Printf("\n✅ Branch '%s' is ready for PR.\n", config.PRBranch)
	prURL := fmt.Sprintf("https://github.com/%s/compare/%s...%s",
		os.Getenv("GITHUB_REPOSITORY"),
		config.PRBase,
		config.PRBranch)
	fmt.Printf("✅ You can create a pull request by visiting:\n   %s\n", prURL)

	return nil
}

// fetchBranches는 기본 브랜치와 소스 브랜치를 가져옵니다.
func fetchBranches(config *config.GitConfig) error {
	fetchBaseCmd := exec.Command("git", "fetch", "origin", config.PRBase)
	if err := fetchBaseCmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch base branch: %v", err)
	}

	fetchBranchCmd := exec.Command("git", "fetch", "origin", config.PRBranch)
	if err := fetchBranchCmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch source branch: %v", err)
	}

	return nil
}

// createGitHubPR는 GitHub API를 사용하여 PR을 생성합니다.
func createGitHubPR(config *config.GitConfig) (map[string]interface{}, error) {
	fmt.Printf("  • Creating pull request from %s to %s... ", config.PRBranch, config.PRBase)

	// PR 데이터 준비
	prData, err := preparePRData(config)
	if err != nil {
		return nil, err
	}

	// GitHub API 호출
	return callGitHubAPI(config, prData)
}

// preparePRData는 PR 생성에 필요한 데이터를 준비합니다.
func preparePRData(config *config.GitConfig) (map[string]interface{}, error) {
	// GitHub Run ID 가져오기
	runID := os.Getenv("GITHUB_RUN_ID")

	// 현재 커밋 SHA 가져오기
	commitCmd := exec.Command("git", "rev-parse", "HEAD")
	commitSHA, err := commitCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get commit SHA: %v", err)
	}
	commitID := strings.TrimSpace(string(commitSHA))

	// PR 제목 설정
	title := config.PRTitle
	if title == "" {
		title = fmt.Sprintf("Auto PR: %s to %s (Run ID: %s)", config.PRBranch, config.PRBase, runID)
	}

	// PR 본문 설정
	body := config.PRBody
	if body == "" {
		body = fmt.Sprintf("Created by Go Git Commit Action\nSource: %s\nTarget: %s\nCommit: %s\nGitHub Run ID: %s",
			config.PRBranch, config.PRBase, commitID, runID)
	}

	// PR 요청 데이터
	prData := map[string]interface{}{
		"title": title,
		"head":  config.PRBranch,
		"base":  config.PRBase,
		"body":  body,
	}

	return prData, nil
}

// callGitHubAPI는 GitHub API를 호출하여 PR을 생성합니다.
func callGitHubAPI(config *config.GitConfig, prData map[string]interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(prData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal PR data: %v", err)
	}

	// GitHub API를 사용하여 PR 생성
	curlCmd := exec.Command("curl", "-s", "-X", "POST",
		"-H", fmt.Sprintf("Authorization: Bearer %s", config.GitHubToken),
		"-H", "Accept: application/vnd.github+json",
		"-H", "Content-Type: application/json",
		fmt.Sprintf("https://api.github.com/repos/%s/pulls", os.Getenv("GITHUB_REPOSITORY")),
		"-d", string(jsonData))

	output, err := curlCmd.CombinedOutput()
	if err != nil {
		fmt.Println("⚠️  Failed to create PR automatically")
		fmt.Printf("Error: %v\n", err)
		fmt.Printf("Response: %s\n", string(output))
		return nil, fmt.Errorf("failed to execute curl command: %v", err)
	}

	// 응답 파싱
	var response map[string]interface{}
	if err := json.Unmarshal(output, &response); err != nil {
		fmt.Printf("Raw response: %s\n", string(output))
		return nil, fmt.Errorf("failed to parse PR response: %v", err)
	}

	return response, nil
}

// handlePRResponse는 PR 생성 응답을 처리합니다.
func handlePRResponse(config *config.GitConfig, response map[string]interface{}, sourceBranch string) error {
	// 오류 메시지 확인
	if errMsg, ok := response["message"].(string); ok {
		fmt.Printf("GitHub API Error: %s\n", errMsg)
		if errors, ok := response["errors"].([]interface{}); ok {
			fmt.Println("Error details:")
			for _, err := range errors {
				if errMap, ok := err.(map[string]interface{}); ok {
					fmt.Printf("  • %v\n", errMap)
					// PR이 이미 존재하는 경우 처리
					if errMap["message"].(string) == "A pull request already exists for somaz94:test." {
						return handleExistingPR(config)
					}
				}
			}
		}
		return fmt.Errorf("GitHub API error: %s", errMsg)
	}

	// PR URL 확인
	if htmlURL, ok := response["html_url"].(string); ok {
		fmt.Println("✅ Done")
		fmt.Printf("Pull request created: %s\n", htmlURL)

		// PR 번호 처리
		if number, ok := response["number"].(float64); ok {
			prNumber := int(number)

			// 라벨 추가
			if len(config.PRLabels) > 0 {
				if err := addLabelsToIssue(config, prNumber); err != nil {
					return err
				}
			}

			// PR 닫기
			if config.PRClosed {
				if err := closePullRequest(config, prNumber); err != nil {
					return err
				}
			}
		}
	} else {
		fmt.Println("⚠️  Failed to create PR")
		fmt.Printf("Response: %v\n", response)
		return fmt.Errorf("failed to get PR URL from response")
	}

	// 소스 브랜치 삭제
	if config.DeleteSourceBranch && config.AutoBranch {
		if err := deleteSourceBranch(sourceBranch); err != nil {
			return err
		}
	}

	return nil
}

// handleExistingPR는 이미 존재하는 PR을 처리합니다.
func handleExistingPR(config *config.GitConfig) error {
	fmt.Println("⚠️  Pull request already exists")

	// 기존 PR 찾기
	searchCmd := exec.Command("curl", "-s",
		"-H", fmt.Sprintf("Authorization: Bearer %s", config.GitHubToken),
		"-H", "Accept: application/vnd.github+json",
		"-H", "X-GitHub-Api-Version: 2022-11-28",
		fmt.Sprintf("https://api.github.com/repos/%s/pulls?head=%s&base=%s",
			os.Getenv("GITHUB_REPOSITORY"),
			config.PRBranch,
			config.PRBase))

	searchOutput, _ := searchCmd.CombinedOutput()
	var prs []map[string]interface{}
	if err := json.Unmarshal(searchOutput, &prs); err == nil && len(prs) > 0 {
		if number, ok := prs[0]["number"].(float64); ok {
			prNumber := int(number)
			fmt.Printf("Found existing PR #%d\n", prNumber)

			// 라벨 추가
			if len(config.PRLabels) > 0 {
				if err := addLabelsToIssue(config, prNumber); err != nil {
					return err
				}
			}

			// PR 닫기
			if config.PRClosed {
				if err := closePullRequest(config, prNumber); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// addLabelsToIssue는 이슈/PR에 라벨을 추가합니다.
func addLabelsToIssue(config *config.GitConfig, prNumber int) error {
	fmt.Printf("  • Adding labels to PR #%d... ", prNumber)
	labelsData := map[string]interface{}{
		"labels": config.PRLabels,
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
		fmt.Printf("Error: %v\n", err)
		fmt.Printf("Response: %s\n", string(labelsOutput))
		return fmt.Errorf("failed to add labels: %v", err)
	} else {
		fmt.Println("✅ Done")
	}

	return nil
}

// closePullRequest는 PR을 닫습니다.
func closePullRequest(config *config.GitConfig, prNumber int) error {
	fmt.Printf("  • Closing pull request #%d... ", prNumber)
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
		fmt.Printf("Error: %v\n", err)
		fmt.Printf("Response: %s\n", string(closeOutput))
		return fmt.Errorf("failed to close PR: %v", err)
	} else {
		fmt.Println("✅ Done")
	}

	return nil
}

// deleteSourceBranch는 소스 브랜치를 삭제합니다.
func deleteSourceBranch(sourceBranch string) error {
	fmt.Printf("\n  • Deleting source branch %s... ", sourceBranch)
	deleteCommand := exec.Command("git", "push", "origin", "--delete", sourceBranch)
	if err := deleteCommand.Run(); err != nil {
		fmt.Println("❌ Failed")
		return fmt.Errorf("failed to delete source branch: %v", err)
	}
	fmt.Println("✅ Done")

	return nil
}
