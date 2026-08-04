package git

import (
	"context"
	"strings"
	"testing"

	"github.com/somaz94/go-git-commit-action/internal/config"
	"github.com/somaz94/go-git-commit-action/internal/gitcmd"
	"github.com/somaz94/go-git-commit-action/internal/output"
)

// key renders a git invocation the way FakeRunner records it, so expectations
// are written in terms of the argument builders rather than literal strings.
func key(args []string) string {
	return gitcmd.Call{Name: gitcmd.CmdGit, Args: args}.Key()
}

// baseConfig returns a config that keeps the workflow on the direct-commit path
// and out of any working-directory change.
func baseConfig() *config.GitConfig {
	return &config.GitConfig{
		UserEmail:     "bot@example.com",
		UserName:      "bot",
		CommitMessage: "chore: auto commit",
		Branch:        "main",
		RepoPath:      ".",
		FilePattern:   ".",
		PRBase:        "main",
		RetryCount:    1,
		Timeout:       30,
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

func TestExecuteCommandBatch_WithFakeRunner(t *testing.T) {
	f := gitcmd.NewFakeRunner()
	commands := []Command{
		{Name: "git", Args: []string{"status"}, Desc: "Status"},
		{Name: "git", Args: []string{"log"}, Desc: "Log"},
	}

	if err := ExecuteCommandBatch(f, commands, "header"); err != nil {
		t.Fatalf("ExecuteCommandBatch() error = %v, want nil", err)
	}
	assertSequence(t, f.Keys(), []string{"git status", "git log"})
}

func TestExecuteCommandBatch_SkipsNothingToCommit(t *testing.T) {
	f := gitcmd.NewFakeRunner().
		Stub("git commit -m msg", gitcmd.FakeResult{Err: gitcmd.Fail(1)})
	commands := []Command{
		{Name: "git", Args: []string{"commit", "-m", "msg"}, Desc: "Commit"},
		{Name: "git", Args: []string{"push"}, Desc: "Push"},
	}

	if err := ExecuteCommandBatch(f, commands, ""); err != nil {
		t.Fatalf("ExecuteCommandBatch() error = %v, want an empty commit to be skipped", err)
	}
	// The batch must continue to the push after skipping.
	if !f.Ran("git push") {
		t.Error("batch stopped after an empty commit, want it to continue")
	}
}

func TestExecuteCommandBatch_AbortsOnRealFailure(t *testing.T) {
	f := gitcmd.NewFakeRunner().
		Stub("git commit -m msg", gitcmd.FakeResult{Err: gitcmd.Fail(128)})
	commands := []Command{
		{Name: "git", Args: []string{"commit", "-m", "msg"}, Desc: "Commit"},
		{Name: "git", Args: []string{"push"}, Desc: "Push"},
	}

	if err := ExecuteCommandBatch(f, commands, ""); err == nil {
		t.Fatal("ExecuteCommandBatch() error = nil, want exit 128 to abort the batch")
	}
	if f.Ran("git push") {
		t.Error("batch continued past a real failure, want it to abort")
	}
}

func TestSetupGitConfig_EmitsExpectedCommands(t *testing.T) {
	f := gitcmd.NewFakeRunner()
	cfg := baseConfig()

	if err := setupGitConfig(f, cfg); err != nil {
		t.Fatalf("setupGitConfig() error = %v, want nil", err)
	}

	assertSequence(t, f.Keys(), []string{
		key(gitcmd.ConfigSafeDirArgs(gitcmd.PathApp)),
		key(gitcmd.ConfigSafeDirArgs(gitcmd.PathGitHubWorkspace)),
		key(gitcmd.ConfigUserEmailArgs(cfg.UserEmail)),
		key(gitcmd.ConfigUserNameArgs(cfg.UserName)),
		key(gitcmd.ConfigListArgs()),
	})
}

func TestSetupGitConfig_PropagatesFailure(t *testing.T) {
	f := gitcmd.NewFakeRunner().
		Stub(key(gitcmd.ConfigUserNameArgs("bot")), gitcmd.FakeResult{Err: gitcmd.Fail(128)})

	if err := setupGitConfig(f, baseConfig()); err == nil {
		t.Fatal("setupGitConfig() error = nil, want the config failure")
	}
}

func TestSetupGitCredentials_RewritesHTTPSRemote(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	f := gitcmd.NewFakeRunner().
		Stub(key(gitcmd.ConfigGetArgs("remote.origin.url")),
			gitcmd.FakeResult{Stdout: "https://github.com/owner/repo.git\n"})

	cfg := baseConfig()
	cfg.GitHubToken = "example-token"

	if err := setupGitCredentials(f, cfg); err != nil {
		t.Fatalf("setupGitCredentials() error = %v, want nil", err)
	}

	want := key(gitcmd.RemoteSetURLArgs(gitcmd.RefOrigin,
		"https://x-access-token:example-token@github.com/owner/repo.git"))
	if !f.Ran(want) {
		t.Errorf("Keys() = %v, want it to contain %q", f.Keys(), want)
	}
}

func TestSetupGitCredentials_SkipsWithoutToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	f := gitcmd.NewFakeRunner()

	if err := setupGitCredentials(f, baseConfig()); err != nil {
		t.Fatalf("setupGitCredentials() error = %v, want nil", err)
	}
	if len(f.Calls()) != 0 {
		t.Errorf("Calls() = %v, want no commands when no token is available", f.Calls())
	}
}

func TestSetupGitCredentials_SkipsNonGitHubRemote(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "example-token")
	f := gitcmd.NewFakeRunner().
		Stub(key(gitcmd.ConfigGetArgs("remote.origin.url")),
			gitcmd.FakeResult{Stdout: "https://gitlab.example.com/owner/repo.git\n"})

	if err := setupGitCredentials(f, baseConfig()); err != nil {
		t.Fatalf("setupGitCredentials() error = %v, want nil", err)
	}
	// A non-GitHub remote must never be rewritten.
	for _, k := range f.Keys() {
		if strings.Contains(k, gitcmd.OptSetURL) {
			t.Errorf("remote was rewritten for a non-GitHub URL: %q", k)
		}
	}
}

