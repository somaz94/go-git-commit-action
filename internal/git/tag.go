package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/somaz94/go-git-commit-action/internal/config"
	"github.com/somaz94/go-git-commit-action/internal/gitcmd"
)

// TagCommand defines a command to be executed for tag operations
type TagCommand struct {
	name string
	args []string
	desc string
}

// TagManager handles all operations related to Git tags.
// It provides methods for creating, deleting, and managing Git tags.
type TagManager struct {
	config *config.GitConfig
}

// NewTagManager creates a new TagManager instance with the provided configuration.
// This is the entry point for all tag-related operations.
func NewTagManager(config *config.GitConfig) *TagManager {
	return &TagManager{config: config}
}

// HandleGitTag orchestrates the Git tag operations based on configuration.
// It determines whether to create or delete tags and handles the operation
// with retry capability for transient errors.
func (tm *TagManager) HandleGitTag(ctx context.Context) error {
	return withRetry(ctx, tm.config.RetryCount, func() error {
		fmt.Println("\n🏷️  Handling Git Tag:")

		// Fetch all tags to ensure we're working with the latest data
		if err := tm.fetchTags(); err != nil {
			return err
		}

		// Either delete or create a tag based on the configuration
		if tm.config.DeleteTag {
			return tm.deleteTag()
		}

		return tm.createTag()
	})
}

// fetchTags retrieves all tags and references from the remote repository.
// This ensures that tag operations have the most up-to-date information.
func (tm *TagManager) fetchTags() error {
	fmt.Printf("  • Fetching tags from remote... ")
	fetchCmd := exec.Command(gitcmd.CmdGit, gitcmd.FetchTagsArgs()...)
	fetchCmd.Stdout = os.Stdout
	fetchCmd.Stderr = os.Stderr

	if err := fetchCmd.Run(); err != nil {
		fmt.Println("❌ Failed")
		return fmt.Errorf("failed to fetch tags: %v", err)
	}

	fmt.Println("✅ Done")
	return nil
}

// deleteTag removes both local and remote tags with the specified name.
// It first deletes the local tag and then pushes the deletion to the remote.
func (tm *TagManager) deleteTag() error {
	fmt.Printf("\n  • Deleting tag: %s\n", tm.config.TagName)

	commands := []TagCommand{
		{gitcmd.CmdGit, gitcmd.TagDeleteArgs(tm.config.TagName), "Deleting local tag"},
		{gitcmd.CmdGit, gitcmd.DeleteRemoteTagArgs(tm.config.TagName), "Deleting remote tag"},
	}

	return tm.executeCommands(commands)
}

// createTag creates a new Git tag and pushes it to the remote repository.
// The tag can point to a specific commit if tag_reference is provided.
func (tm *TagManager) createTag() error {
	// Determine the commit to tag
	targetCommit, err := tm.resolveTargetCommit()
	if err != nil {
		return err
	}

	// Build the tag command arguments
	tagArgs := tm.buildTagArgs(targetCommit)

	// Create a human-readable description of the operation
	desc := tm.buildTagDescription(targetCommit)

	// Execute the tag creation and push commands
	commands := []TagCommand{
		{gitcmd.CmdGit, tagArgs, desc},
		{gitcmd.CmdGit, gitcmd.PushTagArgs(tm.config.TagName, true), "Pushing tag to remote"},
	}

	return tm.executeCommands(commands)
}

// resolveTargetCommit determines the exact commit that will be tagged.
// If tag_reference is not provided, it returns an empty string to tag the current commit.
func (tm *TagManager) resolveTargetCommit() (string, error) {
	// If no reference is specified, tag the current commit
	if tm.config.TagReference == "" {
		return "", nil
	}

	// Verify the reference is valid
	fmt.Printf("  • Verifying reference '%s'... ", tm.config.TagReference)
	verifyCmd := exec.Command(gitcmd.CmdGit, gitcmd.RevParseArgs(tm.config.TagReference)...)
	verifyCmd.Stderr = os.Stderr

	if err := verifyCmd.Run(); err != nil {
		fmt.Println("❌ Failed")
		return "", fmt.Errorf("invalid git reference '%s': %v", tm.config.TagReference, err)
	}
	fmt.Println("✅ Valid")

	// Get the full commit SHA for the reference
	fmt.Printf("  • Resolving commit for '%s'... ", tm.config.TagReference)
	revListCmd := exec.Command(gitcmd.CmdGit, gitcmd.RevListArgs(tm.config.TagReference)...)
	output, err := revListCmd.Output()
	if err != nil {
		fmt.Println("❌ Failed")
		return "", fmt.Errorf("failed to get commit SHA for '%s': %v", tm.config.TagReference, err)
	}

	commitSHA := strings.TrimSpace(string(output))
	fmt.Printf("✅ Found: %s\n", shortenCommitSHA(commitSHA))

	return commitSHA, nil
}

// shortenCommitSHA creates a shorter version of a commit SHA for display.
// It returns the first 8 characters of the commit SHA.
func shortenCommitSHA(sha string) string {
	if len(sha) > 8 {
		return sha[:8]
	}
	return sha
}

// buildTagArgs constructs the arguments for the git tag command.
// It handles different combinations of tag options based on the configuration.
func (tm *TagManager) buildTagArgs(targetCommit string) []string {
	// If we have a message, create an annotated tag
	if tm.config.TagMessage != "" {
		args := gitcmd.TagCreateAnnotatedArgs(tm.config.TagName, tm.config.TagMessage, true)
		// Insert target commit if specified (before the message)
		if targetCommit != "" {
			// Find the position of -m flag and insert commit before it
			for i, arg := range args {
				if arg == gitcmd.OptMessage {
					// Insert targetCommit before -m
					result := make([]string, 0, len(args)+1)
					result = append(result, args[:i]...)
					result = append(result, targetCommit)
					result = append(result, args[i:]...)
					return result
				}
			}
		}
		return args
	}

	// Simple tag without message
	args := gitcmd.TagCreateArgs(tm.config.TagName, true)
	if targetCommit != "" {
		args = append(args, targetCommit)
	}
	return args
}

// buildTagDescription creates a human-readable description of the tag operation.
// It includes details about the tag name and the target commit if applicable.
func (tm *TagManager) buildTagDescription(targetCommit string) string {
	desc := "Creating local tag " + tm.config.TagName

	// Add information about the target commit if available
	if tm.config.TagReference != "" && targetCommit != "" {
		if targetCommit != tm.config.TagReference {
			desc += fmt.Sprintf(" pointing to %s (%s)", tm.config.TagReference, shortenCommitSHA(targetCommit))
		} else {
			desc += fmt.Sprintf(" pointing to %s", shortenCommitSHA(targetCommit))
		}
	}

	return desc
}

// executeCommands runs a sequence of commands and handles the output formatting.
// It provides consistent error handling and status messages for each command.
func (tm *TagManager) executeCommands(commands []TagCommand) error {
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
