package output

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

// Key constants for action outputs.
const (
	KeyCommitSHA    = "commit_sha"
	KeyPRNumber     = "pr_number"
	KeyPRURL        = "pr_url"
	KeyTagName      = "tag_name"
	KeySkipped      = "skipped"
	KeyChangedFiles = "changed_files"
)

// Result holds all output values to be written to GITHUB_OUTPUT.
type Result struct {
	mu     sync.Mutex
	values map[string]string
}

// NewResult creates a new Result instance.
func NewResult() *Result {
	return &Result{
		values: make(map[string]string),
	}
}

// Set stores a key-value pair in the result.
func (r *Result) Set(key, value string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.values[key] = value
}

// Get retrieves a value by key.
func (r *Result) Get(key string) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.values[key]
}

// WriteToGitHubOutput writes all stored values to the GITHUB_OUTPUT file.
// In GitHub Actions, this file path is provided via the GITHUB_OUTPUT env var.
func (r *Result) WriteToGitHubOutput() error {
	outputFile := os.Getenv("GITHUB_OUTPUT")
	if outputFile == "" {
		// Not running in GitHub Actions; print outputs to stdout instead
		r.mu.Lock()
		defer r.mu.Unlock()
		if len(r.values) > 0 {
			fmt.Println("\nAction Outputs:")
			for k, v := range r.values {
				fmt.Printf("  %s=%s\n", k, v)
			}
		}
		return nil
	}

	f, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open GITHUB_OUTPUT file: %v", err)
	}
	defer f.Close()

	r.mu.Lock()
	defer r.mu.Unlock()

	var lines []string
	for k, v := range r.values {
		lines = append(lines, fmt.Sprintf("%s=%s", k, v))
	}

	if len(lines) > 0 {
		_, err = fmt.Fprintln(f, strings.Join(lines, "\n"))
		if err != nil {
			return fmt.Errorf("failed to write to GITHUB_OUTPUT: %v", err)
		}
	}

	fmt.Println("\nAction Outputs:")
	for k, v := range r.values {
		fmt.Printf("  %s=%s\n", k, v)
	}

	return nil
}