func TestSetupGitCredentials_SkipsSSHRemote(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "example-token")
	f := gitcmd.NewFakeRunner().
		Stub(key(gitcmd.ConfigGetArgs("remote.origin.url")),
			gitcmd.FakeResult{Stdout: "git@github.com:owner/repo.git\n"})

	if err := setupGitCredentials(f, baseConfig()); err != nil {
		t.Fatalf("setupGitCredentials() error = %v, want nil", err)
	}
	for _, k := range f.Keys() {
		if strings.Contains(k, gitcmd.OptSetURL) {
			t.Errorf("SSH remote was rewritten: %q", k)
		}
	}
}

func TestSetupGitCredentials_SkipsWhenRemoteLookupFails(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "example-token")
	f := &gitcmd.FakeRunner{Default: gitcmd.FakeResult{Err: gitcmd.Fail(128)}}

	if err := setupGitCredentials(f, baseConfig()); err != nil {
		t.Fatalf("setupGitCredentials() error = %v, want a missing remote to be tolerated", err)
	}
}

func TestHandleBranch_ExistingLocalBranchDoesNothing(t *testing.T) {
	// Both probes succeed → the branch is already checked out.
	f := gitcmd.NewFakeRunner()

	if err := handleBranch(f, baseConfig()); err != nil {
		t.Fatalf("handleBranch() error = %v, want nil", err)
	}
	if len(f.Calls()) != 2 {
		t.Errorf("Calls() = %v, want only the two existence probes", f.Keys())
	}
}

