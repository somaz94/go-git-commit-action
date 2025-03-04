package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/somaz94/go-git-commit-action/internal/config"
)

// TagManager is a structure that handles tasks related to Git tags.
type TagManager struct {
	config *config.GitConfig
}

// NewTagManager creates a new TagManager instance.
func NewTagManager(config *config.GitConfig) *TagManager {
	return &TagManager{config: config}
}

// HandleGitTag handles Git tag operations.
func (tm *TagManager) HandleGitTag(ctx context.Context) error {
	return withRetry(ctx, tm.config.RetryCount, func() error {
		fmt.Println("\n🏷️  Handling Git Tag:")

		// Gets all tags and references.
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

// fetchTags fetches all tags and references.
func (tm *TagManager) fetchTags() error {
	fetchCmd := exec.Command("git", "fetch", "--tags", "--force", "origin")
	if err := fetchCmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch tags: %v", err)
	}
	return nil
}

// deleteTag deletes local and remote tags.
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

// createTag creates a new tag and pushes it to the remote repository.
func (tm *TagManager) createTag() error {
	// Check the target commit for the tag reference
	targetCommit, err := tm.getTargetCommit()
	if err != nil {
		return err
	}

	// Prepare the tag creation command
	tagArgs := tm.buildTagArgs(targetCommit)

	// Create the tag description message
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

// getTargetCommit determines the commit that the tag will point to.
func (tm *TagManager) getTargetCommit() (string, error) {
	if tm.config.TagReference == "" {
		return "", nil
	}

	// Check if the reference is valid
	cmd := exec.Command("git", "rev-parse", "--verify", tm.config.TagReference)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("invalid git reference '%s': %v", tm.config.TagReference, err)
	}

	// Get the commit SHA for the reference
	cmd = exec.Command("git", "rev-list", "-n1", tm.config.TagReference)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit SHA for '%s': %v", tm.config.TagReference, err)
	}

	return strings.TrimSpace(string(output)), nil
}

// buildTagArgs builds the arguments needed for the tag creation command.
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

// buildTagDescription builds the description for the tag creation operation.
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

// executeCommands executes a list of commands and prints the results.
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
