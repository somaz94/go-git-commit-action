package pr

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/somaz94/go-git-commit-action/internal/config"
	"github.com/somaz94/go-git-commit-action/internal/gitcmd"
	"github.com/somaz94/go-git-commit-action/internal/github"
)

// apiCall is one request the fake GitHub API received.
type apiCall struct {
	Method string
	Path   string
	Body   map[string]any
}

// fakeAPI is an httptest-backed stand-in for the GitHub REST API. routes maps
// "<METHOD> <path>" to the handler for that endpoint; anything unrouted returns
// 404 so an unexpected call fails loudly rather than silently passing.
type fakeAPI struct {
	t      *testing.T
	server *httptest.Server

	mu    sync.Mutex
	calls []apiCall

	routes map[string]func(w http.ResponseWriter)
}

func newFakeAPI(t *testing.T) *fakeAPI {
	t.Helper()
	f := &fakeAPI{t: t, routes: make(map[string]func(http.ResponseWriter))}
	f.server = httptest.NewServer(http.HandlerFunc(f.handle))
	t.Cleanup(f.server.Close)
	return f
}

func (f *fakeAPI) handle(w http.ResponseWriter, r *http.Request) {
	raw, _ := io.ReadAll(r.Body)
	var body map[string]any
	_ = json.Unmarshal(raw, &body)

	// The client prefixes every endpoint with /repos/<owner>/<repo>.
	path := strings.TrimPrefix(r.URL.RequestURI(), "/repos/owner/repo")

	f.mu.Lock()
	f.calls = append(f.calls, apiCall{Method: r.Method, Path: path, Body: body})
	handler, ok := f.routes[r.Method+" "+path]
	f.mu.Unlock()

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
		return
	}
	handler(w)
}

// route registers a JSON response for one endpoint.
func (f *fakeAPI) route(methodAndPath string, status int, body string) *fakeAPI {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.routes[methodAndPath] = func(w http.ResponseWriter) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
	return f
}

func (f *fakeAPI) Calls() []apiCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]apiCall(nil), f.calls...)
}

// called reports whether an endpoint was hit, returning the first matching call.
func (f *fakeAPI) called(methodAndPath string) (apiCall, bool) {
	for _, c := range f.Calls() {
		if c.Method+" "+c.Path == methodAndPath {
			return c, true
		}
	}
	return apiCall{}, false
}

// newAPICreator wires a Creator to the fake API and a fake git Runner.
func newAPICreator(t *testing.T, cfg *config.GitConfig, api *fakeAPI) (*Creator, *gitcmd.FakeRunner) {
	t.Helper()
	t.Setenv("GITHUB_REPOSITORY", "owner/repo")
	r := gitcmd.NewFakeRunner().
		Stub(key(gitcmd.RevParseArgs("HEAD")), gitcmd.FakeResult{Stdout: "abc1234\n"})
	client := github.NewClientWithBaseURL(cfg.GitHubToken, api.server.URL)
	return NewCreatorWithClient(cfg, r, client), r
}

func TestCreatePullRequest_PostsToAPI(t *testing.T) {
	api := newFakeAPI(t).
		route("POST /pulls", http.StatusCreated,
			`{"html_url":"https://github.com/owner/repo/pull/7","number":7}`)
	cfg := prConfig()
	cfg.PRTitle = "My title"
	cfg.PRBody = "My body"
	c, _ := newAPICreator(t, cfg, api)

	resp, err := c.CreatePullRequest(context.Background())
	if err != nil {
		t.Fatalf("CreatePullRequest() error = %v, want nil", err)
	}
	if resp.Number != 7 || !resp.HasNumber {
		t.Errorf("Number = %d (has=%v), want 7 (has=true)", resp.Number, resp.HasNumber)
	}
	if resp.HTMLURL != "https://github.com/owner/repo/pull/7" {
		t.Errorf("HTMLURL = %q, want the created PR URL", resp.HTMLURL)
	}

	call, ok := api.called("POST /pulls")
	if !ok {
		t.Fatalf("Calls() = %v, want a POST /pulls", api.Calls())
	}
	for field, want := range map[string]string{
		"title": "My title",
		"body":  "My body",
		"head":  cfg.PRBranch,
		"base":  cfg.PRBase,
	} {
		if got, _ := call.Body[field].(string); got != want {
			t.Errorf("payload[%q] = %q, want %q", field, got, want)
		}
	}
	if _, present := call.Body["draft"]; present {
		t.Error("payload carries a draft flag, want it omitted when pr_draft is off")
	}
}