func TestHandleBranch_CreatesMissingBranch(t *testing.T) {
	cfg := baseConfig()
	cfg.Branch = "feature"
	// Neither the local nor the remote branch resolves; everything else succeeds.
	f := gitcmd.NewFakeRunner().
		Stub(key(gitcmd.RevParseArgs("feature")), gitcmd.FakeResult{Err: gitcmd.Fail(1)}).
		Stub(key(gitcmd.LsRemoteHeadsArgs(gitcmd.RefOrigin, "feature")), gitcmd.FakeResult{Err: gitcmd.Fail(2)})

	if err := handleBranch(f, cfg); err != nil {
		t.Fatalf("handleBranch() error = %v, want nil", err)
	}
	assertSequence(t, f.Keys(), []string{
		key(gitcmd.CheckoutNewBranchArgs("feature")),
		key(gitcmd.PushUpstreamArgs(gitcmd.RefOrigin, "feature")),
	})
}

func TestHandleBranch_ChecksOutRemoteOnlyBranch(t *testing.T) {
	cfg := baseConfig()
	cfg.Branch = "feature"
	// Local rev-parse fails, remote ls-remote succeeds.
	f := gitcmd.NewFakeRunner().
		Stub(key(gitcmd.RevParseArgs("feature")), gitcmd.FakeResult{Err: gitcmd.Fail(1)})

	if err := handleBranch(f, cfg); err != nil {
		t.Fatalf("handleBranch() error = %v, want nil", err)
	}
	assertSequence(t, f.Keys(), []string{
		key(gitcmd.StatusPorcelainArgs()),
		key(gitcmd.StashPushArgs()),
		key(gitcmd.FetchArgs(gitcmd.RefOrigin, "feature")),
		key(gitcmd.CheckoutArgs("feature")),
		key(gitcmd.ResetHardArgs("origin/feature")),
	})
}

func TestGetGitStatus(t *testing.T) {
	f := gitcmd.NewFakeRunner().
		Stub(key(gitcmd.StatusPorcelainArgs()), gitcmd.FakeResult{Stdout: " M a.txt\n"})

	got, err := getGitStatus(f)
	if err != nil {
		t.Fatalf("getGitStatus() error = %v, want nil", err)
	}
	if got != " M a.txt\n" {
		t.Errorf("getGitStatus() = %q, want the raw porcelain output", got)
	}
}

func TestGetGitStatus_Failure(t *testing.T) {
	f := &gitcmd.FakeRunner{Default: gitcmd.FakeResult{Err: gitcmd.Fail(128)}}

	if _, err := getGitStatus(f); err == nil {
		t.Fatal("getGitStatus() error = nil, want the status failure")
	}
}

func TestStashChanges(t *testing.T) {
	f := gitcmd.NewFakeRunner()

	if err := stashChanges(f); err != nil {
		t.Fatalf("stashChanges() error = %v, want nil", err)
	}
	if !f.Ran(key(gitcmd.StashPushArgs())) {
		t.Errorf("Keys() = %v, want a stash push", f.Keys())
	}
}

func TestStashChanges_Failure(t *testing.T) {
	f := &gitcmd.FakeRunner{Default: gitcmd.FakeResult{Err: gitcmd.Fail(1)}}

	if err := stashChanges(f); err == nil {
		t.Fatal("stashChanges() error = nil, want the stash failure")
	}
}

func TestFetchAndCheckout(t *testing.T) {
	cfg := baseConfig()
	cfg.Branch = "release"
	f := gitcmd.NewFakeRunner()

	if err := fetchAndCheckout(f, cfg); err != nil {
		t.Fatalf("fetchAndCheckout() error = %v, want nil", err)
	}
	assertSequence(t, f.Keys(), []string{
		key(gitcmd.FetchArgs(gitcmd.RefOrigin, "release")),
		key(gitcmd.CheckoutArgs("release")),
		key(gitcmd.ResetHardArgs("origin/release")),
	})
}

