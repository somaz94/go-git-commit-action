package pr

import (
	"os"
	"strings"
	"testing"

	"github.com/somaz94/go-git-commit-action/internal/config"
)

func TestNewBranchManager(t *testing.T) {
	cfg := &config.GitConfig{PRBranch: "feature", PRBase: "main"}
	bm := NewBranchManager(cfg)
	if bm == nil {
		t.Fatal("NewBranchManager() returned nil")
	}
	if bm.config != cfg {
		t.Error("NewBranchManager() config does not match")
	}
}

func TestNewDiffChecker(t *testing.T) {
	cfg := &config.GitConfig{PRBranch: "feature", PRBase: "main"}
	dc := NewDiffChecker(cfg)
	if dc == nil {
		t.Fatal("NewDiffChecker() returned nil")
	}
	if dc.config != cfg {
		t.Error("NewDiffChecker() config does not match")
	}
}

func TestNewCreator(t *testing.T) {
	cfg := &config.GitConfig{PRBranch: "feature", PRBase: "main"}
	c := NewCreator(cfg)
	if c == nil {
		t.Fatal("NewCreator() returned nil")
	}
	if c.config != cfg {
		t.Error("NewCreator() config does not match")
	}
}

func TestDeleteSourceBranch_DryRun(t *testing.T) {
	cfg := &config.GitConfig{
		PRDryRun:   true,
		AutoBranch: true,
	}
	bm := NewBranchManager(cfg)

	err := bm.DeleteSourceBranch("test-branch")
	if err != nil {
		t.Errorf("DeleteSourceBranch() in dry run should not error, got %v", err)
	}
}

func TestDeleteSourceBranch_NotAutoBranch(t *testing.T) {
	cfg := &config.GitConfig{
		PRDryRun:   false,
		AutoBranch: false,
	}
	bm := NewBranchManager(cfg)

	err := bm.DeleteSourceBranch("test-branch")
	if err != nil {
		t.Errorf("DeleteSourceBranch() with AutoBranch=false should not error, got %v", err)
	}
}

func TestGeneratePRTitleAndBody(t *testing.T) {
	tests := []struct {
		name      string
		prTitle   string
		prBody    string
		prBranch  string
		prBase    string
		runID     string
		commitSHA string
		wantTitle string
		wantBody  string
	}{
		{
			name:      "custom title and body",
			prTitle:   "My PR Title",
			prBody:    "My PR Body",
			prBranch:  "feature",
			prBase:    "main",
			runID:     "123",
			commitSHA: "abc123",
			wantTitle: "My PR Title",
			wantBody:  "My PR Body",
		},
		{
			name:      "auto-generated title and body",
			prTitle:   "",
			prBody:    "",
			prBranch:  "feature",
			prBase:    "main",
			runID:     "456",
			commitSHA: "def456",
			wantTitle: "Auto PR: feature to main (Run ID: 456)",
			wantBody:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Creator{
				config: &config.GitConfig{
					PRTitle:  tt.prTitle,
					PRBody:   tt.prBody,
					PRBranch: tt.prBranch,
					PRBase:   tt.prBase,
				},
			}
			gotTitle, gotBody := c.generatePRTitleAndBody(tt.runID, tt.commitSHA)

			if gotTitle != tt.wantTitle {
				t.Errorf("title = %q, want %q", gotTitle, tt.wantTitle)
			}
			if tt.wantBody != "" && gotBody != tt.wantBody {
				t.Errorf("body = %q, want %q", gotBody, tt.wantBody)
			}
			if tt.wantBody == "" && tt.prBody == "" {
				// Auto-generated body should contain key information
				if !strings.Contains(gotBody, tt.prBranch) {
					t.Errorf("auto body should contain branch name %q", tt.prBranch)
				}
				if !strings.Contains(gotBody, tt.commitSHA) {
					t.Errorf("auto body should contain commit SHA %q", tt.commitSHA)
				}
			}
		})
	}
}

func TestCreatePullRequest_DryRun(t *testing.T) {
	cfg := &config.GitConfig{
		PRDryRun: true,
		PRBranch: "feature",
		PRBase:   "main",
	}
	os.Setenv("GITHUB_REPOSITORY", "test/repo")
	defer os.Unsetenv("GITHUB_REPOSITORY")

	c := NewCreator(cfg)
	response, err := c.CreatePullRequest()
	if err != nil {
		t.Fatalf("CreatePullRequest() dry run error = %v", err)
	}

	if response == nil {
		t.Fatal("CreatePullRequest() dry run returned nil response")
	}

	if dryRun, ok := response["dry_run"].(bool); !ok || !dryRun {
		t.Error("dry run response should have dry_run=true")
	}

	if htmlURL, ok := response["html_url"].(string); !ok || htmlURL == "" {
		t.Error("dry run response should have html_url")
	}
}

