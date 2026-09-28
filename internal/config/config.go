package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/somaz94/go-git-commit-action/internal/errors"
)

// Input environment variable names
const (
	// User information
	EnvUserEmail = "INPUT_USER_EMAIL"
	EnvUserName  = "INPUT_USER_NAME"

	// Commit settings
	EnvCommitMessage = "INPUT_COMMIT_MESSAGE"
	EnvBranch        = "INPUT_BRANCH"
	EnvRepoPath      = "INPUT_REPOSITORY_PATH"
	EnvFilePattern   = "INPUT_FILE_PATTERN"
	EnvSkipIfEmpty   = "INPUT_SKIP_IF_EMPTY"

	// Tag settings
	EnvTagName      = "INPUT_TAG_NAME"
	EnvTagMessage   = "INPUT_TAG_MESSAGE"
	EnvDeleteTag    = "INPUT_DELETE_TAG"
	EnvTagReference = "INPUT_TAG_REFERENCE"

	// Pull request settings
	EnvCreatePR           = "INPUT_CREATE_PR"
	EnvAutoBranch         = "INPUT_AUTO_BRANCH"
	EnvPRTitle            = "INPUT_PR_TITLE"
	EnvPRBase             = "INPUT_PR_BASE"
	EnvPRBranch           = "INPUT_PR_BRANCH"
	EnvDeleteSourceBranch = "INPUT_DELETE_SOURCE_BRANCH"
	EnvGitHubToken        = "INPUT_GITHUB_TOKEN"
	EnvPRLabels           = "INPUT_PR_LABELS"
	EnvPRBody             = "INPUT_PR_BODY"
	EnvPRClosed           = "INPUT_PR_CLOSED"
	EnvPRDraft            = "INPUT_PR_DRAFT"
	EnvPRReviewers        = "INPUT_PR_REVIEWERS"
	EnvPRAssignees        = "INPUT_PR_ASSIGNEES"
	EnvPRDryRun           = "INPUT_PR_DRY_RUN"

	// Operational settings
	EnvDebug      = "INPUT_DEBUG"
	EnvTimeout    = "INPUT_TIMEOUT"
	EnvRetryCount = "INPUT_RETRY_COUNT"
)

// Default values for configuration parameters
const (
	DefaultCommitMessage = "Auto commit by Go Git Commit Action"
	DefaultBranch        = "main"
	DefaultRepoPath      = "."
	DefaultFilePattern   = "."
	DefaultSkipIfEmpty   = false
	DefaultDeleteTag     = false
	DefaultCreatePR      = false
	DefaultAutoBranch    = false
	DefaultPRTitle       = ""
	DefaultPRBase        = "main"
	DefaultPRBranch      = ""
	DefaultDeleteSource  = false
	DefaultPRClosed      = false
	DefaultPRDraft       = false
	DefaultPRDryRun      = false
	DefaultDebug         = false
	DefaultTimeout       = 30
	DefaultRetryCount    = 3
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
	PRDraft            bool
	PRReviewers        []string
	PRAssignees        []string
	PRDryRun           bool

	// Operational settings
	Debug      bool
	Timeout    int
	RetryCount int
}

// Validate checks that the configuration is valid for the requested operations.
// It verifies that required fields are set based on the actions being performed.
func (c *GitConfig) Validate() error {
	if c.CreatePR {
		if !c.AutoBranch && c.PRBranch == "" {
			return errors.NewConfigError("pr_branch", "must be specified when auto_branch is false and create_pr is true")
		}
		if c.PRBase == "" {
			return errors.NewConfigError("pr_base", "must be specified when create_pr is true")
		}
		if c.GitHubToken == "" {
			return errors.NewConfigError("github_token", "must be specified when create_pr is true")
		}
	}

	if c.TagName != "" && c.DeleteTag {
		if c.TagReference != "" {
			return errors.NewConfigError("tag_reference", "cannot be used with delete_tag")
		}
	}

	return nil
}

