package config

import (
	"os"
	"testing"
)

func TestNewGitConfig(t *testing.T) {
	// Setup test environment variables
	os.Setenv(EnvUserEmail, "test@example.com")
	os.Setenv(EnvUserName, "Test User")
	os.Setenv(EnvCommitMessage, "Test commit")
	os.Setenv(EnvBranch, "develop")
	defer func() {
		os.Unsetenv(EnvUserEmail)
		os.Unsetenv(EnvUserName)
		os.Unsetenv(EnvCommitMessage)
		os.Unsetenv(EnvBranch)
	}()

	cfg, err := NewGitConfig()
	if err != nil {
		t.Fatalf("NewGitConfig() error = %v", err)
	}

	if cfg.UserEmail != "test@example.com" {
		t.Errorf("UserEmail = %v, want test@example.com", cfg.UserEmail)
	}
	if cfg.UserName != "Test User" {
		t.Errorf("UserName = %v, want Test User", cfg.UserName)
	}
	if cfg.CommitMessage != "Test commit" {
		t.Errorf("CommitMessage = %v, want Test commit", cfg.CommitMessage)
	}
	if cfg.Branch != "develop" {
		t.Errorf("Branch = %v, want develop", cfg.Branch)
	}
}

func TestGitConfig_Defaults(t *testing.T) {
	// Set only required fields (user email and name are required but validation doesn't check them in NewGitConfig)
	os.Setenv(EnvUserEmail, "test@example.com")
	os.Setenv(EnvUserName, "Test User")
	defer func() {
		os.Unsetenv(EnvUserEmail)
		os.Unsetenv(EnvUserName)
	}()

	cfg, err := NewGitConfig()
	if err != nil {
		t.Fatalf("NewGitConfig() error = %v", err)
	}

	// Check defaults
	if cfg.CommitMessage != DefaultCommitMessage {
		t.Errorf("CommitMessage = %v, want %v", cfg.CommitMessage, DefaultCommitMessage)
	}
	if cfg.Branch != DefaultBranch {
		t.Errorf("Branch = %v, want %v", cfg.Branch, DefaultBranch)
	}
	if cfg.RepoPath != DefaultRepoPath {
		t.Errorf("RepoPath = %v, want %v", cfg.RepoPath, DefaultRepoPath)
	}
	if cfg.FilePattern != DefaultFilePattern {
		t.Errorf("FilePattern = %v, want %v", cfg.FilePattern, DefaultFilePattern)
	}
	if cfg.SkipIfEmpty != DefaultSkipIfEmpty {
		t.Errorf("SkipIfEmpty = %v, want %v", cfg.SkipIfEmpty, DefaultSkipIfEmpty)
	}
}

