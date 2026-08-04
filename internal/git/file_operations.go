package git

import (
	"github.com/somaz94/go-git-commit-action/internal/git/shared"
	"github.com/somaz94/go-git-commit-action/internal/gitcmd"
)

// StageFiles adds the specified files to the Git staging area.
// It delegates to the shared package to avoid duplication with pr/branch.go.
func StageFiles(r gitcmd.Runner, filePattern string) error {
	return shared.StageFiles(r, filePattern)
}
