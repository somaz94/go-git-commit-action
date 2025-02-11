package main

import (
	"log"

	"github.com/somaz94/go-git-commit-action/internal/config"
	"github.com/somaz94/go-git-commit-action/internal/git"
)

func main() {
	cfg := config.NewGitConfig()

	if err := git.RunGitCommit(cfg); err != nil {
		log.Fatalf("Error executing git commands: %v", err)
	}

	if cfg.TagName != "" {
		if err := git.HandleGitTag(cfg); err != nil {
			log.Fatalf("Error handling git tag: %v", err)
		}
	}
}
