package main

import (
	"log"

	"github.com/somaz94/go-git-commit-action/internal/config"
	"github.com/somaz94/go-git-commit-action/internal/git"
)

func main() {
	cfg := config.NewGitConfig()

	// 1. Git 환경 설정 및 기본 브랜치 처리
	if err := git.SetupGitEnvironment(cfg); err != nil {
		log.Fatalf("Error setting up git environment: %v", err)
	}

	// 2. 일반 커밋 처리
	if err := git.RunGitCommit(cfg); err != nil {
		log.Fatalf("Error executing git commands: %v", err)
	}

	// 3. PR 생성 처리 (설정된 경우)
	if cfg.CreatePR {
		if err := git.HandlePullRequest(cfg); err != nil {
			log.Fatalf("Error creating pull request: %v", err)
		}
	}

	// 4. 태그 처리 (설정된 경우)
	if cfg.TagName != "" {
		if err := git.HandleGitTag(cfg); err != nil {
			log.Fatalf("Error handling git tag: %v", err)
		}
	}
}
