package errors

import "fmt"

// GitError represents an error that occurred during a Git operation.
// It provides structured error information including the operation,
// path (if applicable), and the underlying error.
type GitError struct {
	Op   string // Operation that failed (e.g., "commit", "push", "tag")
	Path string // Path related to the error (optional)
	Err  error  // Underlying error
}

// Error implements the error interface.
func (e *GitError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("%s %s: %v", e.Op, e.Path, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

// Unwrap returns the underlying error for error chain support.
func (e *GitError) Unwrap() error {
	return e.Err
}

// New creates a new GitError with the specified operation and error.
func New(op string, err error) *GitError {
	return &GitError{
		Op:  op,
		Err: err,
	}
}

// NewWithPath creates a new GitError with the specified operation, path, and error.
func NewWithPath(op, path string, err error) *GitError {
	return &GitError{
		Op:   op,
		Path: path,
		Err:  err,
	}
}

// ConfigError represents a configuration validation error.
type ConfigError struct {
	Field   string // Configuration field that failed validation
	Message string // Error message
}

// Error implements the error interface.
func (e *ConfigError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("configuration error in %s: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("configuration error: %s", e.Message)
}

// NewConfigError creates a new ConfigError with field and message.
func NewConfigError(field, message string) *ConfigError {
	return &ConfigError{
		Field:   field,
		Message: message,
	}
}

// NewConfig creates a new ConfigError with just a message (no specific field).
func NewConfig(message string) *ConfigError {
	return &ConfigError{
		Message: message,
	}
}

// RetryError represents an error after multiple retry attempts.
type RetryError struct {
	Attempts int    // Number of retry attempts made
	LastErr  error  // Last error encountered
	Message  string // Operation description
}

// Error implements the error interface.
func (e *RetryError) Error() string {
	return fmt.Sprintf("%s (failed after %d attempts): %v", e.Message, e.Attempts, e.LastErr)
}

// Unwrap returns the underlying error for error chain support.
func (e *RetryError) Unwrap() error {
	return e.LastErr
}

// NewWithContext creates a new RetryError.
func NewWithContext(message string, attempts int, lastErr error) *RetryError {
	return &RetryError{
		Message:  message,
		Attempts: attempts,
		LastErr:  lastErr,
	}
}

// APIError represents an error from the GitHub API.
type APIError struct {
	Operation  string                 // API operation (e.g., "create PR", "add labels")
	StatusCode int                    // HTTP status code (if applicable)
	Message    string                 // Error message from API
	Details    map[string]interface{} // Additional error details
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("GitHub API error (%s) [%d]: %s", e.Operation, e.StatusCode, e.Message)
	}
	return fmt.Sprintf("GitHub API error (%s): %s", e.Operation, e.Message)
}

// NewAPIError creates a new APIError.
func NewAPIError(operation, message string) *APIError {
	return &APIError{
		Operation: operation,
		Message:   message,
	}
}

// NewAPIErrorWithDetails creates a new APIError with additional details.
func NewAPIErrorWithDetails(operation, message string, statusCode int, details map[string]interface{}) *APIError {
	return &APIError{
		Operation:  operation,
		StatusCode: statusCode,
		Message:    message,
		Details:    details,
	}
}
