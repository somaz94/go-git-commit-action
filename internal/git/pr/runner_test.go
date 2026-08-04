package pr

import (
	"strings"
	"testing"

	"github.com/somaz94/go-git-commit-action/internal/config"
	"github.com/somaz94/go-git-commit-action/internal/gitcmd"
)

// key renders a git invocation the way FakeRunner records it, so expectations
// are written in terms of the argument builders rather than literal strings.
func key(args []string) string {
	return gitcmd.Call{Name: gitcmd.CmdGit, Args: args}.Key()
}

func prConfig() *config.GitConfig {
	return &config.GitConfig{
		CommitMessage: "chore: pr commit",
		FilePattern:   ".",
		PRBase:        "main",
		PRBranch:      "feature",
		GitHubToken:   "token",
	}
}

func TestPrepareSourceBranch_ChecksOutExistingBranch(t *testing.T) {
	f := gitcmd.NewFakeRunner()
	bm := NewBranchManagerWithRunner(prConfig(), f)

	got, err := bm.PrepareSourceBranch()
	if err != nil {
		t.Fatalf("PrepareSourceBranch() error = %v, want nil", err)
	}
	if got != "feature" {
		t.Errorf("PrepareSourceBranch() = %q, want %q", got, "feature")
	}
	if !f.Ran(key(gitcmd.CheckoutArgs("feature"))) {
		t.Errorf("Keys() = %v, want a checkout of the configured branch", f.Keys())
	}
	// The existing-branch path must not create anything.
	if f.Ran(key(gitcmd.CheckoutNewBranchArgs("feature"))) {
		t.Error("a new branch was created, want the existing branch checked out")
	}
}

func TestPrepareSourceBranch_CheckoutFailure(t *testing.T) {
	f := &gitcmd.FakeRunner{Default: gitcmd.FakeResult{Err: gitcmd.Fail(1)}}
	bm := NewBranchManagerWithRunner(prConfig(), f)

	if _, err := bm.PrepareSourceBranch(); err == nil {
		t.Fatal("PrepareSourceBranch() error = nil, want the checkout failure")
	}
}

func TestPrepareSourceBranch_CreatesAutoBranch(t *testing.T) {
	cfg := prConfig()
	cfg.AutoBranch = true
	f := gitcmd.NewFakeRunner()
	bm := NewBranchManagerWithRunner(cfg, f)

	got, err := bm.PrepareSourceBranch()
	if err != nil {
		t.Fatalf("PrepareSourceBranch() error = %v, want nil", err)
	}
	if !strings.HasPrefix(got, "update-files-") {
		t.Errorf("PrepareSourceBranch() = %q, want an update-files-<timestamp> name", got)
	}
	// The generated name must be written back to the config for the PR body.
	if cfg.PRBranch != got {
		t.Errorf("cfg.PRBranch = %q, want it updated to %q", cfg.PRBranch, got)
	}

	wantSeq := []string{
		key(gitcmd.CheckoutNewBranchArgs(got)),
		key(gitcmd.AddArgs(".")),
		key(gitcmd.CommitArgs(cfg.CommitMessage)),
		key(gitcmd.PushUpstreamArgs(gitcmd.RefOrigin, got)),
	}
	assertSequence(t, f.Keys(), wantSeq)
}

func TestPrepareSourceBranch_AutoBranchCreateFailure(t *testing.T) {
	cfg := prConfig()
	cfg.AutoBranch = true
	f := &gitcmd.FakeRunner{Default: gitcmd.FakeResult{Err: gitcmd.Fail(1)}}
	bm := NewBranchManagerWithRunner(cfg, f)

	if _, err := bm.PrepareSourceBranch(); err == nil {
		t.Fatal("PrepareSourceBranch() error = nil, want the branch creation failure")
	}
}

func TestFetchBranches_FetchesBaseThenSource(t *testing.T) {
	f := gitcmd.NewFakeRunner()
	bm := NewBranchManagerWithRunner(prConfig(), f)

	if err := bm.FetchBranches(); err != nil {
		t.Fatalf("FetchBranches() error = %v, want nil", err)
	}
	assertSequence(t, f.Keys(), []string{
		key(gitcmd.FetchArgs(gitcmd.RefOrigin, "main")),
		key(gitcmd.FetchArgs(gitcmd.RefOrigin, "feature")),
	})
}

func TestFetchBranches_Failure(t *testing.T) {
	f := gitcmd.NewFakeRunner().
		Stub(key(gitcmd.FetchArgs(gitcmd.RefOrigin, "main")), gitcmd.FakeResult{Err: gitcmd.Fail(128)})
	bm := NewBranchManagerWithRunner(prConfig(), f)

	err := bm.FetchBranches()
	if err == nil {
		t.Fatal("FetchBranches() error = nil, want the fetch failure")
	}
	if !strings.Contains(err.Error(), "base branch") {
		t.Errorf("error = %q, want it to name the failing fetch", err.Error())
	}
	// The second fetch must not run after the first fails.
	if f.Ran(key(gitcmd.FetchArgs(gitcmd.RefOrigin, "feature"))) {
		t.Error("the source fetch ran after the base fetch failed, want it skipped")
	}
}

