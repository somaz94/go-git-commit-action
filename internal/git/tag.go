package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/somaz94/go-git-commit-action/internal/config"
)

func HandleGitTag(config *config.GitConfig) error {
	fmt.Println("\n🏷️  Handling Git Tag:")

	// Fetch all tags and refs
	fetchCmd := exec.Command("git", "fetch", "--tags", "--force", "origin")
	if err := fetchCmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch tags: %v", err)
	}

	if config.DeleteTag {
		// Delete tag
		commands := []struct {
			name string
			args []string
			desc string
		}{
			{"git", []string{"tag", "-d", config.TagName}, "Deleting local tag"},
			{"git", []string{"push", "origin", ":refs/tags/" + config.TagName}, "Deleting remote tag"},
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
		var targetCommit string
		if config.TagReference != "" {
			// Check if reference exists
			cmd := exec.Command("git", "rev-parse", "--verify", config.TagReference)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("invalid git reference '%s': %v", config.TagReference, err)
			}

			// Get commit SHA for the reference
			cmd = exec.Command("git", "rev-list", "-n1", config.TagReference)
			output, err := cmd.Output()
			if err != nil {
				return fmt.Errorf("failed to get commit SHA for '%s': %v", config.TagReference, err)
			}
			targetCommit = strings.TrimSpace(string(output))
		}

		// Create tag
		var tagArgs []string
		if config.TagMessage != "" {
			if targetCommit != "" {
				tagArgs = []string{"tag", "-f", "-a", config.TagName, targetCommit, "-m", config.TagMessage}
			} else {
				tagArgs = []string{"tag", "-f", "-a", config.TagName, "-m", config.TagMessage}
			}
		} else {
			if targetCommit != "" {
				tagArgs = []string{"tag", "-f", config.TagName, targetCommit}
			} else {
				tagArgs = []string{"tag", "-f", config.TagName}
			}
		}

		// Create description message
		desc := "Creating local tag " + config.TagName
		if config.TagReference != "" {
			if targetCommit != config.TagReference {
				desc += fmt.Sprintf(" pointing to %s (%s)", config.TagReference, targetCommit[:8])
			} else {
				desc += fmt.Sprintf(" pointing to %s", targetCommit[:8])
			}
		}

		commands := []struct {
			name string
			args []string
			desc string
		}{
			{"git", tagArgs, desc},
			{"git", []string{"push", "-f", "origin", config.TagName}, "Pushing tag to remote"},
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
	}

	return nil
}
