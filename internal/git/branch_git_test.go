package git

import (
	"context"
	stderrors "errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/somaz94/go-git-commit-action/internal/config"
	"github.com/somaz94/go-git-commit-action/internal/gitcmd"
	"github.com/somaz94/go-git-commit-action/internal/output"
)

// These run real git: both defects reported success, so only the remote shows them.

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
	return strings.TrimSpace(string(out))
}

func repoWithRemote(t *testing.T) (work, bare string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	isolateGitConfigEnv(t)
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GITHUB_TOKEN", "")
	root := t.TempDir()
	bare = filepath.Join(root, "remote.git")
	runGit(t, root, "init", "--bare", "-b", "main", bare)
	work = filepath.Join(root, "work")
	runGit(t, root, "clone", "file://"+bare, work)
	runGit(t, work, "config", "user.email", "test@example.com")
	runGit(t, work, "config", "user.name", "Test")
	runGit(t, work, "checkout", "-q", "-b", "main")
	writeRepoFile(t, work, "a.txt", "1\n")
	runGit(t, work, "add", "a.txt")
	runGit(t, work, "commit", "-qm", "seed")
	runGit(t, work, "push", "-q", "origin", "main")
	t.Chdir(work)
	return work, bare
}

func writeRepoFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func actionConfig(over func(*config.GitConfig)) *config.GitConfig {
	cfg := baseConfig()
	over(cfg)
	return cfg
}

func TestDeleteTag_RealGit_CommitsAndPushesNothing(t *testing.T) {
	work, bare := repoWithRemote(t)
	runGit(t, work, "tag", "-a", "t1", "-m", "t1")
	runGit(t, work, "push", "-q", "origin", "t1")
	// What actually reached a commit in CI: a build artifact left in the tree.
	writeRepoFile(t, work, "coverage.out", "junk\n")
	before := runGit(t, bare, "rev-parse", "main")
	cfg := actionConfig(func(c *config.GitConfig) {
		c.TagName = "t1"
		c.DeleteTag = true
	})
	result := output.NewResult()

	if err := RunGitCommitWithRunner(context.Background(), gitcmd.NewExecRunner(), cfg, result); err != nil {
		t.Fatalf("RunGitCommitWithRunner() error = %v", err)
	}
	if err := NewTagManager(cfg).HandleGitTag(context.Background(), result); err != nil {
		t.Fatalf("HandleGitTag() error = %v", err)
	}

	if got := runGit(t, bare, "rev-parse", "main"); got != before {
		t.Errorf("remote main moved to %s, want it untouched: deleting a tag publishes nothing", got)
	}
	if got := runGit(t, bare, "tag", "-l", "t1"); got != "" {
		t.Errorf("remote tag %q still there, want it deleted", got)
	}
	if got := runGit(t, work, "status", "--porcelain"); got != "?? coverage.out" {
		t.Errorf("status = %q, want the file left untracked", got)
	}
	if got := result.Get(output.KeySkipped); got != "true" {
		t.Errorf("skipped output = %q, want true", got)
	}
}

func TestHeadNotOnBranch_RealGit_FailsInsteadOfLosingTheCommit(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, work string)
		want  string
	}{
		{"another branch", func(t *testing.T, work string) {
			runGit(t, work, "checkout", "-q", "-b", "test")
		}, "checked out on test"},
		{"detached HEAD", func(t *testing.T, work string) {
			runGit(t, work, "checkout", "-q", "--detach")
		}, "a detached HEAD"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			work, bare := repoWithRemote(t)
			tt.setup(t, work)
			writeRepoFile(t, work, "a.txt", "2\n")
			before := runGit(t, bare, "rev-parse", "main")
			headBefore := runGit(t, work, "rev-parse", "HEAD")

			err := RunGitCommitWithRunner(context.Background(), gitcmd.NewExecRunner(), baseConfig(), output.NewResult())
			if !stderrors.Is(err, errHeadNotOnBranch) || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want errHeadNotOnBranch naming %q", err, tt.want)
			}
			if got := runGit(t, work, "rev-parse", "HEAD"); got != headBefore {
				t.Errorf("HEAD moved to %s, want no commit made", got)
			}
			if got := runGit(t, bare, "rev-parse", "main"); got != before {
				t.Errorf("remote main moved to %s", got)
			}
		})
	}
}

// A tag step on another branch with a clean tree loses nothing and must keep working.
func TestHeadNotOnBranch_RealGit_NothingStagedCarriesOn(t *testing.T) {
	work, bare := repoWithRemote(t)
	runGit(t, work, "checkout", "-q", "-b", "test")
	before := runGit(t, bare, "rev-parse", "main")

	if err := RunGitCommitWithRunner(context.Background(), gitcmd.NewExecRunner(), baseConfig(), output.NewResult()); err != nil {
		t.Fatalf("RunGitCommitWithRunner() error = %v, want a clean tree to pass", err)
	}
	if got := runGit(t, bare, "rev-parse", "main"); got != before {
		t.Errorf("remote main moved to %s", got)
	}
}

