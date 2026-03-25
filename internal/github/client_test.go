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

func TestParseHTTPResponse_Success(t *testing.T) {
	input := []byte(`{"key":"value"}
200`)
	body, code, err := parseHTTPResponse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 200 {
		t.Errorf("status code = %d, want 200", code)
	}
	if string(body) != `{"key":"value"}` {
		t.Errorf("body = %q, want %q", string(body), `{"key":"value"}`)
	}
}

func TestParseHTTPResponse_ErrorStatus(t *testing.T) {
	input := []byte(`{"message":"Not Found"}
404`)
	body, code, err := parseHTTPResponse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 404 {
		t.Errorf("status code = %d, want 404", code)
	}
	if string(body) != `{"message":"Not Found"}` {
		t.Errorf("body = %q", string(body))
	}
}

func TestParseHTTPResponse_OnlyStatusCode(t *testing.T) {
	input := []byte("204")
	body, code, err := parseHTTPResponse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 204 {
		t.Errorf("status code = %d, want 204", code)
	}
	if body != nil {
		t.Errorf("body = %q, want nil", string(body))
	}
}

func TestParseHTTPResponse_InvalidStatusCode(t *testing.T) {
	input := []byte(`{"data":"test"}
abc`)
	_, _, err := parseHTTPResponse(input)
	if err == nil {
		t.Error("expected error for invalid status code")
	}
}

func TestParseHTTPResponse_MultilineBody(t *testing.T) {
	input := []byte(`{
  "key": "value",
  "nested": true
}
201`)
	body, code, err := parseHTTPResponse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 201 {
		t.Errorf("status code = %d, want 201", code)
	}
	if len(body) == 0 {
		t.Error("expected non-empty body")
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
