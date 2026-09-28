package git

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/somaz94/go-git-commit-action/internal/config"
	"github.com/somaz94/go-git-commit-action/internal/git/pr"
	"github.com/somaz94/go-git-commit-action/internal/gitcmd"
	"github.com/somaz94/go-git-commit-action/internal/github"
	"github.com/somaz94/go-git-commit-action/internal/output"
)

// fakeGitHub models the part of the GitHub API the PR stage uses: a repeated
// POST /pulls gets 422 while the PR is open and opens a duplicate once it is
// closed. fail[route] answers that many 500s before the route works;
// lose[route] applies the request but still answers 502, a lost reply.
type fakeGitHub struct {
	mu     sync.Mutex
	fail   map[string]int
	lose   map[string]int
	hits   map[string]int
	opened int
	open   int
	head   string
	base   string
}

func (g *fakeGitHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.mu.Lock()
	defer g.mu.Unlock()
	route := r.Method + " " + strings.TrimPrefix(r.URL.Path, "/repos/owner/repo")
	g.hits[route]++
	w.Header().Set("Content-Type", "application/json")
	if g.fail[route] > 0 {
		g.fail[route]--
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"message":"Server Error"}`)
		return
	}
	switch {
	case route == "POST /pulls" && g.open != 0:
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, `{"message":"Validation Failed","errors":[{"message":"A pull request already exists"}]}`)
	case route == "POST /pulls":
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		g.head, _ = body["head"].(string)
		g.base, _ = body["base"].(string)
		g.opened++
		g.open = 6 + g.opened
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{"html_url":"https://github.com/owner/repo/pull/%d","number":%d}`, g.open, g.open)
	case route == "GET /pulls" && g.open != 0:
		fmt.Fprintf(w, `[{"number":%d,"head":{"ref":%q,"repo":{"full_name":"owner/repo"}},"base":{"ref":%q}}]`,
			g.open, g.head, g.base)
	case route == "GET /pulls":
		fmt.Fprint(w, `[]`)
	case r.Method == http.MethodPatch:
		g.open = 0
		if g.lose[route] > 0 {
			g.lose[route]--
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprint(w, `{"message":"Bad Gateway"}`)
			return
		}
		fmt.Fprint(w, `{}`)
	default:
		fmt.Fprint(w, `{}`)
	}
}

func (g *fakeGitHub) count(route string) int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.hits[route]
}

func (g *fakeGitHub) openedPRs() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.opened
}

// useFakeGitHub points the PR stage at a fakeGitHub and drops the retry backoff.
func useFakeGitHub(t *testing.T, fail map[string]int) *fakeGitHub {
	t.Helper()
	g := &fakeGitHub{fail: fail, hits: map[string]int{}}
	srv := httptest.NewServer(g)
	t.Cleanup(srv.Close)
	t.Setenv("GITHUB_REPOSITORY", "owner/repo")
	t.Setenv("GITHUB_TOKEN", "")

	prevCreator, prevDelay := newPRCreator, retryBaseDelay
	newPRCreator = func(cfg *config.GitConfig, r gitcmd.Runner) *pr.Creator {
		return pr.NewCreatorWithClient(cfg, r, github.NewClientWithBaseURL(cfg.GitHubToken, srv.URL))
	}
	retryBaseDelay = 0
	t.Cleanup(func() { newPRCreator, retryBaseDelay = prevCreator, prevDelay })
	return g
}

// prRepo is a stateful fake for the PR flows: the tree is dirty until the
// first commit and clean after it, any branch diff lists a file, and a command
// whose key starts with a fail entry fails that many times first.
func prRepo(cfg *config.GitConfig, fail map[string]int) *gitcmd.FakeRunner {
	committed := false
	f := gitcmd.NewFakeRunner()
	f.Handler = func(_ string, args []string) (string, error) {
		k := key(args)
		for prefix, n := range fail {
			if n > 0 && strings.HasPrefix(k, prefix) {
				fail[prefix]--
				return "", gitcmd.Fail(128)
			}
		}
		switch {
		case k == key(gitcmd.StatusPorcelainArgs()):
			if committed {
				return "", nil
			}
			return " M a.txt\n", nil
		case k == key(gitcmd.CommitArgs(cfg.CommitMessage)):
			if committed {
				return "", gitcmd.Fail(1)
			}
			committed = true
		case args[0] == gitcmd.SubCmdDiff:
			return "M\ta.txt\n", nil
		case k == key(gitcmd.RevParseArgs("HEAD")):
			return "abc1234\n", nil
		}
		return "", nil
	}
	return f
}

