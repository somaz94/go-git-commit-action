package config

import "os"

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
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func NewGitConfig() *GitConfig {
	return &GitConfig{
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
		AutoBranch:         os.Getenv("INPUT_AUTO_BRANCH") != "false",
		PRTitle:            getEnvWithDefault("INPUT_PR_TITLE", "Auto PR by Go Git Commit Action"),
		PRBase:             getEnvWithDefault("INPUT_PR_BASE", "main"),
		PRBranch:           getEnvWithDefault("INPUT_PR_BRANCH", ""),
		DeleteSourceBranch: os.Getenv("INPUT_DELETE_SOURCE_BRANCH") == "true",
		GitHubToken:        os.Getenv("INPUT_GITHUB_TOKEN"),
	}
}