// NewGitConfig creates a new GitConfig instance by reading environment variables.
// It applies default values where applicable and validates the configuration.
func NewGitConfig() (*GitConfig, error) {
	cfg := &GitConfig{
		// User information (no defaults)
		UserEmail: os.Getenv(EnvUserEmail),
		UserName:  os.Getenv(EnvUserName),

		// Commit settings
		CommitMessage: getEnvWithDefault(EnvCommitMessage, DefaultCommitMessage),
		Branch:        getEnvWithDefault(EnvBranch, DefaultBranch),
		RepoPath:      getEnvWithDefault(EnvRepoPath, DefaultRepoPath),
		FilePattern:   getEnvWithDefault(EnvFilePattern, DefaultFilePattern),
		SkipIfEmpty:   getBoolEnv(EnvSkipIfEmpty, DefaultSkipIfEmpty),

		// Tag settings
		TagName:      os.Getenv(EnvTagName),
		TagMessage:   os.Getenv(EnvTagMessage),
		DeleteTag:    getBoolEnv(EnvDeleteTag, DefaultDeleteTag),
		TagReference: os.Getenv(EnvTagReference),

		// Pull request settings
		CreatePR:           getBoolEnv(EnvCreatePR, DefaultCreatePR),
		AutoBranch:         getBoolEnv(EnvAutoBranch, DefaultAutoBranch),
		PRTitle:            getEnvWithDefault(EnvPRTitle, DefaultPRTitle),
		PRBase:             getEnvWithDefault(EnvPRBase, DefaultPRBase),
		PRBranch:           getEnvWithDefault(EnvPRBranch, DefaultPRBranch),
		DeleteSourceBranch: getBoolEnv(EnvDeleteSourceBranch, DefaultDeleteSource),
		GitHubToken:        getGitHubToken(),
		PRLabels:           parseCommaSeparated(os.Getenv(EnvPRLabels)),
		PRBody:             os.Getenv(EnvPRBody),
		PRClosed:           getBoolEnv(EnvPRClosed, DefaultPRClosed),
		PRDraft:            getBoolEnv(EnvPRDraft, DefaultPRDraft),
		PRReviewers:        parseCommaSeparated(os.Getenv(EnvPRReviewers)),
		PRAssignees:        parseCommaSeparated(os.Getenv(EnvPRAssignees)),
		PRDryRun:           getBoolEnv(EnvPRDryRun, DefaultPRDryRun),

		// Operational settings
		Debug:      getBoolEnv(EnvDebug, DefaultDebug),
		Timeout:    getIntEnv(EnvTimeout, DefaultTimeout),
		RetryCount: getIntEnv(EnvRetryCount, DefaultRetryCount),
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// getEnvWithDefault retrieves an environment variable value or returns
// the specified default value if the variable is not set or empty.
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getBoolEnv parses key case-insensitively; unset or unparsable values
// (e.g. "yes", "on") silently fall back to defaultValue.
func getBoolEnv(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	b, err := strconv.ParseBool(strings.ToLower(value))
	if err != nil {
		return defaultValue
	}
	return b
}

// getIntEnv retrieves an integer environment variable value.
// It parses the string value to an integer, returning the default value
// if the variable is not set, empty, or cannot be parsed as an integer.
func getIntEnv(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return i
}

// parseCommaSeparated converts a comma-separated string into a slice of strings.
// It trims whitespace from each item and filters out empty ones.
// Used for labels, reviewers, assignees, and other comma-delimited inputs.
func parseCommaSeparated(labelsStr string) []string {
	if labelsStr == "" {
		return nil
	}

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

// getGitHubToken prefers INPUT_GITHUB_TOKEN and falls back to GITHUB_TOKEN.
// Under the action both come from the github_token input, which has no default.
func getGitHubToken() string {
	if token := os.Getenv(EnvGitHubToken); token != "" {
		return token
	}

	return os.Getenv("GITHUB_TOKEN")
}
