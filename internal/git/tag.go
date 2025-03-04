package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/somaz94/go-git-commit-action/internal/config"
)

// TagManager는 Git 태그 관련 작업을 처리하는 구조체입니다.
type TagManager struct {
	config *config.GitConfig
}

// NewTagManager는 새로운 TagManager 인스턴스를 생성합니다.
func NewTagManager(config *config.GitConfig) *TagManager {
	return &TagManager{config: config}
}

// HandleGitTag는 Git 태그 작업을 처리합니다.
func (tm *TagManager) HandleGitTag(ctx context.Context) error {
	return withRetry(ctx, tm.config.RetryCount, func() error {
		fmt.Println("\n🏷️  Handling Git Tag:")

		// 모든 태그와 참조를 가져옵니다.
		if err := tm.fetchTags(); err != nil {
			return err
		}

		if tm.config.DeleteTag {
			return tm.deleteTag()
		} else {
			return tm.createTag()
		}
	})
}

// fetchTags는 모든 태그와 참조를 가져옵니다.
func (tm *TagManager) fetchTags() error {
	fetchCmd := exec.Command("git", "fetch", "--tags", "--force", "origin")
	if err := fetchCmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch tags: %v", err)
	}
	return nil
}

// deleteTag는 로컬 및 원격 태그를 삭제합니다.
func (tm *TagManager) deleteTag() error {
	commands := []struct {
		name string
		args []string
		desc string
	}{
		{"git", []string{"tag", "-d", tm.config.TagName}, "Deleting local tag"},
		{"git", []string{"push", "origin", ":refs/tags/" + tm.config.TagName}, "Deleting remote tag"},
	}

	return tm.executeCommands(commands)
}

// createTag는 새 태그를 생성하고 원격 저장소에 푸시합니다.
func (tm *TagManager) createTag() error {
	// 태그 참조 대상 커밋 확인
	targetCommit, err := tm.getTargetCommit()
	if err != nil {
		return err
	}

	// 태그 생성 명령 준비
	tagArgs := tm.buildTagArgs(targetCommit)

	// 태그 설명 메시지 생성
	desc := tm.buildTagDescription(targetCommit)

	commands := []struct {
		name string
		args []string
		desc string
	}{
		{"git", tagArgs, desc},
		{"git", []string{"push", "-f", "origin", tm.config.TagName}, "Pushing tag to remote"},
	}

	return tm.executeCommands(commands)
}

// getTargetCommit은 태그가 가리킬 커밋을 결정합니다.
func (tm *TagManager) getTargetCommit() (string, error) {
	if tm.config.TagReference == "" {
		return "", nil
	}

	// 참조가 유효한지 확인
	cmd := exec.Command("git", "rev-parse", "--verify", tm.config.TagReference)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("invalid git reference '%s': %v", tm.config.TagReference, err)
	}

	// 참조에 대한 커밋 SHA 가져오기
	cmd = exec.Command("git", "rev-list", "-n1", tm.config.TagReference)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit SHA for '%s': %v", tm.config.TagReference, err)
	}

	return strings.TrimSpace(string(output)), nil
}

// buildTagArgs는 태그 생성 명령에 필요한 인수를 구성합니다.
func (tm *TagManager) buildTagArgs(targetCommit string) []string {
	var tagArgs []string

	if tm.config.TagMessage != "" {
		if targetCommit != "" {
			tagArgs = []string{"tag", "-f", "-a", tm.config.TagName, targetCommit, "-m", tm.config.TagMessage}
		} else {
			tagArgs = []string{"tag", "-f", "-a", tm.config.TagName, "-m", tm.config.TagMessage}
		}
	} else {
		if targetCommit != "" {
			tagArgs = []string{"tag", "-f", tm.config.TagName, targetCommit}
		} else {
			tagArgs = []string{"tag", "-f", tm.config.TagName}
		}
	}

	return tagArgs
}

// buildTagDescription은 태그 생성 작업에 대한 설명을 생성합니다.
func (tm *TagManager) buildTagDescription(targetCommit string) string {
	desc := "Creating local tag " + tm.config.TagName

	if tm.config.TagReference != "" && targetCommit != "" {
		if targetCommit != tm.config.TagReference {
			desc += fmt.Sprintf(" pointing to %s (%s)", tm.config.TagReference, targetCommit[:8])
		} else {
			desc += fmt.Sprintf(" pointing to %s", targetCommit[:8])
		}
	}

	return desc
}

// executeCommands는 명령 목록을 실행하고 결과를 출력합니다.
func (tm *TagManager) executeCommands(commands []struct {
	name string
	args []string
	desc string
}) error {
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

	return nil
}
