package errors

import (
	"errors"
	"strings"
	"testing"
)

func TestGitError_Error(t *testing.T) {
	tests := []struct {
		name     string
		gitError *GitError
		want     string
	}{
		{
			name: "error with path",
			gitError: &GitError{
				Op:   "commit",
				Path: "/path/to/repo",
				Err:  errors.New("failed to commit"),
			},
			want: "commit /path/to/repo: failed to commit",
		},
		{
			name: "error without path",
			gitError: &GitError{
				Op:  "push",
				Err: errors.New("failed to push"),
			},
			want: "push: failed to push",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.gitError.Error()
			if got != tt.want {
				t.Errorf("GitError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGitError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	gitErr := &GitError{
		Op:  "test",
		Err: originalErr,
	}

	unwrapped := gitErr.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("GitError.Unwrap() = %v, want %v", unwrapped, originalErr)
	}
}

func TestNew(t *testing.T) {
	err := errors.New("test error")
	gitErr := New("operation", err)

	if gitErr.Op != "operation" {
		t.Errorf("New() Op = %v, want %v", gitErr.Op, "operation")
	}
	if gitErr.Err != err {
		t.Errorf("New() Err = %v, want %v", gitErr.Err, err)
	}
	if gitErr.Path != "" {
		t.Errorf("New() Path = %v, want empty string", gitErr.Path)
	}
}

func TestNewWithPath(t *testing.T) {
	err := errors.New("test error")
	gitErr := NewWithPath("operation", "/test/path", err)

	if gitErr.Op != "operation" {
		t.Errorf("NewWithPath() Op = %v, want %v", gitErr.Op, "operation")
	}
	if gitErr.Path != "/test/path" {
		t.Errorf("NewWithPath() Path = %v, want %v", gitErr.Path, "/test/path")
	}
	if gitErr.Err != err {
		t.Errorf("NewWithPath() Err = %v, want %v", gitErr.Err, err)
	}
}

func TestConfigError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ConfigError
		expected string
	}{
		{
			name: "error with field",
			err: &ConfigError{
				Field:   "pr_branch",
				Message: "must be specified when auto_branch is false",
			},
			expected: "configuration error in pr_branch: must be specified when auto_branch is false",
		},
		{
			name: "error without field",
			err: &ConfigError{
				Message: "invalid configuration",
			},
			expected: "configuration error: invalid configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("ConfigError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewConfigError(t *testing.T) {
	configErr := NewConfigError("test_field", "test message")

	if configErr.Field != "test_field" {
		t.Errorf("NewConfigError() Field = %v, want %v", configErr.Field, "test_field")
	}
	if configErr.Message != "test message" {
		t.Errorf("NewConfigError() Message = %v, want %v", configErr.Message, "test message")
	}
}

func TestNewConfig(t *testing.T) {
	configErr := NewConfig("test message")

	if configErr.Field != "" {
		t.Errorf("NewConfig() Field = %v, want empty", configErr.Field)
	}
	if configErr.Message != "test message" {
		t.Errorf("NewConfig() Message = %v, want %v", configErr.Message, "test message")
	}
}

func TestRetryError_Error(t *testing.T) {
	originalErr := errors.New("connection timeout")
	retryErr := &RetryError{
		Message:  "operation failed after retries",
		Attempts: 3,
		LastErr:  originalErr,
	}

	expected := "operation failed after retries (failed after 3 attempts): connection timeout"
	got := retryErr.Error()

	if got != expected {
		t.Errorf("RetryError.Error() = %v, want %v", got, expected)
	}
}

func TestRetryError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	retryErr := &RetryError{
		Message:  "retry failed",
		Attempts: 5,
		LastErr:  originalErr,
	}

	unwrapped := retryErr.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("RetryError.Unwrap() = %v, want %v", unwrapped, originalErr)
	}
}

func TestNewWithContext(t *testing.T) {
	originalErr := errors.New("test error")
	retryErr := NewWithContext("test operation", 3, originalErr)

	if retryErr.Message != "test operation" {
		t.Errorf("NewWithContext() Message = %v, want %v", retryErr.Message, "test operation")
	}
	if retryErr.Attempts != 3 {
		t.Errorf("NewWithContext() Attempts = %v, want 3", retryErr.Attempts)
	}
	if retryErr.LastErr != originalErr {
		t.Errorf("NewWithContext() LastErr = %v, want %v", retryErr.LastErr, originalErr)
	}
}

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		apiError *APIError
		want     string
	}{
		{
			name: "error with status code",
			apiError: &APIError{
				Operation:  "create PR",
				StatusCode: 404,
				Message:    "not found",
			},
			want: "GitHub API error (create PR) [404]: not found",
		},
		{
			name: "error without status code",
			apiError: &APIError{
				Operation: "add labels",
				Message:   "failed to add labels",
			},
			want: "GitHub API error (add labels): failed to add labels",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.apiError.Error()
			if got != tt.want {
				t.Errorf("APIError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewAPIError(t *testing.T) {
	apiErr := NewAPIError("test operation", "test message")

	if apiErr.Operation != "test operation" {
		t.Errorf("NewAPIError() Operation = %v, want %v", apiErr.Operation, "test operation")
	}
	if apiErr.Message != "test message" {
		t.Errorf("NewAPIError() Message = %v, want %v", apiErr.Message, "test message")
	}
	if apiErr.StatusCode != 0 {
		t.Errorf("NewAPIError() StatusCode = %v, want 0", apiErr.StatusCode)
	}
}

func TestNewAPIErrorWithDetails(t *testing.T) {
	details := map[string]interface{}{
		"field": "value",
		"count": 123,
	}
	apiErr := NewAPIErrorWithDetails("test operation", "test message", 400, details)

	if apiErr.Operation != "test operation" {
		t.Errorf("NewAPIErrorWithDetails() Operation = %v, want %v", apiErr.Operation, "test operation")
	}
	if apiErr.Message != "test message" {
		t.Errorf("NewAPIErrorWithDetails() Message = %v, want %v", apiErr.Message, "test message")
	}
	if apiErr.StatusCode != 400 {
		t.Errorf("NewAPIErrorWithDetails() StatusCode = %v, want 400", apiErr.StatusCode)
	}
	if apiErr.Details == nil {
		t.Error("NewAPIErrorWithDetails() Details is nil")
	}
	if len(apiErr.Details) != 2 {
		t.Errorf("NewAPIErrorWithDetails() Details length = %v, want 2", len(apiErr.Details))
	}
}

func TestErrorChaining(t *testing.T) {
	// Test error chain with errors.Is
	originalErr := errors.New("original error")
	gitErr := New("operation", originalErr)

	if !errors.Is(gitErr, originalErr) {
		t.Error("errors.Is() should return true for wrapped error")
	}
}

func TestErrorMessages(t *testing.T) {
	// Test that error messages contain expected information
	tests := []struct {
		name     string
		err      error
		contains []string
	}{
		{
			name:     "git error with path",
			err:      NewWithPath("clone", "/repo", errors.New("permission denied")),
			contains: []string{"clone", "/repo", "permission denied"},
		},
		{
			name:     "config error",
			err:      NewConfigError("github_token", "required for PR creation"),
			contains: []string{"configuration error", "github_token", "required for PR creation"},
		},
		{
			name:     "api error",
			err:      NewAPIErrorWithDetails("create PR", "validation failed", 422, nil),
			contains: []string{"GitHub API error", "create PR", "422", "validation failed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := tt.err.Error()
			for _, substr := range tt.contains {
				if !strings.Contains(errMsg, substr) {
					t.Errorf("Error message %q should contain %q", errMsg, substr)
				}
			}
		})
	}
}
