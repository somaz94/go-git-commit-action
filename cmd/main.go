package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/somaz94/go-git-commit-action/internal/config"
	"github.com/somaz94/go-git-commit-action/internal/git"
	"github.com/somaz94/go-git-commit-action/internal/output"
)

func main() {
	// Handle OS signals for graceful shutdown
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

	result := output.NewResult()

	if err := git.RunGitCommit(ctx, cfg, result); err != nil {
		log.Fatalf("Error executing git commands: %v", err)
	}

	if cfg.TagName != "" {
		tagManager := git.NewTagManager(cfg)
		if err := tagManager.HandleGitTag(ctx, result); err != nil {
			log.Fatalf("Error handling git tag: %v", err)
		}
	}

	if err := result.WriteToGitHubOutput(); err != nil {
		log.Printf("[WARN] Failed to write action outputs: %v", err)
	}
}
