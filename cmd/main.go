package main

import (
	"log"

	"github.com/somaz94/go-git-commit-action/internal/config"
	"github.com/somaz94/go-git-commit-action/internal/git"
)

func main() {
	cfg := config.NewGitConfig()

	// 1. Git 커밋 실행
	if err := git.RunGitCommit(cfg); err != nil {
		log.Fatalf("Error executing git commands: %v", err)
	}

	// 2. PR 생성 (설정된 경우)
	if cfg.CreatePR {
		if err := git.CreatePullRequest(cfg); err != nil {
			log.Fatalf("Error creating pull request: %v", err)
		}
	}

	// 3. 태그 처리 (설정된 경우)
	if cfg.TagName != "" {
		if err := git.HandleGitTag(cfg); err != nil {
			log.Fatalf("Error handling git tag: %v", err)
		}
	}
}
