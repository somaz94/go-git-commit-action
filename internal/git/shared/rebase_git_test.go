package shared

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/somaz94/go-git-commit-action/internal/gitcmd"
)

// These run real git, because what a fake cannot show is whether the rebase works in
// the shallow clone actions/checkout makes, where the shared history is cut off.

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
	return strings.TrimSpace(string(out))
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// racedCheckout returns a shallow checkout and a second clone that has pushed a commit
// touching otherFile after the checkout was made: the state a data job finds when
// someone pushes while it runs.
func racedCheckout(t *testing.T, otherFile string) (work, other string) {
	t.Helper()
	return racedCheckoutDepth(t, otherFile, true)
}

func racedCheckoutDepth(t *testing.T, otherFile string, shallow bool) (work, other string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	root := t.TempDir()
	bare := filepath.Join(root, "remote.git")
	git(t, root, "init", "--bare", "-b", "main", bare)

	clone := func(name string, extra ...string) string {
		dir := filepath.Join(root, name)
		args := append([]string{"clone"}, extra...)
		git(t, root, append(args, "file://"+bare, dir)...)
		git(t, dir, "config", "user.email", "test@example.com")
		git(t, dir, "config", "user.name", "Test")
		return dir
	}

	seed := clone("seed")
	git(t, seed, "checkout", "-b", "main")
	for i, name := range []string{"data.json", "code.go", "notes.md"} {
		writeFile(t, seed, name, "v1\n")
		git(t, seed, "add", name)
		git(t, seed, "commit", "-m", "seed "+string(rune('a'+i)))
	}
	git(t, seed, "push", "origin", "main")

	// file:// because a local path would ignore --depth.
	if shallow {
		work = clone("work", "--depth=1", "--branch", "main")
	} else {
		work = clone("work", "--branch", "main")
	}
	other = clone("other", "--branch", "main")
	writeFile(t, other, otherFile, "theirs\n")
	git(t, other, "add", otherFile)
	git(t, other, "commit", "-m", "theirs")
	git(t, other, "push", "origin", "main")
	return work, other
}

func commitAndPushIn(t *testing.T, dir string, opts CommitPushOptions) error {
	t.Helper()
	noPushDelay(t)
	t.Chdir(dir)
	return CommitAndPush(context.Background(), gitcmd.NewExecRunner(), "ours", "main", opts)
}

func TestRebaseOnReject_RealGit_ShallowCheckoutPublishesOnTopOfTheRemote(t *testing.T) {
	work, other := racedCheckout(t, "code.go")
	if got := git(t, work, "rev-parse", "--is-shallow-repository"); got != "true" {
		t.Fatalf("is-shallow-repository = %s, want the shape actions/checkout makes", got)
	}
	writeFile(t, work, "data.json", "ours\n")
	git(t, work, "add", "data.json")
	// An untracked file is what a data job's run leaves behind, and must not refuse it.
	writeFile(t, work, "scratch.tmp", "x\n")

	if err := commitAndPushIn(t, work, CommitPushOptions{PushAttempts: 3, RebaseOnReject: true}); err != nil {
		t.Fatalf("CommitAndPush() error = %v", err)
	}

	git(t, other, "pull", "--ff-only", "origin", "main")
	if got := git(t, other, "log", "-2", "--format=%s"); got != "ours\ntheirs" {
		t.Errorf("remote history = %q, want ours on top of theirs", got)
	}
	data, _ := os.ReadFile(filepath.Join(other, "code.go"))
	if string(data) != "theirs\n" {
		t.Errorf("remote code.go = %q, want their change kept", data)
	}
	if got := git(t, other, "show", "--stat", "--format=", "HEAD"); strings.Contains(got, "scratch.tmp") {
		t.Errorf("the commit carries scratch.tmp, which was never staged:\n%s", got)
	}
}

// No autostash: a stash that cannot be reapplied exits 0 and leaves conflict markers.
func TestRebaseOnReject_RealGit_DirtyTrackedFileFailsAndKeepsTheEdit(t *testing.T) {
	work, other := racedCheckout(t, "notes.md")
	remoteBefore := git(t, other, "rev-parse", "HEAD")
	writeFile(t, work, "data.json", "ours\n")
	git(t, work, "add", "data.json")
	writeFile(t, work, "notes.md", "uncommitted\n")

	err := commitAndPushIn(t, work, CommitPushOptions{PushAttempts: 3, RebaseOnReject: true})
	if !errors.Is(err, ErrPushAfterCommit) {
		t.Fatalf("CommitAndPush() error = %v, want ErrPushAfterCommit", err)
	}
	git(t, other, "fetch", "origin", "main")
	if got := git(t, other, "rev-parse", "FETCH_HEAD"); got != remoteBefore {
		t.Errorf("remote moved to %s, want it untouched", got)
	}
	if got := git(t, work, "status", "--porcelain"); got != "M notes.md" {
		t.Errorf("status = %q, want only the uncommitted edit", got)
	}
	if got := git(t, work, "stash", "list"); got != "" {
		t.Errorf("stash = %q, want nothing stashed", got)
	}
	if got := git(t, work, "log", "-1", "--format=%s"); got != "ours" {
		t.Errorf("HEAD = %q, want the local commit kept", got)
	}
}

