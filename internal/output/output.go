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
func (r *Result) WriteToGitHubOutput() (err error) {
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

	r.mu.Lock()
	defer r.mu.Unlock()

	f, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open GITHUB_OUTPUT file: %w", err)
	}
	// Surface a flush/close failure (e.g. ENOSPC) as the function error
	// so dropped action outputs are not silently ignored.
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("failed to close GITHUB_OUTPUT file: %w", cerr)
		}
	}()

	var lines []string
	for k, v := range r.values {
		lines = append(lines, fmt.Sprintf("%s=%s", k, v))
	}

	if len(lines) > 0 {
		_, err = fmt.Fprintln(f, strings.Join(lines, "\n"))
		if err != nil {
			return fmt.Errorf("failed to write to GITHUB_OUTPUT: %w", err)
		}
	}

	fmt.Println("\nAction Outputs:")
	for k, v := range r.values {
		fmt.Printf("  %s=%s\n", k, v)
	}

	return nil
}
