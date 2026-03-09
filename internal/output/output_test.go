package output

import (
	"os"
	"strings"
	"testing"
)

func TestNewResult(t *testing.T) {
	r := NewResult()
	if r == nil {
		t.Fatal("NewResult() returned nil")
	}
	if r.values == nil {
		t.Fatal("NewResult() values map is nil")
	}
}

func TestResult_SetAndGet(t *testing.T) {
	r := NewResult()

	r.Set(KeyCommitSHA, "abc123")
	r.Set(KeyPRNumber, "42")
	r.Set(KeySkipped, "false")

	if got := r.Get(KeyCommitSHA); got != "abc123" {
		t.Errorf("Get(%s) = %q, want %q", KeyCommitSHA, got, "abc123")
	}
	if got := r.Get(KeyPRNumber); got != "42" {
		t.Errorf("Get(%s) = %q, want %q", KeyPRNumber, got, "42")
	}
	if got := r.Get(KeySkipped); got != "false" {
		t.Errorf("Get(%s) = %q, want %q", KeySkipped, got, "false")
	}
}

func TestResult_GetEmpty(t *testing.T) {
	r := NewResult()
	if got := r.Get("nonexistent"); got != "" {
		t.Errorf("Get(nonexistent) = %q, want empty string", got)
	}
}

func TestResult_SetOverwrite(t *testing.T) {
	r := NewResult()
	r.Set(KeyCommitSHA, "first")
	r.Set(KeyCommitSHA, "second")

	if got := r.Get(KeyCommitSHA); got != "second" {
		t.Errorf("Get(%s) = %q, want %q after overwrite", KeyCommitSHA, got, "second")
	}
}

func TestResult_WriteToGitHubOutput_NoEnvVar(t *testing.T) {
	// When GITHUB_OUTPUT is not set, should print to stdout without error
	os.Unsetenv("GITHUB_OUTPUT")

	r := NewResult()
	r.Set(KeySkipped, "true")
	r.Set(KeyChangedFiles, "0")

	if err := r.WriteToGitHubOutput(); err != nil {
		t.Fatalf("WriteToGitHubOutput() error = %v", err)
	}
}

func TestResult_WriteToGitHubOutput_WithFile(t *testing.T) {
	// Create a temp file to simulate GITHUB_OUTPUT
	tmpFile, err := os.CreateTemp("", "github-output-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	t.Setenv("GITHUB_OUTPUT", tmpFile.Name())

	r := NewResult()
	r.Set(KeyCommitSHA, "abc123def456")
	r.Set(KeySkipped, "false")
	r.Set(KeyChangedFiles, "3")

	if err := r.WriteToGitHubOutput(); err != nil {
		t.Fatalf("WriteToGitHubOutput() error = %v", err)
	}

	// Read the file and verify contents
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	output := string(content)
	if !strings.Contains(output, "commit_sha=abc123def456") {
		t.Errorf("Output file should contain commit_sha=abc123def456, got: %s", output)
	}
	if !strings.Contains(output, "skipped=false") {
		t.Errorf("Output file should contain skipped=false, got: %s", output)
	}
	if !strings.Contains(output, "changed_files=3") {
		t.Errorf("Output file should contain changed_files=3, got: %s", output)
	}
}

func TestResult_WriteToGitHubOutput_EmptyResult(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "github-output-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	t.Setenv("GITHUB_OUTPUT", tmpFile.Name())

	r := NewResult()
	if err := r.WriteToGitHubOutput(); err != nil {
		t.Fatalf("WriteToGitHubOutput() error = %v", err)
	}

	// File should be empty (or unchanged)
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	if len(content) != 0 {
		t.Errorf("Output file should be empty for no results, got: %s", string(content))
	}
}

func TestResult_WriteToGitHubOutput_InvalidPath(t *testing.T) {
	t.Setenv("GITHUB_OUTPUT", "/nonexistent/path/output")

	r := NewResult()
	r.Set(KeySkipped, "true")

	err := r.WriteToGitHubOutput()
	if err == nil {
		t.Fatal("WriteToGitHubOutput() should return error for invalid path")
	}
}

func TestKeyConstants(t *testing.T) {
	// Verify key constants are defined correctly
	keys := map[string]string{
		"commit_sha":    KeyCommitSHA,
		"pr_number":     KeyPRNumber,
		"pr_url":        KeyPRURL,
		"tag_name":      KeyTagName,
		"skipped":       KeySkipped,
		"changed_files": KeyChangedFiles,
	}

	for expected, got := range keys {
		if got != expected {
			t.Errorf("Key constant = %q, want %q", got, expected)
		}
	}
}
