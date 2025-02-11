package main

import (
	"log"

	"github.com/somaz94/go-git-commit-action/internal/config"
	"github.com/somaz94/go-git-commit-action/internal/git"
)

func main() {
	cfg := config.NewGitConfig()

	// 일반 커밋 처리
	if err := git.RunGitCommit(cfg); err != nil {
		log.Fatalf("Error executing git commands: %v", err)
	}

	// PR 생성 처리 (별도 분리)
	if cfg.CreatePR {
		if err := git.HandlePullRequest(cfg); err != nil {
			log.Fatalf("Error creating pull request: %v", err)
		}
	}

	// 태그 처리
	if cfg.TagName != "" {
		if err := git.HandleGitTag(cfg); err != nil {
			log.Fatalf("Error handling git tag: %v", err)
		}
	}
}
