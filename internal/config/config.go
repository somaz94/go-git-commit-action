package config

import (
	"fmt"
	"os"
	"strings"
)

type GitConfig struct {
	UserEmail          string
	UserName           string
	CommitMessage      string
	Branch             string
	RepoPath           string
	FilePattern        string
	TagName            string
	TagMessage         string
	DeleteTag          bool
	TagReference       string
	CreatePR           bool
	AutoBranch         bool
	PRTitle            string
	PRBase             string
	PRBranch           string
	DeleteSourceBranch bool
	GitHubToken        string
	PRLabels           []string
	PRBody             string
	SkipIfEmpty        bool
	PRClosed           bool
	PRDryRun           bool
	Debug              bool
	Timeout            int
	RetryCount         int
}

func (c *GitConfig) Validate() error {
	if c.CreatePR {
		if !c.AutoBranch && c.PRBranch == "" {
			return fmt.Errorf("pr_branch must be specified when auto_branch is false and create_pr is true")
		}
		if c.PRBase == "" {
			return fmt.Errorf("pr_base must be specified when create_pr is true")
		}
		if c.GitHubToken == "" {
			return fmt.Errorf("github_token must be specified when create_pr is true")
		}
	}
	return nil
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func splitAndTrim(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func NewGitConfig() (*GitConfig, error) {
	cfg := &GitConfig{
		UserEmail:          os.Getenv("INPUT_USER_EMAIL"),
		UserName:           os.Getenv("INPUT_USER_NAME"),
		CommitMessage:      getEnvWithDefault("INPUT_COMMIT_MESSAGE", "Auto commit by Go Git Commit Action"),
		Branch:             getEnvWithDefault("INPUT_BRANCH", "main"),
		RepoPath:           getEnvWithDefault("INPUT_REPOSITORY_PATH", "."),
		FilePattern:        getEnvWithDefault("INPUT_FILE_PATTERN", "."),
		TagName:            os.Getenv("INPUT_TAG_NAME"),
		TagMessage:         os.Getenv("INPUT_TAG_MESSAGE"),
		DeleteTag:          os.Getenv("INPUT_DELETE_TAG") == "true",
		TagReference:       os.Getenv("INPUT_TAG_REFERENCE"),
		CreatePR:           os.Getenv("INPUT_CREATE_PR") == "true",
		AutoBranch:         os.Getenv("INPUT_AUTO_BRANCH") == "true",
		PRTitle:            getEnvWithDefault("INPUT_PR_TITLE", ""),
		PRBase:             getEnvWithDefault("INPUT_PR_BASE", "main"),
		PRBranch:           getEnvWithDefault("INPUT_PR_BRANCH", ""),
		DeleteSourceBranch: os.Getenv("INPUT_DELETE_SOURCE_BRANCH") == "true",
		GitHubToken:        os.Getenv("INPUT_GITHUB_TOKEN"),
		PRLabels:           splitAndTrim(os.Getenv("INPUT_PR_LABELS")),
		PRBody:             os.Getenv("INPUT_PR_BODY"),
		SkipIfEmpty:        os.Getenv("INPUT_SKIP_IF_EMPTY") == "true",
		PRClosed:           os.Getenv("INPUT_PR_CLOSED") == "true",
		PRDryRun:           os.Getenv("INPUT_PR_DRY_RUN") == "true",
		Debug:              os.Getenv("INPUT_DEBUG") == "true",
		Timeout:            30,
		RetryCount:         3,
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %v", err)
	}

	return cfg, nil
}