// countPrefix counts the recorded commands whose key starts with prefix.
func countPrefix(f *gitcmd.FakeRunner, prefix string) int {
	n := 0
	for _, k := range f.Keys() {
		if strings.HasPrefix(k, prefix) {
			n++
		}
	}
	return n
}

func prStageConfig(autoBranch, skipIfEmpty bool) *config.GitConfig {
	cfg := baseConfig()
	cfg.CreatePR = true
	cfg.AutoBranch = autoBranch
	cfg.SkipIfEmpty = skipIfEmpty
	cfg.PRBranch = "feature"
	cfg.GitHubToken = "token"
	cfg.PRLabels = []string{"automated"}
	cfg.RetryCount = 3
	return cfg
}

// Regression: a PR step failing after the push reran the whole workflow on a
// clean tree, which reported a skip (skip_if_empty), zeroed changed_files, or
// cut a second auto branch whose empty commit failed.
func TestRunGitCommitWithRunner_PRStepFailureAfterPushRetriesInPlace(t *testing.T) {
	fetchBase := key(gitcmd.FetchArgs(gitcmd.RefOrigin, "main"))
	const labels = "POST /issues/7/labels"
	tests := []struct {
		name       string
		autoBranch bool
		skip       bool
		gitFail    map[string]int
		apiFail    map[string]int
	}{
		{"manual, fetch, skip_if_empty", false, true, map[string]int{fetchBase: 1}, nil},
		{"manual, fetch", false, false, map[string]int{fetchBase: 1}, nil},
		{"auto, fetch, skip_if_empty", true, true, map[string]int{fetchBase: 1}, nil},
		{"auto, fetch", true, false, map[string]int{fetchBase: 1}, nil},
		{"manual, create, skip_if_empty", false, true, nil, map[string]int{"POST /pulls": 1}},
		{"manual, labels, skip_if_empty", false, true, nil, map[string]int{labels: 1}},
		{"auto, labels", true, false, nil, map[string]int{labels: 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := prStageConfig(tt.autoBranch, tt.skip)
			wantLabelCalls := 1 + tt.apiFail[labels]
			api := useFakeGitHub(t, tt.apiFail)
			f := prRepo(cfg, tt.gitFail)
			result := output.NewResult()

			if err := RunGitCommitWithRunner(context.Background(), f, cfg, result); err != nil {
				t.Fatalf("RunGitCommitWithRunner() error = %v, want the PR step retried in place", err)
			}
			if got := countPrefix(f, key(gitcmd.ConfigListArgs())); got != 1 {
				t.Errorf("workflow ran %d times, want 1", got)
			}
			if got := countPrefix(f, key(gitcmd.CommitArgs(cfg.CommitMessage))); got != 1 {
				t.Errorf("commits = %d, want 1", got)
			}
			if got := countPrefix(f, key(gitcmd.CheckoutNewBranchArgs(""))); tt.autoBranch && got != 1 {
				t.Errorf("auto branches cut = %d, want 1", got)
			}
			if got := api.openedPRs(); got != 1 {
				t.Errorf("PRs opened = %d, want 1", got)
			}
			if got := api.count(labels); got != wantLabelCalls {
				t.Errorf("label calls = %d, want %d", got, wantLabelCalls)
			}
			for k, want := range map[string]string{
				output.KeySkipped:      "false",
				output.KeyChangedFiles: "1",
				output.KeyCommitSHA:    "abc1234",
				output.KeyPRURL:        "https://github.com/owner/repo/pull/7",
				output.KeyPRNumber:     "7",
			} {
				if got := result.Get(k); got != want {
					t.Errorf("%s output = %q, want %q", k, got, want)
				}
			}
		})
	}
}