func TestCreatePullRequest_DraftFlag(t *testing.T) {
	api := newFakeAPI(t).
		route("POST /pulls", http.StatusCreated, `{"html_url":"u","number":1}`)
	cfg := prConfig()
	cfg.PRDraft = true
	c, _ := newAPICreator(t, cfg, api)

	if _, err := c.CreatePullRequest(context.Background()); err != nil {
		t.Fatalf("CreatePullRequest() error = %v, want nil", err)
	}
	call, _ := api.called("POST /pulls")
	if draft, _ := call.Body["draft"].(bool); !draft {
		t.Errorf("payload[draft] = %v, want true", call.Body["draft"])
	}
}

// With no explicit title/body, the generated body must carry the resolved
// commit SHA read through the git Runner.
func TestCreatePullRequest_GeneratedBodyCarriesCommitSHA(t *testing.T) {
	api := newFakeAPI(t).
		route("POST /pulls", http.StatusCreated, `{"html_url":"u","number":1}`)
	c, _ := newAPICreator(t, prConfig(), api)

	if _, err := c.CreatePullRequest(context.Background()); err != nil {
		t.Fatalf("CreatePullRequest() error = %v, want nil", err)
	}
	call, _ := api.called("POST /pulls")
	body, _ := call.Body["body"].(string)
	if !strings.Contains(body, "abc1234") {
		t.Errorf("generated body = %q, want it to carry the commit SHA", body)
	}
}

// A failure reading the commit SHA must abort before any API call.
func TestCreatePullRequest_CommitSHAFailureAborts(t *testing.T) {
	api := newFakeAPI(t)
	t.Setenv("GITHUB_REPOSITORY", "owner/repo")
	cfg := prConfig()
	r := &gitcmd.FakeRunner{Default: gitcmd.FakeResult{Err: gitcmd.Fail(128)}}
	c := NewCreatorWithClient(cfg, r, github.NewClientWithBaseURL(cfg.GitHubToken, api.server.URL))

	if _, err := c.CreatePullRequest(context.Background()); err == nil {
		t.Fatal("CreatePullRequest() error = nil, want the rev-parse failure")
	}
	if len(api.Calls()) != 0 {
		t.Errorf("Calls() = %v, want no API call once the SHA could not be read", api.Calls())
	}
}

func TestHandlePRResponse_AppliesLabelsReviewersAssignees(t *testing.T) {
	api := newFakeAPI(t).
		route("POST /issues/7/labels", http.StatusOK, `[]`).
		route("POST /pulls/7/requested_reviewers", http.StatusCreated, `{}`).
		route("POST /issues/7/assignees", http.StatusCreated, `{}`)
	cfg := prConfig()
	cfg.PRLabels = []string{"automated"}
	cfg.PRReviewers = []string{"alice"}
	cfg.PRAssignees = []string{"bob"}
	c, _ := newAPICreator(t, cfg, api)

	resp := PRResponse{HTMLURL: "https://example.com/pull/7", Number: 7, HasNumber: true}
	if err := c.HandlePRResponse(context.Background(), resp, "feature"); err != nil {
		t.Fatalf("HandlePRResponse() error = %v, want nil", err)
	}

	for _, want := range []string{
		"POST /issues/7/labels",
		"POST /pulls/7/requested_reviewers",
		"POST /issues/7/assignees",
	} {
		if _, ok := api.called(want); !ok {
			t.Errorf("Calls() = %v, want it to contain %q", api.Calls(), want)
		}
	}
}