func TestCheckIfEmpty(t *testing.T) {
	tests := []struct {
		name        string
		status      string
		skipIfEmpty bool
		want        bool
	}{
		{"changes present, skip enabled", " M a.txt\n", true, false},
		{"no changes, skip enabled", "", true, true},
		{"no changes, skip disabled", "", false, false},
		{"changes present, skip disabled", " M a.txt\n", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := baseConfig()
			cfg.SkipIfEmpty = tt.skipIfEmpty
			f := gitcmd.NewFakeRunner().
				Stub(key(gitcmd.StatusPorcelainArgs()), gitcmd.FakeResult{Stdout: tt.status})

			got, err := checkIfEmpty(f, cfg)
			if err != nil {
				t.Fatalf("checkIfEmpty() error = %v, want nil", err)
			}
			if got != tt.want {
				t.Errorf("checkIfEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

// A failing branch diff is informational only and must not abort the check.
func TestCheckIfEmpty_ToleratesDiffFailure(t *testing.T) {
	cfg := baseConfig()
	cfg.PRBranch = "feature"
	f := gitcmd.NewFakeRunner().
		Stub(key(gitcmd.StatusPorcelainArgs()), gitcmd.FakeResult{Stdout: " M a.txt\n"}).
		Stub(key(gitcmd.DiffNameOnlyArgs("origin/main", "feature")), gitcmd.FakeResult{Err: gitcmd.Fail(128)})

	got, err := checkIfEmpty(f, cfg)
	if err != nil {
		t.Fatalf("checkIfEmpty() error = %v, want a diff failure to be tolerated", err)
	}
	if got {
		t.Error("checkIfEmpty() = true, want false when local changes exist")
	}
}

func TestCheckIfEmpty_StatusFailure(t *testing.T) {
	f := &gitcmd.FakeRunner{Default: gitcmd.FakeResult{Err: gitcmd.Fail(128)}}

	if _, err := checkIfEmpty(f, baseConfig()); err == nil {
		t.Fatal("checkIfEmpty() error = nil, want the status failure")
	}
}

func TestCountChangedFiles(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   int
	}{
		{"empty", "", 0},
		{"single file", " M a.txt\n", 1},
		{"multiple files", " M a.txt\n?? b.txt\n D c.txt\n", 3},
		{"trailing blank lines ignored", " M a.txt\n\n\n", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := gitcmd.NewFakeRunner().
				Stub(key(gitcmd.StatusPorcelainArgs()), gitcmd.FakeResult{Stdout: tt.status})

			if got := countChangedFiles(f); got != tt.want {
				t.Errorf("countChangedFiles() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCountChangedFiles_FailureReturnsZero(t *testing.T) {
	f := &gitcmd.FakeRunner{Default: gitcmd.FakeResult{Err: gitcmd.Fail(128)}}

	if got := countChangedFiles(f); got != 0 {
		t.Errorf("countChangedFiles() = %d, want 0 when status fails", got)
	}
}

func TestCommitChanges_StagesCommitsPushesAndRecordsSHA(t *testing.T) {
	cfg := baseConfig()
	cfg.FilePattern = "a.txt b.txt"
	f := gitcmd.NewFakeRunner().
		Stub(key(gitcmd.RevParseArgs("HEAD")), gitcmd.FakeResult{Stdout: "cafebabe\n"})
	result := output.NewResult()

	if err := commitChanges(f, cfg, result); err != nil {
		t.Fatalf("commitChanges() error = %v, want nil", err)
	}

	assertSequence(t, f.Keys(), []string{
		key(gitcmd.AddArgs("a.txt")),
		key(gitcmd.AddArgs("b.txt")),
		key(gitcmd.CommitArgs(cfg.CommitMessage)),
		key(gitcmd.PushArgs(gitcmd.RefOrigin, cfg.Branch)),
		key(gitcmd.RevParseArgs("HEAD")),
	})
	if got := result.Get(output.KeyCommitSHA); got != "cafebabe" {
		t.Errorf("commit_sha output = %q, want %q", got, "cafebabe")
	}
}

func TestCommitChanges_StageFailureAborts(t *testing.T) {
	f := gitcmd.NewFakeRunner().
		Stub(key(gitcmd.AddArgs(".")), gitcmd.FakeResult{Err: gitcmd.Fail(128)})

	if err := commitChanges(f, baseConfig(), output.NewResult()); err == nil {
		t.Fatal("commitChanges() error = nil, want the staging failure")
	}
	if f.Ran(key(gitcmd.CommitArgs("chore: auto commit"))) {
		t.Error("commit ran after a staging failure, want it skipped")
	}
}

func TestStageFiles_DelegatesToShared(t *testing.T) {
	f := gitcmd.NewFakeRunner()

	if err := StageFiles(f, "x.txt"); err != nil {
		t.Fatalf("StageFiles() error = %v, want nil", err)
	}
	if !f.Ran(key(gitcmd.AddArgs("x.txt"))) {
		t.Errorf("Keys() = %v, want a git add for the pattern", f.Keys())
	}
}

func TestRunGitCommitWithRunner_SkipsWhenEmpty(t *testing.T) {
	cfg := baseConfig()
	cfg.SkipIfEmpty = true
	// Empty status → nothing to commit.
	f := gitcmd.NewFakeRunner()
	result := output.NewResult()

	if err := RunGitCommitWithRunner(context.Background(), f, cfg, result); err != nil {
		t.Fatalf("RunGitCommitWithRunner() error = %v, want nil", err)
	}
	if got := result.Get(output.KeySkipped); got != "true" {
		t.Errorf("skipped output = %q, want %q", got, "true")
	}
	if got := result.Get(output.KeyChangedFiles); got != "0" {
		t.Errorf("changed_files output = %q, want %q", got, "0")
	}
	if f.Ran(key(gitcmd.CommitArgs(cfg.CommitMessage))) {
		t.Error("a commit was issued on the skip path, want none")
	}
}

func TestRunGitCommitWithRunner_CommitsWhenChangesExist(t *testing.T) {
	cfg := baseConfig()
	cfg.SkipIfEmpty = true
	f := gitcmd.NewFakeRunner().
		Stub(key(gitcmd.StatusPorcelainArgs()), gitcmd.FakeResult{Stdout: " M a.txt\n"}).
		Stub(key(gitcmd.RevParseArgs("HEAD")), gitcmd.FakeResult{Stdout: "abc1234\n"})
	result := output.NewResult()

	if err := RunGitCommitWithRunner(context.Background(), f, cfg, result); err != nil {
		t.Fatalf("RunGitCommitWithRunner() error = %v, want nil", err)
	}
	if got := result.Get(output.KeySkipped); got != "false" {
		t.Errorf("skipped output = %q, want %q", got, "false")
	}
	if got := result.Get(output.KeyChangedFiles); got != "1" {
		t.Errorf("changed_files output = %q, want %q", got, "1")
	}
	if !f.Ran(key(gitcmd.CommitArgs(cfg.CommitMessage))) {
		t.Errorf("Keys() = %v, want a commit to be issued", f.Keys())
	}
}

func TestRunGitCommitWithRunner_InvalidConfigFails(t *testing.T) {
	cfg := baseConfig()
	cfg.CreatePR = true
	cfg.PRBranch = "" // create_pr without auto_branch requires pr_branch
	f := gitcmd.NewFakeRunner()

	if err := RunGitCommitWithRunner(context.Background(), f, cfg, output.NewResult()); err == nil {
		t.Fatal("RunGitCommitWithRunner() error = nil, want the validation failure")
	}
}

func TestRunGitCommitWithRunner_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := RunGitCommitWithRunner(ctx, gitcmd.NewFakeRunner(), baseConfig(), output.NewResult())
	if err == nil {
		t.Fatal("RunGitCommitWithRunner() error = nil, want the cancelled context to abort")
	}
}