// A retry after the PR exists must neither POST again (the closed PR would get
// a duplicate) nor repeat the follow-ups that already succeeded.
func TestRunGitCommitWithRunner_PRStepRetryResumesAfterCreation(t *testing.T) {
	cfg := prStageConfig(true, true)
	cfg.PRClosed = true
	cfg.DeleteSourceBranch = true
	deleteBranch := key(gitcmd.PushDeleteBranchArgs(gitcmd.RefOrigin, ""))
	api := useFakeGitHub(t, nil)
	f := prRepo(cfg, map[string]int{deleteBranch: 1})
	result := output.NewResult()

	if err := RunGitCommitWithRunner(context.Background(), f, cfg, result); err != nil {
		t.Fatalf("RunGitCommitWithRunner() error = %v, want the branch deletion retried", err)
	}
	if got := api.openedPRs(); got != 1 {
		t.Errorf("PRs opened = %d, want 1", got)
	}
	for route, want := range map[string]int{
		"POST /pulls":           1,
		"POST /issues/7/labels": 1,
		"PATCH /pulls/7":        1,
	} {
		if got := api.count(route); got != want {
			t.Errorf("%s calls = %d, want %d", route, got, want)
		}
	}
	if got := countPrefix(f, deleteBranch); got != 2 {
		t.Errorf("branch deletions = %d, want 2 (one failure, one success)", got)
	}
	if got := result.Get(output.KeySkipped); got != "false" {
		t.Errorf("skipped output = %q, want %q", got, "false")
	}
}

// An already-open PR is closed but the close reply is lost: the retry must look
// the PR up again rather than POST, which would open a duplicate.
func TestRunGitCommitWithRunner_PRStepRetryDoesNotRepostAfterAlreadyExists(t *testing.T) {
	cfg := prStageConfig(false, false)
	cfg.PRClosed = true
	api := useFakeGitHub(t, nil)
	api.mu.Lock()
	api.open, api.head, api.base = 5, cfg.PRBranch, cfg.PRBase
	api.lose = map[string]int{"PATCH /pulls/5": 1}
	api.mu.Unlock()
	f := prRepo(cfg, nil)

	if err := RunGitCommitWithRunner(context.Background(), f, cfg, output.NewResult()); err != nil {
		t.Fatalf("RunGitCommitWithRunner() error = %v, want the retry to find the PR already closed", err)
	}
	if got := api.openedPRs(); got != 0 {
		t.Errorf("PRs opened = %d, want 0 (no duplicate of the closed PR)", got)
	}
	if got := api.count("POST /pulls"); got != 1 {
		t.Errorf("POST /pulls calls = %d, want 1", got)
	}
}

// A PR step that keeps failing after the push must fail the action, not rerun
// the workflow into a skip.
func TestRunGitCommitWithRunner_PersistentPRStepFailureFails(t *testing.T) {
	fetchBase := key(gitcmd.FetchArgs(gitcmd.RefOrigin, "main"))
	for _, skip := range []bool{true, false} {
		t.Run(fmt.Sprintf("skip_if_empty=%v", skip), func(t *testing.T) {
			cfg := prStageConfig(false, skip)
			useFakeGitHub(t, nil)
			f := prRepo(cfg, map[string]int{fetchBase: 99})

			err := RunGitCommitWithRunner(context.Background(), f, cfg, output.NewResult())
			if !errors.Is(err, errPRStage) {
				t.Fatalf("RunGitCommitWithRunner() error = %v, want errPRStage", err)
			}
			if got := countPrefix(f, key(gitcmd.ConfigListArgs())); got != 1 {
				t.Errorf("workflow ran %d times, want 1", got)
			}
			if got := countPrefix(f, fetchBase); got != cfg.RetryCount {
				t.Errorf("base fetches = %d, want %d", got, cfg.RetryCount)
			}
		})
	}
}

func TestWithRetry_StopsOnPRStage(t *testing.T) {
	calls := 0
	err := withRetry(context.Background(), 3, func() error {
		calls++
		return fmt.Errorf("create pull request: %w", errPRStage)
	})
	if !errors.Is(err, errPRStage) {
		t.Fatalf("withRetry() error = %v, want errPRStage", err)
	}
	if calls != 1 {
		t.Errorf("operation calls = %d, want 1", calls)
	}
}
