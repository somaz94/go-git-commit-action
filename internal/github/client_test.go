package github

import (
	"os"
	"testing"
)

func TestNewClient(t *testing.T) {
	os.Setenv("GITHUB_REPOSITORY", "test/repo")
	defer os.Unsetenv("GITHUB_REPOSITORY")

	client := NewClient("test-token")
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}
	if client.token != "test-token" {
		t.Errorf("token = %q, want %q", client.token, "test-token")
	}
	if client.repo != "test/repo" {
		t.Errorf("repo = %q, want %q", client.repo, "test/repo")
	}
}

func TestNewClient_EmptyToken(t *testing.T) {
	client := NewClient("")
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}
	if client.token != "" {
		t.Errorf("token = %q, want empty", client.token)
	}
}

func TestRepo(t *testing.T) {
	os.Setenv("GITHUB_REPOSITORY", "owner/repo-name")
	defer os.Unsetenv("GITHUB_REPOSITORY")

	client := NewClient("token")
	if got := client.Repo(); got != "owner/repo-name" {
		t.Errorf("Repo() = %q, want %q", got, "owner/repo-name")
	}
}

func TestRepo_Empty(t *testing.T) {
	os.Unsetenv("GITHUB_REPOSITORY")

	client := NewClient("token")
	if got := client.Repo(); got != "" {
		t.Errorf("Repo() = %q, want empty", got)
	}
}

func TestConstants(t *testing.T) {
	if apiBaseURL != "https://api.github.com" {
		t.Errorf("apiBaseURL = %q, want %q", apiBaseURL, "https://api.github.com")
	}
	if apiVersion != "2022-11-28" {
		t.Errorf("apiVersion = %q, want %q", apiVersion, "2022-11-28")
	}
	if acceptHeader != "application/vnd.github+json" {
		t.Errorf("acceptHeader = %q, want %q", acceptHeader, "application/vnd.github+json")
	}
}

func TestPost_MarshalError(t *testing.T) {
	client := NewClient("token")
	// channels cannot be marshaled to JSON
	_, err := client.Post("/test", make(chan int))
	if err == nil {
		t.Error("Post() with unmarshalable data should return error")
	}
}

func TestPatch_MarshalError(t *testing.T) {
	client := NewClient("token")
	_, err := client.Patch("/test", make(chan int))
	if err == nil {
		t.Error("Patch() with unmarshalable data should return error")
	}
}