// A force-push removed a commit: rebasing would replay it, and publish it back with
// no force on our side, undoing a cleanup such as removing a leaked secret.
func TestRebaseOnReject_RealGit_RewrittenRemoteIsNotUndone(t *testing.T) {
	for _, shallow := range []bool{true, false} {
		work, other := racedCheckoutDepth(t, "code.go", shallow)
		// Back past the commit that added notes.md, which is also the checkout's parent.
		git(t, other, "reset", "--hard", "HEAD~2")
		git(t, other, "commit", "--allow-empty", "-m", "removed the leak")
		git(t, other, "push", "--force", "origin", "main")
		rewritten := git(t, other, "rev-parse", "HEAD")

		writeFile(t, work, "data.json", "ours\n")
		git(t, work, "add", "data.json")

		err := commitAndPushIn(t, work, CommitPushOptions{PushAttempts: 3, RebaseOnReject: true})
		if !errors.Is(err, ErrPushAfterCommit) || !strings.Contains(err.Error(), "rewritten") {
			t.Fatalf("shallow=%v: error = %v, want a refusal naming the rewrite", shallow, err)
		}
		git(t, other, "fetch", "origin", "main")
		if got := git(t, other, "rev-parse", "FETCH_HEAD"); got != rewritten {
			t.Errorf("shallow=%v: remote moved to %s, want it left at the rewrite %s", shallow, got, rewritten)
		}
		if tree := git(t, other, "ls-tree", "--name-only", "FETCH_HEAD"); strings.Contains(tree, "notes.md") {
			t.Errorf("shallow=%v: notes.md came back:\n%s", shallow, tree)
		}
	}
}

// The push sends refs/heads/main, so rewriting whatever HEAD is on could not help it.
func TestRebaseOnReject_RealGit_OtherBranchIsLeftAlone(t *testing.T) {
	work, _ := racedCheckout(t, "code.go")
	git(t, work, "checkout", "-q", "-b", "feature")
	writeFile(t, work, "data.json", "ours\n")
	git(t, work, "add", "data.json")
	writeFile(t, work, "notes.md", "feature\n")
	git(t, work, "add", "notes.md")

	err := commitAndPushIn(t, work, CommitPushOptions{PushAttempts: 2, RebaseOnReject: true})
	if err == nil {
		t.Fatal("CommitAndPush() error = nil, want the stale main rejected")
	}
	if got := git(t, work, "rev-list", "--count", "HEAD"); got != "2" {
		t.Errorf("feature has %s commits, want its own 2, not rewritten onto the remote", got)
	}
}

func TestRebaseOnReject_RealGit_ConflictLeavesTheRemoteAndCheckoutClean(t *testing.T) {
	work, other := racedCheckout(t, "data.json")
	remoteBefore := git(t, other, "rev-parse", "HEAD")
	writeFile(t, work, "data.json", "ours\n")
	git(t, work, "add", "data.json")

	err := commitAndPushIn(t, work, CommitPushOptions{PushAttempts: 3, RebaseOnReject: true})
	if !errors.Is(err, ErrPushAfterCommit) {
		t.Fatalf("CommitAndPush() error = %v, want ErrPushAfterCommit", err)
	}

	git(t, other, "fetch", "origin", "main")
	if got := git(t, other, "rev-parse", "FETCH_HEAD"); got != remoteBefore {
		t.Errorf("remote moved to %s, want it untouched at %s", got, remoteBefore)
	}
	if got := git(t, work, "status", "--porcelain"); got != "" {
		t.Errorf("status = %q, want a clean tree after the abort", got)
	}
	if _, statErr := os.Stat(filepath.Join(work, ".git", "rebase-merge")); !os.IsNotExist(statErr) {
		t.Errorf("a rebase is still in progress: %v", statErr)
	}
	if got := git(t, work, "log", "-1", "--format=%s"); got != "ours" {
		t.Errorf("HEAD = %q, want the local commit kept after the abort", got)
	}
}

func TestRebaseOnReject_RealGit_OffStaysRejected(t *testing.T) {
	work, other := racedCheckout(t, "code.go")
	remoteBefore := git(t, other, "rev-parse", "HEAD")
	writeFile(t, work, "data.json", "ours\n")
	git(t, work, "add", "data.json")

	err := commitAndPushIn(t, work, CommitPushOptions{PushAttempts: 2})
	if !errors.Is(err, ErrPushAfterCommit) {
		t.Fatalf("CommitAndPush() error = %v, want ErrPushAfterCommit", err)
	}
	git(t, other, "fetch", "origin", "main")
	if got := git(t, other, "rev-parse", "FETCH_HEAD"); got != remoteBefore {
		t.Errorf("remote moved to %s, want it untouched", got)
	}
}
