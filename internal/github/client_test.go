package github

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
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
	if client.baseURL != apiBaseURL {
		t.Errorf("baseURL = %q, want %q", client.baseURL, apiBaseURL)
	}
	if client.httpClient == nil {
		t.Error("httpClient is nil")
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
	_, err := client.Post(context.Background(), "/test", make(chan int))
	if err == nil {
		t.Error("Post() with unmarshalable data should return error")
	}
}

func TestPatch_MarshalError(t *testing.T) {
	client := NewClient("token")
	_, err := client.Patch(context.Background(), "/test", make(chan int))
	if err == nil {
		t.Error("Patch() with unmarshalable data should return error")
	}
}

// testClient returns a Client whose requests are directed at the given test
// server URL, keeping the same 30s timeout as production.
func testClient(url string) *Client {
	return &Client{
		token:      "test-token",
		repo:       "owner/repo",
		baseURL:    url,
		httpClient: &http.Client{Timeout: requestTimeout},
	}
}

func TestPost_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify method, path, and the standard GitHub headers are sent.
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/repos/owner/repo/pulls" {
			t.Errorf("path = %q, want /repos/owner/repo/pulls", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", got, "Bearer test-token")
		}
		if got := r.Header.Get("Accept"); got != acceptHeader {
			t.Errorf("Accept = %q, want %q", got, acceptHeader)
		}
		if got := r.Header.Get("X-GitHub-Api-Version"); got != apiVersion {
			t.Errorf("X-GitHub-Api-Version = %q, want %q", got, apiVersion)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", got)
		}
		body, _ := io.ReadAll(r.Body)
		var sent map[string]interface{}
		if err := json.Unmarshal(body, &sent); err != nil {
			t.Errorf("request body is not valid JSON: %v", err)
		}
		if sent["title"] != "hello" {
			t.Errorf("request body title = %v, want hello", sent["title"])
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"number":7,"html_url":"https://example.com/pr/7"}`))
	}))
	defer srv.Close()

	resp, err := testClient(srv.URL).Post(context.Background(), "/pulls", map[string]interface{}{"title": "hello"})
	if err != nil {
		t.Fatalf("Post() error = %v", err)
	}
	if resp["html_url"] != "https://example.com/pr/7" {
		t.Errorf("html_url = %v, want https://example.com/pr/7", resp["html_url"])
	}
	if resp["number"].(float64) != 7 {
		t.Errorf("number = %v, want 7", resp["number"])
	}
}

func TestPost_ArraySuccess(t *testing.T) {
	// Some endpoints (labels) return an array on success → map result is nil.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"name":"bug"}]`))
	}))
	defer srv.Close()

	resp, err := testClient(srv.URL).Post(context.Background(), "/issues/1/labels", map[string]interface{}{"labels": []string{"bug"}})
	if err != nil {
		t.Fatalf("Post() error = %v", err)
	}
	if resp != nil {
		t.Errorf("resp = %v, want nil for array response", resp)
	}
}

func TestRequest_ErrorBodyWithMessage(t *testing.T) {
	// Non-2xx with a JSON body is returned (nil error) so the caller can
	// inspect the "message" field (e.g. "A pull request already exists").
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"message":"A pull request already exists"}`))
	}))
	defer srv.Close()

	resp, err := testClient(srv.URL).Post(context.Background(), "/pulls", map[string]interface{}{"title": "x"})
	if err != nil {
		t.Fatalf("Post() error = %v, want nil (error body returned to caller)", err)
	}
	if resp["message"] != "A pull request already exists" {
		t.Errorf("message = %v", resp["message"])
	}
}

func TestRequest_ErrorUnparseableBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`not json`))
	}))
	defer srv.Close()

	_, err := testClient(srv.URL).Post(context.Background(), "/pulls", map[string]interface{}{"title": "x"})
	if err == nil {
		t.Fatal("Post() with 500 + non-JSON body should return an error")
	}
}

func TestPatch_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %q, want PATCH", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"state":"closed"}`))
	}))
	defer srv.Close()

	resp, err := testClient(srv.URL).Patch(context.Background(), "/pulls/1", map[string]string{"state": "closed"})
	if err != nil {
		t.Fatalf("Patch() error = %v", err)
	}
	if resp["state"] != "closed" {
		t.Errorf("state = %v, want closed", resp["state"])
	}
}

func TestGetArray_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if got := r.Header.Get("Content-Type"); got != "" {
			t.Errorf("GET should not set Content-Type, got %q", got)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"number":11},{"number":12}]`))
	}))
	defer srv.Close()

	result, err := testClient(srv.URL).GetArray(context.Background(), "/pulls?head=x")
	if err != nil {
		t.Fatalf("GetArray() error = %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("len(result) = %d, want 2", len(result))
	}
	if result[0]["number"].(float64) != 11 {
		t.Errorf("result[0].number = %v, want 11", result[0]["number"])
	}
}

func TestGetArray_ErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Not Found"}`))
	}))
	defer srv.Close()

	_, err := testClient(srv.URL).GetArray(context.Background(), "/pulls")
	if err == nil {
		t.Fatal("GetArray() with 404 should return an error")
	}
}
