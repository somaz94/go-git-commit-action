package gitcmd

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestDiffNameOnlyArgs_ShallowClone runs the real diff in the kind of
// checkout Actions produces: depth 1, with main and the PR branch diverged.
func TestDiffNameOnlyArgs_ShallowClone(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(t.TempDir(), "gitconfig"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")

	root := t.TempDir()
	origin := filepath.Join(root, "origin.git")
	seed := filepath.Join(root, "seed")
	shallow := filepath.Join(root, "shallow")
	git := func(dir string, args ...string) string {
		t.Helper()
		full := append([]string{"-C", dir, "-c", "user.name=t", "-c", "user.email=t@example.com", "-c", "init.defaultBranch=main"}, args...)
		out, err := exec.Command("git", full...).CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return string(out)
	}

	git(root, "init", "-q", "--bare", origin)
	git(root, "init", "-q", seed)
	git(seed, "commit", "-q", "--allow-empty", "-m", "base")
	git(seed, "checkout", "-q", "-b", "feature")
	git(seed, "commit", "-q", "--allow-empty", "-m", "feature")
	git(seed, "checkout", "-q", "main")
	git(seed, "commit", "-q", "--allow-empty", "-m", "main")
	git(seed, "push", "-q", origin, "main", "feature")
	git(root, "clone", "-q", "--depth", "1", "--branch", "main", "file://"+origin, shallow)
	git(shallow, "fetch", "-q", "--depth", "1", "origin", "feature:feature")

	args := append([]string{"-C", shallow}, DiffNameOnlyArgs("origin/main", "feature")...)
	if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
		t.Fatalf("diff in a shallow clone failed: %v\n%s", err, strings.TrimSpace(string(out)))
	}
}