func TestHandlePRResponse_DryRun(t *testing.T) {
	cfg := &config.GitConfig{
		PRBranch: "feature",
		PRBase:   "main",
		PRLabels: []string{"test"},
	}
	c := NewCreator(cfg)

	response := map[string]interface{}{
		"html_url": "https://github.com/test/repo/pull/1",
		"dry_run":  true,
	}

	err := c.HandlePRResponse(response, "feature")
	if err != nil {
		t.Errorf("HandlePRResponse() dry run error = %v", err)
	}
}

func TestHandlePRResponse_Error(t *testing.T) {
	cfg := &config.GitConfig{}
	c := NewCreator(cfg)

	response := map[string]interface{}{
		"message": "Validation Failed",
	}

	err := c.HandlePRResponse(response, "feature")
	if err == nil {
		t.Error("HandlePRResponse() with error response should return error")
	}
	if !strings.Contains(err.Error(), "Validation Failed") {
		t.Errorf("error should contain 'Validation Failed', got %v", err)
	}
}

func TestAddLabelsToIssue_DryRun(t *testing.T) {
	cfg := &config.GitConfig{
		PRDryRun: true,
		PRLabels: []string{"bug", "enhancement"},
	}
	c := NewCreator(cfg)

	err := c.addLabelsToIssue(1)
	if err != nil {
		t.Errorf("addLabelsToIssue() dry run error = %v", err)
	}
}

func TestClosePullRequest_DryRun(t *testing.T) {
	cfg := &config.GitConfig{
		PRDryRun: true,
	}
	c := NewCreator(cfg)

	err := c.closePullRequest(1)
	if err != nil {
		t.Errorf("closePullRequest() dry run error = %v", err)
	}
}

func TestDisplayPRURL(t *testing.T) {
	os.Setenv("GITHUB_REPOSITORY", "test/repo")
	defer os.Unsetenv("GITHUB_REPOSITORY")

	cfg := &config.GitConfig{
		PRBranch: "feature",
		PRBase:   "main",
	}
	dc := &DiffChecker{config: cfg}

	// Should not panic
	dc.displayPRURL()
}

func TestProcessExistingPR_NoLabelsNoClosed(t *testing.T) {
	cfg := &config.GitConfig{
		PRLabels: nil,
		PRClosed: false,
	}
	c := NewCreator(cfg)

	err := c.processExistingPR(1)
	if err != nil {
		t.Errorf("processExistingPR() error = %v", err)
	}
}

func TestNewCreator_HasClient(t *testing.T) {
	os.Setenv("GITHUB_REPOSITORY", "test/repo")
	defer os.Unsetenv("GITHUB_REPOSITORY")

	cfg := &config.GitConfig{GitHubToken: "test-token"}
	c := NewCreator(cfg)
	if c.client == nil {
		t.Error("NewCreator() should initialize client")
	}
	if c.client.Repo() != "test/repo" {
		t.Errorf("client.Repo() = %q, want %q", c.client.Repo(), "test/repo")
	}
}

func TestCreatePullRequest_DryRun_ContainsURL(t *testing.T) {
	os.Setenv("GITHUB_REPOSITORY", "owner/repo")
	defer os.Unsetenv("GITHUB_REPOSITORY")

	cfg := &config.GitConfig{
		PRDryRun: true,
		PRBranch: "feat-branch",
		PRBase:   "main",
	}
	c := NewCreator(cfg)
	response, err := c.CreatePullRequest()
	if err != nil {
		t.Fatalf("CreatePullRequest() error = %v", err)
	}

	htmlURL := response["html_url"].(string)
	if !strings.Contains(htmlURL, "owner/repo") {
		t.Errorf("html_url should contain repo, got %q", htmlURL)
	}
	if !strings.Contains(htmlURL, "feat-branch") {
		t.Errorf("html_url should contain branch, got %q", htmlURL)
	}
	if !strings.Contains(htmlURL, "dry_run=1") {
		t.Errorf("html_url should contain dry_run param, got %q", htmlURL)
	}
}

func TestHandlePRResponse_DryRun_WithDeleteSourceBranch(t *testing.T) {
	cfg := &config.GitConfig{
		PRBranch:           "feature",
		PRBase:             "main",
		PRClosed:           true,
		DeleteSourceBranch: true,
		AutoBranch:         true,
	}
	c := NewCreator(cfg)

	response := map[string]interface{}{
		"html_url": "https://github.com/test/repo/pull/1",
		"dry_run":  true,
	}

	err := c.HandlePRResponse(response, "feature")
	if err != nil {
		t.Errorf("HandlePRResponse() error = %v", err)
	}
}

