package git

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/somaz94/go-git-commit-action/internal/config"
)

// HandleGitTag 태그 처리 메인 함수
func HandleGitTag(config *config.GitConfig) error {
	fmt.Println("\n🏷️  Handling Git Tag:")

	if err := fetchTags(); err != nil {
		return err
	}

	if config.DeleteTag {
		return deleteTag(config)
	}
	return createTag(config)
}

// fetchTags 태그 가져오기
func fetchTags() error {
	fetchCmd := exec.Command("git", "fetch", "--tags", "--force", "origin")
	if err := fetchCmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch tags: %v", err)
	}
	return nil
}

// deleteTag 태그 삭제
func deleteTag(config *config.GitConfig) error {
	commands := []struct {
		name string
		args []string
		desc string
	}{
		{"git", []string{"tag", "-d", config.TagName}, "Deleting local tag"},
		{"git", []string{"push", "origin", ":refs/tags/" + config.TagName}, "Deleting remote tag"},
	}

	return executeGitCommands(commands)
}

// createTag 태그 생성
func createTag(config *config.GitConfig) error {
	targetCommit, err := getTargetCommit(config)
	if err != nil {
		return err
	}

	tagArgs := buildTagArgs(config, targetCommit)
	desc := buildTagDescription(config, targetCommit)

	commands := []struct {
		name string
		args []string
		desc string
	}{
		{"git", tagArgs, desc},
		{"git", []string{"push", "-f", "origin", config.TagName}, "Pushing tag to remote"},
	}

	return executeGitCommands(commands)
}

// Helper functions
func getTargetCommit(config *config.GitConfig) (string, error) {
	if config.TagReference == "" {
		return "", nil
	}

	// Check if reference exists
	if err := exec.Command("git", "rev-parse", "--verify", config.TagReference).Run(); err != nil {
		return "", fmt.Errorf("invalid git reference '%s': %v", config.TagReference, err)
	}

	// Get commit SHA
	cmd := exec.Command("git", "rev-list", "-n1", config.TagReference)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit SHA for '%s': %v", config.TagReference, err)
	}

	return strings.TrimSpace(string(output)), nil
}

func buildTagArgs(config *config.GitConfig, targetCommit string) []string {
	if config.TagMessage != "" {
		if targetCommit != "" {
			return []string{"tag", "-f", "-a", config.TagName, targetCommit, "-m", config.TagMessage}
		}
		return []string{"tag", "-f", "-a", config.TagName, "-m", config.TagMessage}
	}

	if targetCommit != "" {
		return []string{"tag", "-f", config.TagName, targetCommit}
	}
	return []string{"tag", "-f", config.TagName}
}

func buildTagDescription(config *config.GitConfig, targetCommit string) string {
	desc := "Creating local tag " + config.TagName
	if config.TagReference != "" && targetCommit != "" {
		if targetCommit != config.TagReference {
			desc += fmt.Sprintf(" pointing to %s (%s)", config.TagReference, targetCommit[:8])
		} else {
			desc += fmt.Sprintf(" pointing to %s", targetCommit[:8])
		}
	}
	return desc
}
