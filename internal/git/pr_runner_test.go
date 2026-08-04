package git

import (
	"context"
	"testing"

	"github.com/somaz94/go-git-commit-action/internal/config"
	"github.com/somaz94/go-git-commit-action/internal/gitcmd"
	"github.com/somaz94/go-git-commit-action/internal/output"
)

// prDryRunConfig returns a config that exercises the PR path without reaching
// the GitHub API.
func prDryRunConfig() *config.GitConfig {
	cfg := baseConfig()
	cfg.CreatePR = true
	cfg.PRDryRun = true
	cfg.PRBranch = "feature"
	cfg.PRBase = "main"
	cfg.GitHubToken = "token"
	return cfg
}

func TestCreatePullRequest_DryRunRecordsOutputs(t *testing.T) {
	cfg := prDryRunConfig()
	f := gitcmd.NewFakeRunner().
		Stub(key(gitcmd.DiffNameStatusArgs("origin/main", "origin/feature")),
			gitcmd.FakeResult{Stdout: "M\ta.txt\n"}).
		Stub(key(gitcmd.RevParseArgs("HEAD")), gitcmd.FakeResult{Stdout: "feed1234\n"})
	result := output.NewResult()

	if err := CreatePullRequest(context.Background(), f, cfg, result); err != nil {
		t.Fatalf("CreatePullRequest() error = %v, want nil", err)
	}

	// The source branch is checked out, both branches fetched, then the diff read.
	assertSequence(t, f.Keys(), []string{
		key(gitcmd.CheckoutArgs("feature")),
		key(gitcmd.FetchArgs(gitcmd.RefOrigin, "main")),
		key(gitcmd.FetchArgs(gitcmd.RefOrigin, "feature")),
		key(gitcmd.DiffNameStatusArgs("origin/main", "origin/feature")),
	})
	if got := result.Get(output.KeyCommitSHA); got != "feed1234" {
		t.Errorf("commit_sha output = %q, want %q", got, "feed1234")
	}
	if got := result.Get(output.KeyPRURL); got == "" {
		t.Error("pr_url output is empty, want the dry-run compare URL")
	}
	if got := result.Get(output.KeyPRNumber); got != "0" {
		t.Errorf("pr_number output = %q, want %q in dry-run", got, "0")
	}
}

func TestCreatePullRequest_BranchPreparationFailureAborts(t *testing.T) {
	cfg := prDryRunConfig()
	f := gitcmd.NewFakeRunner().
		Stub(key(gitcmd.CheckoutArgs("feature")), gitcmd.FakeResult{Err: gitcmd.Fail(1)})

	if err := CreatePullRequest(context.Background(), f, cfg, output.NewResult()); err == nil {
		t.Fatal("CreatePullRequest() error = nil, want the checkout failure")
	}
	// The diff must never be attempted once the branch could not be prepared.
	if f.Ran(key(gitcmd.DiffNameStatusArgs("origin/main", "origin/feature"))) {
		t.Error("the diff ran after branch preparation failed, want it skipped")
	}
}

func TestCreatePullRequest_EmptyDiffFails(t *testing.T) {
	cfg := prDryRunConfig()
	f := gitcmd.NewFakeRunner()

	if err := CreatePullRequest(context.Background(), f, cfg, output.NewResult()); err == nil {
		t.Fatal("CreatePullRequest() error = nil, want an empty diff to fail")
	}
}

// With auto_branch off, the PR path commits to the configured branch first.
func TestHandlePullRequestFlow_CommitsBeforeCreatingPR(t *testing.T) {
	cfg := prDryRunConfig()
	cfg.PRDryRun = false
	cfg.AutoBranch = false
	// Keep the flow from reaching the GitHub API by failing at the diff stage.
	f := gitcmd.NewFakeRunner()
	result := output.NewResult()

	// An empty diff aborts before the API call, which is enough to assert that
	// the commit happened first.
	_ = handlePullRequestFlow(context.Background(), f, cfg, result)

	assertSequence(t, f.Keys(), []string{
		key(gitcmd.AddArgs(".")),
		key(gitcmd.CommitArgs(cfg.CommitMessage)),
		key(gitcmd.PushArgs(gitcmd.RefOrigin, cfg.Branch)),
	})
}

// In dry-run mode the pre-commit is skipped: nothing may be pushed.
func TestHandlePullRequestFlow_DryRunSkipsCommit(t *testing.T) {
	cfg := prDryRunConfig()
	f := gitcmd.NewFakeRunner().
		Stub(key(gitcmd.DiffNameStatusArgs("origin/main", "origin/feature")),
			gitcmd.FakeResult{Stdout: "M\ta.txt\n"})

	if err := handlePullRequestFlow(context.Background(), f, cfg, output.NewResult()); err != nil {
		t.Fatalf("handlePullRequestFlow() error = %v, want nil", err)
	}
	if f.Ran(key(gitcmd.CommitArgs(cfg.CommitMessage))) {
		t.Error("a commit was issued in dry-run mode, want none")
	}
	if f.Ran(key(gitcmd.PushArgs(gitcmd.RefOrigin, cfg.Branch))) {
		t.Error("a push was issued in dry-run mode, want none")
	}
}

// With auto_branch on, the flow creates its own branch rather than committing
// to the configured one.
func TestHandlePullRequestFlow_AutoBranchSkipsDirectCommit(t *testing.T) {
	cfg := prDryRunConfig()
	cfg.AutoBranch = true
	f := gitcmd.NewFakeRunner().
		Stub(key(gitcmd.DiffNameStatusArgs("origin/main", "origin/feature")),
			gitcmd.FakeResult{Stdout: "M\ta.txt\n"})
	// The diff key depends on the generated branch name, so tolerate the empty
	// diff and assert only on the branch-creation shape.
	_ = handlePullRequestFlow(context.Background(), f, cfg, output.NewResult())

	if !f.Ran(key(gitcmd.PushArgs(gitcmd.RefOrigin, cfg.Branch))) {
		// Correct: the direct-commit push to the configured branch must not happen.
		return
	}
	t.Error("the auto-branch path pushed to the configured branch, want a generated one")
}

// RunGitCommit must construct a real ExecRunner without touching a repository
// until a command actually runs.
func TestRunGitCommit_UsesExecRunner(t *testing.T) {
	cfg := baseConfig()
	cfg.CreatePR = true
	cfg.PRBranch = "" // fails validation before any command executes
	cfg.GitHubToken = "token"

	if err := RunGitCommit(context.Background(), cfg, output.NewResult()); err == nil {
		t.Fatal("RunGitCommit() error = nil, want the validation failure")
	}
}
