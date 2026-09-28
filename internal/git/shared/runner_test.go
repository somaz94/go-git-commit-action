package shared

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/somaz94/go-git-commit-action/internal/gitcmd"
)

// key renders a git invocation the way FakeRunner records it, so expectations
// are written in terms of the argument builders rather than literal strings.
func key(args []string) string {
	return gitcmd.Call{Name: gitcmd.CmdGit, Args: args}.Key()
}

func TestRunStep_Success(t *testing.T) {
	f := gitcmd.NewFakeRunner()

	if err := RunStep(f, "Doing a thing", gitcmd.CmdGit, gitcmd.StatusPorcelainArgs()...); err != nil {
		t.Fatalf("RunStep() error = %v, want nil", err)
	}
	if got, want := f.Keys(), key(gitcmd.StatusPorcelainArgs()); len(got) != 1 || got[0] != want {
		t.Errorf("Keys() = %v, want [%q]", got, want)
	}
}

func TestRunStep_PropagatesError(t *testing.T) {
	f := &gitcmd.FakeRunner{Default: gitcmd.FakeResult{Err: gitcmd.Fail(128)}}

	err := RunStep(f, "Doing a thing", gitcmd.CmdGit, "status")
	if err == nil {
		t.Fatal("RunStep() error = nil, want the runner failure")
	}
	if code, ok := gitcmd.ExitCodeOf(err); !ok || code != 128 {
		t.Errorf("ExitCodeOf() = (%d, %v), want (128, true)", code, ok)
	}
}