func TestHandlePRResponse_ClosesPR(t *testing.T) {
	api := newFakeAPI(t).route("PATCH /pulls/7", http.StatusOK, `{}`)
	cfg := prConfig()
	cfg.PRClosed = true
	c, _ := newAPICreator(t, cfg, api)

	resp := PRResponse{HTMLURL: "u", Number: 7, HasNumber: true}
	if err := c.HandlePRResponse(context.Background(), resp, "feature"); err != nil {
		t.Fatalf("HandlePRResponse() error = %v, want nil", err)
	}
	call, ok := api.called("PATCH /pulls/7")
	if !ok {
		t.Fatalf("Calls() = %v, want a PATCH closing the PR", api.Calls())
	}
	if state, _ := call.Body["state"].(string); state != "closed" {
		t.Errorf("payload[state] = %q, want %q", state, "closed")
	}
}

// A failing label call must surface as an error, not be swallowed.
func TestHandlePRResponse_LabelFailurePropagates(t *testing.T) {
	api := newFakeAPI(t) // no route → 404
	cfg := prConfig()
	cfg.PRLabels = []string{"automated"}
	c, _ := newAPICreator(t, cfg, api)

	resp := PRResponse{HTMLURL: "u", Number: 7, HasNumber: true}
	err := c.HandlePRResponse(context.Background(), resp, "feature")
	if err == nil {
		t.Fatal("HandlePRResponse() error = nil, want the label failure to propagate")
	}
}

// A rejected label call must surface as an error. GitHub returns 4xx with a
// JSON body, which the client hands back as (body, nil) rather than an error.
func TestApplyToPR_RejectedCloseFails(t *testing.T) {
	api := newFakeAPI(t).
		route("PATCH /pulls/7", http.StatusForbidden, `{"message":"Resource not accessible"}`)
	cfg := prConfig()
	cfg.PRClosed = true
	c, _ := newAPICreator(t, cfg, api)

	resp := PRResponse{HTMLURL: "u", Number: 7, HasNumber: true}
	err := c.HandlePRResponse(context.Background(), resp, "feature")
	if err == nil {
		t.Fatal("HandlePRResponse() error = nil, want the rejected close to fail")
	}
	if !strings.Contains(err.Error(), "Resource not accessible") {
		t.Errorf("error = %q, want it to carry the API message", err.Error())
	}
}

// An "already exists" error must look up the open PR and apply operations to it
// rather than failing the action.
func TestHandlePRResponse_ExistingPRIsReused(t *testing.T) {
	cfg := prConfig()
	cfg.PRLabels = []string{"automated"}
	api := newFakeAPI(t).
		route("GET /pulls?head="+cfg.PRBranch+"&base="+cfg.PRBase, http.StatusOK,
			`[{"number":42}]`).
		route("POST /issues/42/labels", http.StatusOK, `[]`)
	c, _ := newAPICreator(t, cfg, api)

	resp := PRResponse{
		Message: "Validation Failed",
		Errors:  []any{map[string]any{"message": "A pull request already exists for owner:feature."}},
	}
	if err := c.HandlePRResponse(context.Background(), resp, "feature"); err != nil {
		t.Fatalf("HandlePRResponse() error = %v, want the existing PR to be reused", err)
	}
	if _, ok := api.called("POST /issues/42/labels"); !ok {
		t.Errorf("Calls() = %v, want labels applied to the existing PR #42", api.Calls())
	}
}

