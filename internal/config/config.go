package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// GitConfig holds all configuration parameters for the Git commit action.
// It encapsulates user settings, commit options, tag settings, PR configuration,
// and operational parameters.
type GitConfig struct {
	// User information
	UserEmail string
	UserName  string

	// Commit settings
	CommitMessage string
	Branch        string
	RepoPath      string
	FilePattern   string
	SkipIfEmpty   bool

	// Tag settings
	TagName      string
	TagMessage   string
	DeleteTag    bool
	TagReference string

	// Pull request settings
	CreatePR           bool
	AutoBranch         bool
	PRTitle            string
	PRBase             string
	PRBranch           string
	DeleteSourceBranch bool
	GitHubToken        string
	PRLabels           []string
	PRBody             string
	PRClosed           bool
	PRDryRun           bool

	// Operational settings
	Debug      bool
	Timeout    int
	RetryCount int
}

// Validate checks that the configuration is valid for the requested operations.
// It verifies that required fields are set based on the actions being performed.
func (c *GitConfig) Validate() error {
	// Validate pull request configuration
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

	// Validate tag configuration
	if c.TagName != "" && c.DeleteTag {
		if c.TagReference != "" {
			return fmt.Errorf("tag_reference cannot be used with delete_tag")
		}
	}

	return nil
}

// NewGitConfig creates a new GitConfig instance by reading environment variables.
// It applies default values where applicable and validates the configuration.
func NewGitConfig() (*GitConfig, error) {
	cfg := &GitConfig{
		// User information with no defaults
		UserEmail: os.Getenv("INPUT_USER_EMAIL"),
		UserName:  os.Getenv("INPUT_USER_NAME"),

		// Commit settings with defaults
		CommitMessage: getEnvWithDefault("INPUT_COMMIT_MESSAGE", "Auto commit by Go Git Commit Action"),
		Branch:        getEnvWithDefault("INPUT_BRANCH", "main"),
		RepoPath:      getEnvWithDefault("INPUT_REPOSITORY_PATH", "."),
		FilePattern:   getEnvWithDefault("INPUT_FILE_PATTERN", "."),
		SkipIfEmpty:   getEnvBool("INPUT_SKIP_IF_EMPTY", false),

		// Tag settings
		TagName:      os.Getenv("INPUT_TAG_NAME"),
		TagMessage:   os.Getenv("INPUT_TAG_MESSAGE"),
		DeleteTag:    getEnvBool("INPUT_DELETE_TAG", false),
		TagReference: os.Getenv("INPUT_TAG_REFERENCE"),

		// Pull request settings
		CreatePR:           getEnvBool("INPUT_CREATE_PR", false),
		AutoBranch:         getEnvBool("INPUT_AUTO_BRANCH", false),
		PRTitle:            getEnvWithDefault("INPUT_PR_TITLE", ""),
		PRBase:             getEnvWithDefault("INPUT_PR_BASE", "main"),
		PRBranch:           getEnvWithDefault("INPUT_PR_BRANCH", ""),
		DeleteSourceBranch: getEnvBool("INPUT_DELETE_SOURCE_BRANCH", false),
		GitHubToken:        os.Getenv("INPUT_GITHUB_TOKEN"),
		PRLabels:           parseLabels(os.Getenv("INPUT_PR_LABELS")),
		PRBody:             os.Getenv("INPUT_PR_BODY"),
		PRClosed:           getEnvBool("INPUT_PR_CLOSED", false),
		PRDryRun:           getEnvBool("INPUT_PR_DRY_RUN", false),

		// Operational settings
		Debug:      getEnvBool("INPUT_DEBUG", false),
		Timeout:    getEnvInt("INPUT_TIMEOUT", 30),
		RetryCount: getEnvInt("INPUT_RETRY_COUNT", 3),
	}

	// Validate the configuration after setting all values
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %v", err)
	}

	return cfg, nil
}

// getEnvWithDefault retrieves an environment variable or returns a default value if not set.
// This helper function simplifies the handling of environment variables with default values.
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool retrieves a boolean environment variable or returns a default value if not set.
// It converts strings like "true", "yes", "1" to true, and everything else to false.
func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return value == "true" || value == "yes" || value == "1"
}

// getEnvInt retrieves an integer environment variable or returns a default value if not set.
// It handles conversion from string to int and falls back to the default on any error.
func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}

// parseLabels converts a comma-separated string of labels into a slice of strings.
// It trims whitespace from each label and filters out empty ones.
func parseLabels(labelsStr string) []string {
	if labelsStr == "" {
		return nil
	}

	// Split by comma and process each part
	parts := strings.Split(labelsStr, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