func TestStageFiles_IssuesOneAddPerPattern(t *testing.T) {
	f := gitcmd.NewFakeRunner()

	if err := StageFiles(f, "a.txt  b.txt   c.txt"); err != nil {
		t.Fatalf("StageFiles() error = %v, want nil", err)
	}

	want := []string{
		key(gitcmd.AddArgs("a.txt")),
		key(gitcmd.AddArgs("b.txt")),
		key(gitcmd.AddArgs("c.txt")),
	}
	got := f.Keys()
	if len(got) != len(want) {
		t.Fatalf("Keys() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Keys()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestStageFiles_StopsAtFirstFailure(t *testing.T) {
	f := gitcmd.NewFakeRunner().
		Stub(key(gitcmd.AddArgs("bad.txt")), gitcmd.FakeResult{Err: gitcmd.Fail(128)})

	err := StageFiles(f, "bad.txt good.txt")
	if err == nil {
		t.Fatal("StageFiles() error = nil, want the add failure")
	}
	if !strings.Contains(err.Error(), "bad.txt") {
		t.Errorf("error = %q, want it to name the failing pattern", err.Error())
	}
	// The second pattern must never be attempted.
	if f.Ran(key(gitcmd.AddArgs("good.txt"))) {
		t.Error("staging continued past a failing pattern, want it to stop")
	}
}

func TestStageFiles_NoPatternsIssuesNoCommands(t *testing.T) {
	f := gitcmd.NewFakeRunner()

	if err := StageFiles(f, "   "); err != nil {
		t.Fatalf("StageFiles() error = %v, want nil", err)
	}
	if len(f.Calls()) != 0 {
		t.Errorf("Calls() = %v, want no commands for a blank pattern", f.Calls())
	}
}

func TestCommitAndPush_CommitThenPush(t *testing.T) {
	f := gitcmd.NewFakeRunner()

	if err := CommitAndPush(context.Background(), f, "chore: msg", "main", CommitPushOptions{}); err != nil {
		t.Fatalf("CommitAndPush() error = %v, want nil", err)
	}

	want := []string{
		key(gitcmd.CommitArgs("chore: msg")),
		key(gitcmd.PushArgs(gitcmd.RefOrigin, "main")),
	}
	got := f.Keys()
	if len(got) != len(want) {
		t.Fatalf("Keys() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Keys()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCommitAndPush_SetUpstream(t *testing.T) {
	f := gitcmd.NewFakeRunner()

	if err := CommitAndPush(context.Background(), f, "msg", "feature", CommitPushOptions{SetUpstream: true}); err != nil {
		t.Fatalf("CommitAndPush() error = %v, want nil", err)
	}

	want := key(gitcmd.PushUpstreamArgs(gitcmd.RefOrigin, "feature"))
	if !f.Ran(want) {
		t.Errorf("Keys() = %v, want it to contain %q", f.Keys(), want)
	}
}

// An empty commit is tolerated, and because nothing was committed there is
// nothing to publish — the push must be skipped rather than attempted.
func TestCommitAndPush_TolerateNothingToCommit(t *testing.T) {
	// git exits 1 from "commit" when there is nothing staged.
	f := gitcmd.NewFakeRunner().
		Stub(key(gitcmd.CommitArgs("msg")), gitcmd.FakeResult{Err: gitcmd.Fail(1)})

	err := CommitAndPush(context.Background(), f, "msg", "main", CommitPushOptions{TolerateNothingToCommit: true})
	if err != nil {
		t.Fatalf("CommitAndPush() error = %v, want the empty commit tolerated", err)
	}
	if f.Ran(key(gitcmd.PushArgs(gitcmd.RefOrigin, "main"))) {
		t.Errorf("Keys() = %v, want no push after an empty commit", f.Keys())
	}
}

// A push that would have been rejected must never be reached once the commit
// was skipped, so a stale local branch cannot fail the action.
func TestCommitAndPush_EmptyCommitSkipsAFailingPush(t *testing.T) {
	f := gitcmd.NewFakeRunner().
		Stub(key(gitcmd.CommitArgs("msg")), gitcmd.FakeResult{Err: gitcmd.Fail(1)}).
		Stub(key(gitcmd.PushArgs(gitcmd.RefOrigin, "main")), gitcmd.FakeResult{Err: gitcmd.Fail(1)})

	if err := CommitAndPush(context.Background(), f, "msg", "main", CommitPushOptions{TolerateNothingToCommit: true}); err != nil {
		t.Fatalf("CommitAndPush() error = %v, want the rejected push never to be attempted", err)
	}
}

// The upstream variant follows the same rule.
func TestCommitAndPush_EmptyCommitSkipsUpstreamPush(t *testing.T) {
	f := gitcmd.NewFakeRunner().
		Stub(key(gitcmd.CommitArgs("msg")), gitcmd.FakeResult{Err: gitcmd.Fail(1)})

	if err := CommitAndPush(context.Background(), f, "msg", "feature", CommitPushOptions{
		SetUpstream: true, TolerateNothingToCommit: true,
	}); err != nil {
		t.Fatalf("CommitAndPush() error = %v, want nil", err)
	}
	if f.Ran(key(gitcmd.PushUpstreamArgs(gitcmd.RefOrigin, "feature"))) {
		t.Errorf("Keys() = %v, want no upstream push after an empty commit", f.Keys())
	}
}

func TestCommitAndPush_EmptyCommitFailsWhenNotTolerated(t *testing.T) {
	f := gitcmd.NewFakeRunner().
		Stub(key(gitcmd.CommitArgs("msg")), gitcmd.FakeResult{Err: gitcmd.Fail(1)})

	err := CommitAndPush(context.Background(), f, "msg", "main", CommitPushOptions{})
	if err == nil {
		t.Fatal("CommitAndPush() error = nil, want the commit failure")
	}
	if f.Ran(key(gitcmd.PushArgs(gitcmd.RefOrigin, "main"))) {
		t.Error("push ran after a failed commit, want it skipped")
	}
}

// Only exit code 1 means "nothing to commit"; any other failure is real even
// when the caller opted into tolerance.
func TestCommitAndPush_OtherExitCodeNotTolerated(t *testing.T) {
	f := gitcmd.NewFakeRunner().
		Stub(key(gitcmd.CommitArgs("msg")), gitcmd.FakeResult{Err: gitcmd.Fail(128)})

	err := CommitAndPush(context.Background(), f, "msg", "main", CommitPushOptions{TolerateNothingToCommit: true})
	if err == nil {
		t.Fatal("CommitAndPush() error = nil, want exit 128 to stay fatal")
	}
}

func TestCommitAndPush_PushFailure(t *testing.T) {
	f := gitcmd.NewFakeRunner().
		Stub(key(gitcmd.PushArgs(gitcmd.RefOrigin, "main")), gitcmd.FakeResult{Err: gitcmd.Fail(1)})

	err := CommitAndPush(context.Background(), f, "msg", "main", CommitPushOptions{})
	if err == nil {
		t.Fatal("CommitAndPush() error = nil, want the push failure")
	}
	if !strings.Contains(err.Error(), "push") {
		t.Errorf("error = %q, want it to mention the push", err.Error())
	}
	if !errors.Is(err, ErrPushAfterCommit) {
		t.Errorf("error = %v, want it to wrap ErrPushAfterCommit", err)
	}
}

// noPushDelay removes the push backoff for the duration of the test.
func noPushDelay(t *testing.T) {
	t.Helper()
	prev := pushRetryDelay
	pushRetryDelay = 0
	t.Cleanup(func() { pushRetryDelay = prev })
}

// countKey reports how many recorded calls match key.
func countKey(f *gitcmd.FakeRunner, key string) int {
	n := 0
	for _, k := range f.Keys() {
		if k == key {
			n++
		}
	}
	return n
}

// Only the push is retried; the commit it publishes must run exactly once.
func TestCommitAndPush_RetriesOnlyThePush(t *testing.T) {
	noPushDelay(t)
	tests := []struct {
		name       string
		failures   int
		attempts   int
		wantErr    bool
		wantPushes int
	}{
		{"transient failure recovers", 1, 3, false, 2},
		{"attempts exhausted", 5, 3, true, 3},
		{"zero attempts still pushes once", 0, 0, false, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pushes := 0
			f := gitcmd.NewFakeRunner()
			f.Handler = func(_ string, args []string) (string, error) {
				if args[0] != gitcmd.SubCmdPush {
					return "", nil
				}
				pushes++
				if pushes <= tt.failures {
					return "", gitcmd.Fail(1)
				}
				return "", nil
			}

			err := CommitAndPush(context.Background(), f, "msg", "main", CommitPushOptions{PushAttempts: tt.attempts})
			if (err != nil) != tt.wantErr {
				t.Fatalf("CommitAndPush() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && !errors.Is(err, ErrPushAfterCommit) {
				t.Errorf("error = %v, want it to wrap ErrPushAfterCommit", err)
			}
			if pushes != tt.wantPushes {
				t.Errorf("push attempts = %d, want %d", pushes, tt.wantPushes)
			}
			if n := countKey(f, key(gitcmd.CommitArgs("msg"))); n != 1 {
				t.Errorf("commit ran %d times, want 1", n)
			}
		})
	}
}

func TestCommitAndPush_CancelledDuringPushBackoff(t *testing.T) {
	prev := pushRetryDelay
	pushRetryDelay = time.Hour
	t.Cleanup(func() { pushRetryDelay = prev })
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	f := gitcmd.NewFakeRunner().
		Stub(key(gitcmd.PushArgs(gitcmd.RefOrigin, "main")), gitcmd.FakeResult{Err: gitcmd.Fail(1)})

	err := CommitAndPush(ctx, f, "msg", "main", CommitPushOptions{PushAttempts: 3})
	if !errors.Is(err, context.Canceled) || !errors.Is(err, ErrPushAfterCommit) {
		t.Fatalf("CommitAndPush() error = %v, want context.Canceled wrapped in ErrPushAfterCommit", err)
	}
	if n := countKey(f, key(gitcmd.PushArgs(gitcmd.RefOrigin, "main"))); n != 1 {
		t.Errorf("push attempts = %d, want 1 before the cancelled backoff", n)
	}
}

func TestCurrentCommitSHA(t *testing.T) {
	f := gitcmd.NewFakeRunner().
		Stub(key(gitcmd.RevParseArgs("HEAD")), gitcmd.FakeResult{Stdout: "  deadbeef\n"})

	sha, err := CurrentCommitSHA(f)
	if err != nil {
		t.Fatalf("CurrentCommitSHA() error = %v, want nil", err)
	}
	if sha != "deadbeef" {
		t.Errorf("CurrentCommitSHA() = %q, want %q (trimmed)", sha, "deadbeef")
	}
}

func TestCurrentCommitSHA_Failure(t *testing.T) {
	f := &gitcmd.FakeRunner{Default: gitcmd.FakeResult{Err: gitcmd.Fail(128)}}

	if _, err := CurrentCommitSHA(f); err == nil {
		t.Fatal("CurrentCommitSHA() error = nil, want the rev-parse failure")
	}
}

func TestIsNothingToCommitExit_FakeExitCodes(t *testing.T) {
	if !isNothingToCommitExit(gitcmd.Fail(1)) {
		t.Error("isNothingToCommitExit(exit 1) = false, want true")
	}
	if isNothingToCommitExit(gitcmd.Fail(2)) {
		t.Error("isNothingToCommitExit(exit 2) = true, want false")
	}
}