// An error message that is not "already exists" must fail.
func TestHandlePRResponse_OtherAPIErrorFails(t *testing.T) {
	api := newFakeAPI(t)
	c, _ := newAPICreator(t, prConfig(), api)

	resp := PRResponse{
		Message: "Bad credentials",
		Errors:  []any{map[string]any{"message": "no access"}},
	}
	if err := c.HandlePRResponse(context.Background(), resp, "feature"); err == nil {
		t.Fatal("HandlePRResponse() error = nil, want the API error to fail the action")
	}
}

// When the lookup finds no open PR, the run completes without applying anything.
func TestHandlePRResponse_ExistingPRLookupEmpty(t *testing.T) {
	cfg := prConfig()
	cfg.PRLabels = []string{"automated"}
	api := newFakeAPI(t).
		route("GET /pulls?head="+cfg.PRBranch+"&base="+cfg.PRBase, http.StatusOK, `[]`)
	c, _ := newAPICreator(t, cfg, api)

	resp := PRResponse{
		Message: "Validation Failed",
		Errors:  []any{map[string]any{"message": "A pull request already exists for owner:feature."}},
	}
	if err := c.HandlePRResponse(context.Background(), resp, "feature"); err != nil {
		t.Fatalf("HandlePRResponse() error = %v, want an empty lookup to be tolerated", err)
	}
	if _, ok := api.called("POST /issues/42/labels"); ok {
		t.Error("labels were applied with no PR found, want none")
	}
}

func TestHandlePRResponse_ExistingPRLookupFails(t *testing.T) {
	api := newFakeAPI(t) // no route → 404
	c, _ := newAPICreator(t, prConfig(), api)

	resp := PRResponse{
		Message: "Validation Failed",
		Errors:  []any{map[string]any{"message": "A pull request already exists for owner:feature."}},
	}
	if err := c.HandlePRResponse(context.Background(), resp, "feature"); err == nil {
		t.Fatal("HandlePRResponse() error = nil, want the failed lookup to propagate")
	}
}

// A successful response with no URL is treated as a failure.
func TestHandlePRResponse_MissingURLFails(t *testing.T) {
	api := newFakeAPI(t)
	c, _ := newAPICreator(t, prConfig(), api)

	err := c.HandlePRResponse(context.Background(), PRResponse{}, "feature")
	if err == nil {
		t.Fatal("HandlePRResponse() error = nil, want a missing PR URL to fail")
	}
}

// After a successful PR the auto-generated source branch is deleted.
func TestHandlePRResponse_DeletesAutoSourceBranch(t *testing.T) {
	api := newFakeAPI(t)
	cfg := prConfig()
	cfg.AutoBranch = true
	cfg.DeleteSourceBranch = true
	c, r := newAPICreator(t, cfg, api)

	resp := PRResponse{HTMLURL: "u"}
	if err := c.HandlePRResponse(context.Background(), resp, "update-files-x"); err != nil {
		t.Fatalf("HandlePRResponse() error = %v, want nil", err)
	}
	want := key(gitcmd.PushDeleteBranchArgs(gitcmd.RefOrigin, "update-files-x"))
	if !r.Ran(want) {
		t.Errorf("Keys() = %v, want it to contain %q", r.Keys(), want)
	}
}

func TestNewClientWithBaseURL_TargetsGivenHost(t *testing.T) {
	api := newFakeAPI(t).route("POST /pulls", http.StatusCreated, `{"number":1}`)
	t.Setenv("GITHUB_REPOSITORY", "owner/repo")

	client := github.NewClientWithBaseURL("token", api.server.URL)
	if _, err := client.Post(context.Background(), "/pulls", map[string]string{"a": "b"}); err != nil {
		t.Fatalf("Post() error = %v, want nil", err)
	}
	if _, ok := api.called("POST /pulls"); !ok {
		t.Errorf("Calls() = %v, want the request to reach the injected host", api.Calls())
	}
	if client.Repo() != "owner/repo" {
		t.Errorf("Repo() = %q, want %q", client.Repo(), "owner/repo")
	}
}
