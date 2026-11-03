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