func TestHeadOnBranch_RealGit_StillPublishes(t *testing.T) {
	work, bare := repoWithRemote(t)
	writeRepoFile(t, work, "a.txt", "2\n")

	if err := RunGitCommitWithRunner(context.Background(), gitcmd.NewExecRunner(), baseConfig(), output.NewResult()); err != nil {
		t.Fatalf("RunGitCommitWithRunner() error = %v", err)
	}
	if got := runGit(t, bare, "log", "-1", "--format=%s", "main"); got != baseConfig().CommitMessage {
		t.Errorf("remote main = %q, want the commit published", got)
	}
}

// `symbolic-ref --short` prints heads/<branch> when a tag has the branch's name, which
// failed this check on a healthy checkout. A slash in the name must also survive.
func TestHeadOnBranch_RealGit_SameNamedTagAndSlashDoNotFailTheCheck(t *testing.T) {
	work, _ := repoWithRemote(t)
	runGit(t, work, "checkout", "-q", "-b", "feat/x")
	runGit(t, work, "tag", "feat/x")
	writeRepoFile(t, work, "a.txt", "2\n")
	runGit(t, work, "add", "a.txt")

	if err := ensureCommitReachesBranch(gitcmd.NewExecRunner(), "feat/x"); err != nil {
		t.Errorf("error = %v, want HEAD recognised on feat/x", err)
	}
	err := ensureCommitReachesBranch(gitcmd.NewExecRunner(), "x")
	if !stderrors.Is(err, errHeadNotOnBranch) || !strings.Contains(err.Error(), "checked out on feat/x,") {
		t.Errorf("error = %v, want a mismatch naming the plain branch", err)
	}
}

// These are all valid push sources, and were published before the check existed.
func TestHeadOnBranch_RealGit_EveryNameAPushAcceptsStillPublishes(t *testing.T) {
	for _, branch := range []string{"refs/heads/main", "heads/main", "HEAD"} {
		t.Run(branch, func(t *testing.T) {
			work, bare := repoWithRemote(t)
			writeRepoFile(t, work, "a.txt", "2\n")
			cfg := actionConfig(func(c *config.GitConfig) { c.Branch = branch })

			if err := RunGitCommitWithRunner(context.Background(), gitcmd.NewExecRunner(), cfg, output.NewResult()); err != nil {
				t.Fatalf("RunGitCommitWithRunner() error = %v", err)
			}
			if got := runGit(t, bare, "log", "-1", "--format=%s", "main"); got != cfg.CommitMessage {
				t.Errorf("remote main = %q, want the commit published", got)
			}
		})
	}
}

// An output that came back empty must not swallow the step's commit.
func TestDeleteTag_RealGit_WithoutATagNameStillCommits(t *testing.T) {
	work, bare := repoWithRemote(t)
	writeRepoFile(t, work, "a.txt", "2\n")
	cfg := actionConfig(func(c *config.GitConfig) { c.DeleteTag = true })

	if err := RunGitCommitWithRunner(context.Background(), gitcmd.NewExecRunner(), cfg, output.NewResult()); err != nil {
		t.Fatalf("RunGitCommitWithRunner() error = %v", err)
	}
	if got := runGit(t, bare, "log", "-1", "--format=%s", "main"); got != cfg.CommitMessage {
		t.Errorf("remote main = %q, want the commit published", got)
	}
}

// A tag is not a branch, so a push of it would not publish the commit on HEAD.
func TestHeadNotOnBranch_RealGit_TagNameIsAMismatch(t *testing.T) {
	work, _ := repoWithRemote(t)
	runGit(t, work, "tag", "rel")
	writeRepoFile(t, work, "a.txt", "2\n")
	runGit(t, work, "add", "a.txt")
	if err := ensureCommitReachesBranch(gitcmd.NewExecRunner(), "rel"); !stderrors.Is(err, errHeadNotOnBranch) {
		t.Errorf("error = %v, want errHeadNotOnBranch", err)
	}
}

func tagOnlyConfig(over func(*config.GitConfig)) *config.GitConfig {
	return actionConfig(func(c *config.GitConfig) {
		c.TagName = "v9"
		c.TagMessage = "v9"
		c.CommitMessage = config.DefaultCommitMessage
		c.FilePattern = config.DefaultFilePattern
		if over != nil {
			over(c)
		}
	})
}

