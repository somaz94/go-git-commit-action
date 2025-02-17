package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/somaz94/go-git-commit-action/internal/config"
	"github.com/somaz94/go-git-commit-action/internal/git"
)

func main() {
	// 시그널 처리
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	cfg, err := config.NewGitConfig()
	if err != nil {
		log.Fatalf("Failed to initialize configuration: %v", err)
	}

	if err := git.RunGitCommit(cfg); err != nil {
		log.Fatalf("Error executing git commands: %v", err)
	}

	if cfg.TagName != "" {
		tagManager := git.NewTagManager(cfg)
		if err := tagManager.HandleGitTag(ctx); err != nil {
			log.Fatalf("Error handling git tag: %v", err)
		}
	}
}