func TestDeleteSourceBranch_PushesDeletion(t *testing.T) {
	cfg := prConfig()
	cfg.AutoBranch = true
	f := gitcmd.NewFakeRunner()
	bm := NewBranchManagerWithRunner(cfg, f)

	if err := bm.DeleteSourceBranch("update-files-x"); err != nil {
		t.Fatalf("DeleteSourceBranch() error = %v, want nil", err)
	}
	want := key(gitcmd.PushDeleteBranchArgs(gitcmd.RefOrigin, "update-files-x"))
	if !f.Ran(want) {
		t.Errorf("Keys() = %v, want it to contain %q", f.Keys(), want)
	}
}

func TestDeleteSourceBranch_Failure(t *testing.T) {
	cfg := prConfig()
	cfg.AutoBranch = true
	f := &gitcmd.FakeRunner{Default: gitcmd.FakeResult{Err: gitcmd.Fail(1)}}
	bm := NewBranchManagerWithRunner(cfg, f)

	if err := bm.DeleteSourceBranch("update-files-x"); err == nil {
		t.Fatal("DeleteSourceBranch() error = nil, want the delete failure")
	}
}

func TestCheckBranchDifferences_ReportsChangedFiles(t *testing.T) {
	f := gitcmd.NewFakeRunner().
		Stub(key(gitcmd.DiffNameStatusArgs("origin/main", "origin/feature")),
			gitcmd.FakeResult{Stdout: "M\ta.txt\nA\tb.txt\n"})
	dc := NewDiffCheckerWithRunner(prConfig(), f)

	if err := dc.CheckBranchDifferences(); err != nil {
		t.Fatalf("CheckBranchDifferences() error = %v, want nil", err)
	}
	assertSequence(t, f.Keys(), []string{
		key(gitcmd.FetchArgs(gitcmd.RefOrigin, "main")),
		key(gitcmd.FetchArgs(gitcmd.RefOrigin, "feature")),
		key(gitcmd.DiffNameStatusArgs("origin/main", "origin/feature")),
	})
}

// With no diff and skip_if_empty off, an empty PR is an error.
func TestCheckBranchDifferences_NoChangesFails(t *testing.T) {
	f := gitcmd.NewFakeRunner()
	dc := NewDiffCheckerWithRunner(prConfig(), f)

	err := dc.CheckBranchDifferences()
	if err == nil {
		t.Fatal("CheckBranchDifferences() error = nil, want an empty diff to fail")
	}
	if !strings.Contains(err.Error(), "no changes") {
		t.Errorf("error = %q, want it to report no changes", err.Error())
	}
}

// With skip_if_empty on, an empty diff is tolerated.
func TestCheckBranchDifferences_NoChangesSkipped(t *testing.T) {
	cfg := prConfig()
	cfg.SkipIfEmpty = true
	dc := NewDiffCheckerWithRunner(cfg, gitcmd.NewFakeRunner())

	if err := dc.CheckBranchDifferences(); err != nil {
		t.Fatalf("CheckBranchDifferences() error = %v, want an empty diff to be skipped", err)
	}
}

func TestCheckBranchDifferences_FetchFailure(t *testing.T) {
	f := &gitcmd.FakeRunner{Default: gitcmd.FakeResult{Err: gitcmd.Fail(128)}}
	dc := NewDiffCheckerWithRunner(prConfig(), f)

	if err := dc.CheckBranchDifferences(); err == nil {
		t.Fatal("CheckBranchDifferences() error = nil, want the fetch failure")
	}
}

func TestNewConstructors_DefaultToExecRunner(t *testing.T) {
	cfg := prConfig()

	if NewBranchManager(cfg).runner == nil {
		t.Error("NewBranchManager() runner = nil, want a default ExecRunner")
	}
	if NewDiffChecker(cfg).runner == nil {
		t.Error("NewDiffChecker() runner = nil, want a default ExecRunner")
	}
	if NewCreator(cfg).runner == nil {
		t.Error("NewCreator() runner = nil, want a default ExecRunner")
	}
}

// assertSequence checks that got contains want as an ordered subsequence.
func assertSequence(t *testing.T, got, want []string) {
	t.Helper()
	i := 0
	for _, g := range got {
		if i < len(want) && g == want[i] {
			i++
		}
	}
	if i != len(want) {
		t.Errorf("command sequence\n got: %v\nwant (in order): %v", got, want)
	}
}