func runWithTag(t *testing.T, cfg *config.GitConfig, result *output.Result) {
	t.Helper()
	if err := RunGitCommitWithRunner(context.Background(), gitcmd.NewExecRunner(), cfg, result); err != nil {
		t.Fatalf("RunGitCommitWithRunner() error = %v", err)
	}
	if err := NewTagManager(cfg).HandleGitTag(context.Background(), result); err != nil {
		t.Fatalf("HandleGitTag() error = %v", err)
	}
}

func TestTagOnly_RealGit_PublishesTheTagAndNothingElse(t *testing.T) {
	work, bare := repoWithRemote(t)
	writeRepoFile(t, work, "scratch.tmp", "junk\n")
	writeRepoFile(t, work, "a.txt", "changed\n")
	before := runGit(t, bare, "rev-parse", "main")
	result := output.NewResult()

	runWithTag(t, tagOnlyConfig(nil), result)

	if got := runGit(t, bare, "rev-parse", "main"); got != before {
		t.Errorf("remote main moved to %s, want it untouched", got)
	}
	if got := runGit(t, bare, "rev-parse", "v9^{commit}"); got != before {
		t.Errorf("tag points at %s, want %s", got, before)
	}
	// What a clean-tree tag step always reported, so steps reading them keep working.
	if got := result.Get(output.KeySkipped); got != "false" {
		t.Errorf("skipped output = %q, want false", got)
	}
	if got := result.Get(output.KeyCommitSHA); got != before {
		t.Errorf("commit_sha output = %q, want HEAD %s", got, before)
	}
	if got := result.Get(output.KeyChangedFiles); got != "0" {
		t.Errorf("changed_files output = %q, want 0", got)
	}
	if got := runGit(t, work, "status", "--porcelain"); !strings.Contains(got, "scratch.tmp") || !strings.Contains(got, "a.txt") {
		t.Errorf("status = %q, want the tree left as it was", got)
	}
}

func TestTagOnly_RealGit_TagReferenceWinsAndLeaksNothing(t *testing.T) {
	work, bare := repoWithRemote(t)
	first := runGit(t, work, "rev-parse", "HEAD")
	writeRepoFile(t, work, "a.txt", "2\n")
	runGit(t, work, "add", "a.txt")
	runGit(t, work, "commit", "-qm", "second")
	runGit(t, work, "push", "-q", "origin", "main")
	second := runGit(t, work, "rev-parse", "HEAD")
	writeRepoFile(t, work, "scratch.tmp", "junk\n")

	runWithTag(t, tagOnlyConfig(func(c *config.GitConfig) { c.TagReference = first }), output.NewResult())

	if got := runGit(t, bare, "rev-parse", "v9^{commit}"); got != first {
		t.Errorf("tag points at %s, want the reference %s", got, first)
	}
	if got := runGit(t, bare, "rev-parse", "main"); got != second {
		t.Errorf("remote main moved to %s", got)
	}
}

// Documented: a branch that exists nowhere is created, as it always was.
func TestTagOnly_RealGit_BranchMissingEverywhereIsCreated(t *testing.T) {
	work, bare := repoWithRemote(t)
	writeRepoFile(t, work, "scratch.tmp", "junk\n")

	runWithTag(t, tagOnlyConfig(func(c *config.GitConfig) { c.Branch = "brand-new" }), output.NewResult())

	if got := runGit(t, bare, "branch", "--list", "brand-new"); got == "" {
		t.Error("remote brand-new is missing, want it created")
	}
	if got := runGit(t, bare, "ls-tree", "--name-only", "brand-new"); strings.Contains(got, "scratch.tmp") {
		t.Errorf("scratch.tmp was published on the new branch:\n%s", got)
	}
}

func TestTagPlacementWarning_RealGit(t *testing.T) {
	work, _ := repoWithRemote(t)
	if got := tagPlacementWarning(gitcmd.NewExecRunner(), "main"); got != "" {
		t.Errorf("on the remote tip: warning = %q, want none", got)
	}
	writeRepoFile(t, work, "a.txt", "2\n")
	runGit(t, work, "add", "a.txt")
	runGit(t, work, "commit", "-qm", "unpushed")
	if got := tagPlacementWarning(gitcmd.NewExecRunner(), "main"); !strings.Contains(got, "not on origin/main") {
		t.Errorf("with an unpushed commit: warning = %q", got)
	}
	runGit(t, work, "checkout", "-q", "-b", "other")
	if got := tagPlacementWarning(gitcmd.NewExecRunner(), "main"); !strings.Contains(got, "HEAD is not on") {
		t.Errorf("on another branch: warning = %q", got)
	}
	runGit(t, work, "checkout", "-q", "--detach")
	if got := tagPlacementWarning(gitcmd.NewExecRunner(), "main"); !strings.Contains(got, "HEAD is not on") {
		t.Errorf("detached: warning = %q", got)
	}
}