func TestGitConfig_ValidatePR(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*GitConfig)
		wantErr   bool
	}{
		{
			name: "valid PR config with manual branch",
			setupFunc: func(c *GitConfig) {
				c.CreatePR = true
				c.AutoBranch = false
				c.PRBranch = "feature"
				c.PRBase = "main"
				c.GitHubToken = "token"
			},
			wantErr: false,
		},
		{
			name: "valid PR config with auto branch",
			setupFunc: func(c *GitConfig) {
				c.CreatePR = true
				c.AutoBranch = true
				c.PRBase = "main"
				c.GitHubToken = "token"
			},
			wantErr: false,
		},
		{
			name: "missing PR branch when auto_branch is false",
			setupFunc: func(c *GitConfig) {
				c.CreatePR = true
				c.AutoBranch = false
				c.PRBase = "main"
				c.GitHubToken = "token"
			},
			wantErr: true,
		},
		{
			name: "missing PR base",
			setupFunc: func(c *GitConfig) {
				c.CreatePR = true
				c.AutoBranch = false
				c.PRBranch = "feature"
				c.GitHubToken = "token"
			},
			wantErr: true,
		},
		{
			name: "missing GitHub token",
			setupFunc: func(c *GitConfig) {
				c.CreatePR = true
				c.AutoBranch = false
				c.PRBranch = "feature"
				c.PRBase = "main"
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &GitConfig{}
			tt.setupFunc(cfg)
			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGitConfig_ValidateTag(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*GitConfig)
		wantErr   bool
	}{
		{
			name: "valid tag creation",
			setupFunc: func(c *GitConfig) {
				c.TagName = "v1.0.0"
				c.DeleteTag = false
			},
			wantErr: false,
		},
		{
			name: "valid tag deletion",
			setupFunc: func(c *GitConfig) {
				c.TagName = "v1.0.0"
				c.DeleteTag = true
			},
			wantErr: false,
		},
		{
			name: "invalid: tag_reference with delete_tag",
			setupFunc: func(c *GitConfig) {
				c.TagName = "v1.0.0"
				c.DeleteTag = true
				c.TagReference = "main"
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &GitConfig{}
			tt.setupFunc(cfg)
			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetBoolEnv(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue bool
		want         bool
	}{
		{"empty string uses default", "", false, false},
		{"true string", "true", false, true},
		{"false string", "false", true, false},
		{"1 is true", "1", false, true},
		{"0 is false", "0", true, false},
		{"t is true", "t", false, true},
		{"f is false", "f", true, false},
		{"TRUE is true", "TRUE", false, true},
		{"FALSE is false", "FALSE", true, false},
		{"invalid string uses default", "invalid", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := "TEST_BOOL_ENV"
			if tt.envValue != "" {
				os.Setenv(key, tt.envValue)
				defer os.Unsetenv(key)
			}
			got := getBoolEnv(key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getBoolEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetIntEnv(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue int
		want         int
	}{
		{"empty string uses default", "", 10, 10},
		{"valid integer", "42", 10, 42},
		{"zero", "0", 10, 0},
		{"negative", "-5", 10, -5},
		{"invalid string uses default", "abc", 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := "TEST_INT_ENV"
			if tt.envValue != "" {
				os.Setenv(key, tt.envValue)
				defer os.Unsetenv(key)
			}
			got := getIntEnv(key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getIntEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseLabels(t *testing.T) {
	tests := []struct {
		name      string
		labelsStr string
		want      []string
	}{
		{
			name:      "empty string",
			labelsStr: "",
			want:      nil,
		},
		{
			name:      "single label",
			labelsStr: "bug",
			want:      []string{"bug"},
		},
		{
			name:      "multiple labels",
			labelsStr: "bug,enhancement,documentation",
			want:      []string{"bug", "enhancement", "documentation"},
		},
		{
			name:      "labels with spaces",
			labelsStr: "bug, enhancement, documentation",
			want:      []string{"bug", "enhancement", "documentation"},
		},
		{
			name:      "labels with extra spaces",
			labelsStr: " bug , enhancement , documentation ",
			want:      []string{"bug", "enhancement", "documentation"},
		},
		{
			name:      "empty labels filtered out",
			labelsStr: "bug,,enhancement",
			want:      []string{"bug", "enhancement"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLabels(tt.labelsStr)
			if len(got) != len(tt.want) {
				t.Errorf("parseLabels() length = %v, want %v", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("parseLabels()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestGetGitHubToken(t *testing.T) {
	tests := []struct {
		name           string
		inputToken     string
		githubToken    string
		expectedResult string
	}{
		{
			name:           "INPUT_GITHUB_TOKEN takes priority",
			inputToken:     "input-token",
			githubToken:    "github-token",
			expectedResult: "input-token",
		},
		{
			name:           "falls back to GITHUB_TOKEN",
			inputToken:     "",
			githubToken:    "github-token",
			expectedResult: "github-token",
		},
		{
			name:           "returns empty if both are empty",
			inputToken:     "",
			githubToken:    "",
			expectedResult: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Unsetenv(EnvGitHubToken)
			os.Unsetenv("GITHUB_TOKEN")

			// Setup test environment
			if tt.inputToken != "" {
				os.Setenv(EnvGitHubToken, tt.inputToken)
			}
			if tt.githubToken != "" {
				os.Setenv("GITHUB_TOKEN", tt.githubToken)
			}

			got := getGitHubToken()
			if got != tt.expectedResult {
				t.Errorf("getGitHubToken() = %v, want %v", got, tt.expectedResult)
			}

			// Cleanup
			os.Unsetenv(EnvGitHubToken)
			os.Unsetenv("GITHUB_TOKEN")
		})
	}
}

func TestGetEnvWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue string
		want         string
	}{
		{"empty uses default", "", "default", "default"},
		{"set value is used", "custom", "default", "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := "TEST_ENV_VAR"
			if tt.envValue != "" {
				os.Setenv(key, tt.envValue)
				defer os.Unsetenv(key)
			} else {
				os.Unsetenv(key)
			}
			got := getEnvWithDefault(key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getEnvWithDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}