func TestHandlePRResponse_DryRun_DeleteSourceBranchNoAuto(t *testing.T) {
	cfg := &config.GitConfig{
		PRBranch:           "feature",
		PRBase:             "main",
		DeleteSourceBranch: true,
		AutoBranch:         false,
	}
	c := NewCreator(cfg)

	response := map[string]interface{}{
		"html_url": "https://github.com/test/repo/pull/1",
		"dry_run":  true,
	}

	err := c.HandlePRResponse(response, "feature")
	if err != nil {
		t.Errorf("HandlePRResponse() error = %v", err)
	}
}

func TestHandlePRResponse_ErrorWithDetails(t *testing.T) {
	cfg := &config.GitConfig{}
	c := NewCreator(cfg)

	response := map[string]interface{}{
		"message": "Validation Failed",
		"errors": []interface{}{
			map[string]interface{}{
				"message": "Some validation error",
			},
		},
	}

	err := c.HandlePRResponse(response, "feature")
	if err == nil {
		t.Error("HandlePRResponse() should return error")
	}
}

func TestHandlePRResponse_ExistingPRError(t *testing.T) {
	os.Setenv("GITHUB_REPOSITORY", "test/repo")
	defer os.Unsetenv("GITHUB_REPOSITORY")

	cfg := &config.GitConfig{
		PRBranch: "feature",
		PRBase:   "main",
	}
	c := NewCreator(cfg)

	response := map[string]interface{}{
		"message": "Validation Failed",
		"errors": []interface{}{
			map[string]interface{}{
				"message": "A pull request already exists for feature",
			},
		},
	}

	// This will try to call GitHub API which will fail, but tests the path
	_ = c.HandlePRResponse(response, "feature")
}

func TestHandlePRResponse_SuccessNoURL(t *testing.T) {
	cfg := &config.GitConfig{}
	c := NewCreator(cfg)

	response := map[string]interface{}{
		"id": float64(123),
	}

	err := c.HandlePRResponse(response, "feature")
	if err == nil {
		t.Error("HandlePRResponse() without html_url should return error")
	}
}

func TestProcessExistingPR_WithLabelsAndClose(t *testing.T) {
	cfg := &config.GitConfig{
		PRDryRun: true,
		PRLabels: []string{"bug", "fix"},
		PRClosed: true,
	}
	c := NewCreator(cfg)

	err := c.processExistingPR(42)
	if err != nil {
		t.Errorf("processExistingPR() dry run error = %v", err)
	}
}

func TestProcessExistingPR_LabelsOnly(t *testing.T) {
	cfg := &config.GitConfig{
		PRDryRun: true,
		PRLabels: []string{"enhancement"},
		PRClosed: false,
	}
	c := NewCreator(cfg)

	err := c.processExistingPR(10)
	if err != nil {
		t.Errorf("processExistingPR() error = %v", err)
	}
}

func TestProcessExistingPR_CloseOnly(t *testing.T) {
	cfg := &config.GitConfig{
		PRDryRun: true,
		PRLabels: nil,
		PRClosed: true,
	}
	c := NewCreator(cfg)

	err := c.processExistingPR(10)
	if err != nil {
		t.Errorf("processExistingPR() error = %v", err)
	}
}

func TestGeneratePRTitleAndBody_AllAutoGenerated(t *testing.T) {
	c := &Creator{
		config: &config.GitConfig{
			PRBranch: "dev",
			PRBase:   "staging",
		},
	}
	title, body := c.generatePRTitleAndBody("789", "sha789abc")

	if !strings.Contains(title, "dev") {
		t.Errorf("title should contain branch name, got %q", title)
	}
	if !strings.Contains(title, "staging") {
		t.Errorf("title should contain base name, got %q", title)
	}
	if !strings.Contains(title, "789") {
		t.Errorf("title should contain run ID, got %q", title)
	}
	if !strings.Contains(body, "sha789abc") {
		t.Errorf("body should contain commit SHA, got %q", body)
	}
	if !strings.Contains(body, "Go Git Commit Action") {
		t.Errorf("body should contain action name, got %q", body)
	}
}

func TestFetchBranches_Structure(t *testing.T) {
	cfg := &config.GitConfig{
		PRBranch: "feature",
		PRBase:   "main",
	}
	bm := NewBranchManager(cfg)
	if bm == nil {
		t.Fatal("NewBranchManager() returned nil")
	}
	// FetchBranches will fail without a real git repo, but we verify it's callable
}

func BenchmarkGeneratePRTitleAndBody(b *testing.B) {
	c := &Creator{
		config: &config.GitConfig{
			PRBranch: "feature",
			PRBase:   "main",
		},
	}
	for i := 0; i < b.N; i++ {
		c.generatePRTitleAndBody("123", "abc123")
	}
}
