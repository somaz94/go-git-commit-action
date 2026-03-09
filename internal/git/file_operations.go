package git

import (
	"github.com/somaz94/go-git-commit-action/internal/config"
	"github.com/somaz94/go-git-commit-action/internal/git/shared"
	"github.com/somaz94/go-git-commit-action/internal/gitcmd"
)

// StageFiles adds the specified files to the Git staging area.
// It delegates to the shared package to avoid duplication with pr/branch.go.
func StageFiles(filePattern string) error {
	return shared.StageFiles(filePattern)
}

// CommitAndPush commits the staged changes and pushes them to the remote branch.
func CommitAndPush(cfg *config.GitConfig, branch string) error {
	commitPushCommands := []Command{
		{gitcmd.CmdGit, gitcmd.CommitArgs(cfg.CommitMessage), "Committing changes"},
		{gitcmd.CmdGit, gitcmd.PushUpstreamArgs(gitcmd.RefOrigin, branch), "Pushing changes"},
	}

	return ExecuteCommandBatch(commitPushCommands, "")
}