// Naming what to commit keeps the commit-then-tag flow.
func TestTagOnly_RealGit_NamingACommitInputStillCommits(t *testing.T) {
	for name, over := range map[string]func(*config.GitConfig){
		"commit_message": func(c *config.GitConfig) { c.CommitMessage = "release" },
		"file_pattern":   func(c *config.GitConfig) { c.FilePattern = "a.txt" },
		"skip_if_empty":  func(c *config.GitConfig) { c.SkipIfEmpty = true },
	} {
		t.Run(name, func(t *testing.T) {
			work, bare := repoWithRemote(t)
			writeRepoFile(t, work, "a.txt", "changed\n")
			before := runGit(t, bare, "rev-parse", "main")

			runWithTag(t, tagOnlyConfig(over), output.NewResult())

			main := runGit(t, bare, "rev-parse", "main")
			if main == before {
				t.Error("remote main did not move, want the commit published")
			}
			if got := runGit(t, bare, "rev-parse", "v9^{commit}"); got != main {
				t.Errorf("tag points at %s, want the new commit %s", got, main)
			}
		})
	}
}

// Skipping the commit must not skip the checkout: `branch` picks the tagged commit.
func TestTagOnly_RealGit_BranchStillDecidesWhatIsTagged(t *testing.T) {
	work, bare := repoWithRemote(t)
	runGit(t, work, "checkout", "-q", "-b", "target")
	writeRepoFile(t, work, "b.txt", "1\n")
	runGit(t, work, "add", "b.txt")
	runGit(t, work, "commit", "-qm", "on target")
	runGit(t, work, "push", "-q", "origin", "target")
	tip := runGit(t, work, "rev-parse", "HEAD")
	runGit(t, work, "checkout", "-q", "main")
	runGit(t, work, "branch", "-qD", "target")

	runWithTag(t, tagOnlyConfig(func(c *config.GitConfig) { c.Branch = "target" }), output.NewResult())

	if got := runGit(t, bare, "rev-parse", "v9^{commit}"); got != tip {
		t.Errorf("tag points at %s, want the tip of target %s", got, tip)
	}
}

// A branch the action checks out itself ends with HEAD on it, so it must pass the check.
func TestHeadNotOnBranch_RealGit_RemoteOnlyBranchIsCheckedOut(t *testing.T) {
	work, bare := repoWithRemote(t)
	runGit(t, work, "checkout", "-q", "-b", "target")
	runGit(t, work, "push", "-q", "origin", "target")
	runGit(t, work, "checkout", "-q", "main")
	runGit(t, work, "branch", "-qD", "target")
	writeRepoFile(t, work, "a.txt", "2\n")
	cfg := actionConfig(func(c *config.GitConfig) { c.Branch = "target" })

	if err := RunGitCommitWithRunner(context.Background(), gitcmd.NewExecRunner(), cfg, output.NewResult()); err != nil {
		t.Fatalf("RunGitCommitWithRunner() error = %v", err)
	}
	if got := runGit(t, bare, "log", "-1", "--format=%s", "target"); got != cfg.CommitMessage {
		t.Errorf("remote target = %q, want the commit published", got)
	}
}

func TestHeadNotOnBranch_IsNotRetried(t *testing.T) {
	isolateGitConfigEnv(t)
	cfg := baseConfig()
	cfg.RetryCount = 3
	f, _ := pushingRepo(cfg, 0)
	inner := f.Handler
	f.Handler = func(name string, args []string) (string, error) {
		switch key(args) {
		case key(gitcmd.CurrentBranchArgs()):
			return gitcmd.BranchRef("test") + "\n", nil
		case key(gitcmd.StagedNamesArgs()):
			return "a.txt\n", nil
		}
		return inner(name, args)
	}

	err := RunGitCommitWithRunner(context.Background(), f, cfg, output.NewResult())
	if !stderrors.Is(err, errHeadNotOnBranch) {
		t.Fatalf("error = %v, want errHeadNotOnBranch", err)
	}
	probes := 0
	for _, k := range f.Keys() {
		if k == (gitcmd.Call{Name: gitcmd.CmdGit, Args: gitcmd.StagedNamesArgs()}).Key() {
			probes++
		}
	}
	if probes != 1 {
		t.Errorf("checked %d times, want one: a rerun finds the same branch", probes)
	}
}

func TestEnsureCommitReachesBranch_UnknownStagedStateCarriesOn(t *testing.T) {
	f := gitcmd.NewFakeRunner()
	f.Handler = func(_ string, args []string) (string, error) {
		if key(args) == key(gitcmd.StagedNamesArgs()) {
			return "", gitcmd.Fail(128)
		}
		return "test\n", nil
	}
	if err := ensureCommitReachesBranch(f, "main"); err != nil {
		t.Errorf("error = %v, want the old path when the state is unknown", err)
	}
}
