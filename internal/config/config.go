package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
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
		PRLabels:           parseLabels(os.Getenv(EnvPRLabels)),
		PRBody:             os.Getenv(EnvPRBody),
		PRClosed:           getBoolEnv(EnvPRClosed, DefaultPRClosed),
		PRDryRun:           getBoolEnv(EnvPRDryRun, DefaultPRDryRun),

		// Operational settings
		Debug:      getBoolEnv(EnvDebug, DefaultDebug),
		Timeout:    getIntEnv(EnvTimeout, DefaultTimeout),
		RetryCount: getIntEnv(EnvRetryCount, DefaultRetryCount),
	}

	// Validate the configuration after setting all values
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %v", err)
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

// getBoolEnv retrieves a boolean environment variable value.
// It parses the string value to a boolean, returning the default value
// if the variable is not set, empty, or cannot be parsed as a boolean.
// Accepts: true, false, 1, 0, t, f, T, F, TRUE, FALSE, True, False
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

// getGitHubToken retrieves the GitHub token from various sources.
// Priority order:
// 1. INPUT_GITHUB_TOKEN (user-provided token via action input)
// 2. GITHUB_TOKEN (automatically available in GitHub Actions)
// This allows the action to work without explicit token configuration in most cases.
func getGitHubToken() string {
	// First check if user explicitly provided a token
	if token := os.Getenv(EnvGitHubToken); token != "" {
		return token
	}

	// Fall back to the automatically available GITHUB_TOKEN
	return os.Getenv("GITHUB_TOKEN")
}
