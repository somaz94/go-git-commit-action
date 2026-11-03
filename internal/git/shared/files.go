package shared

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/somaz94/go-git-commit-action/internal/gitcmd"
)

// ErrPushAfterCommit reports a push that still failed after its commit
// succeeded. The work then lives only in the local HEAD, so callers must not
// rerun the commit step: a rerun finds a clean tree and reports success.
var ErrPushAfterCommit = errors.New("failed to push")

// pushRetryDelay is the linear backoff unit between push attempts.
var pushRetryDelay = time.Second

// RunStep executes a single command with the standard
// "  - <desc>... " → "Done" / "FAILED" progress feedback used across the
// git packages. On failure it prints "FAILED" and returns the raw error so
// the caller can wrap it with the appropriate typed error.
//
// Output streaming is the Runner's responsibility: gitcmd.ExecRunner forwards
// stdout and stderr to the console, while a fake records the call instead.
func RunStep(r gitcmd.Runner, desc, name string, args ...string) error {
	fmt.Printf("  - %s... ", desc)
	if err := r.Run(name, args...); err != nil {
		fmt.Println("FAILED")
		return err
	}
	fmt.Println("Done")
	return nil
}

// StageFiles adds the specified files to the Git staging area.
// It handles multiple file patterns separated by spaces.
func StageFiles(r gitcmd.Runner, filePattern string) error {
	fmt.Printf("  - Adding files... ")

	for _, pattern := range strings.Fields(filePattern) {
		if err := executeGitAdd(r, pattern); err != nil {
			fmt.Println("FAILED")
			return fmt.Errorf("failed to add pattern %s: %w", pattern, err)
		}
	}

	fmt.Println("Done")
	return nil
}

// CommitPushOptions configures CommitAndPush behavior.
type CommitPushOptions struct {
	// SetUpstream pushes with "-u" to set the upstream tracking reference
	// (used when pushing a freshly created branch).
	SetUpstream bool
	// TolerateNothingToCommit treats a "nothing to commit" result (git commit
	// exit code 1) as success and skips the push, instead of failing. Used on
	// the direct-commit path where an empty commit must not abort the action.
	TolerateNothingToCommit bool
	// PushAttempts is how many times the push is tried after a successful
	// commit; values below 1 mean a single attempt.
	PushAttempts int
	// RebaseOnReject rebases the commit onto the remote branch before a retry
	// when the remote has moved ahead of the commit's parent. It needs
	// PushAttempts of at least 2.
	RebaseOnReject bool
}

// isNothingToCommitExit reports whether err is a "git commit" exit-code-1
// failure, which git returns when there is nothing staged to commit.
func isNothingToCommitExit(err error) bool {
	code, ok := gitcmd.ExitCodeOf(err)
	return ok && code == 1
}

// CommitAndPush commits the staged changes and pushes them to the remote branch.
// Behavior is controlled by opts (upstream tracking, empty-commit tolerance and
// push attempts). A push that never succeeds returns ErrPushAfterCommit.
func CommitAndPush(ctx context.Context, r gitcmd.Runner, commitMessage, branch string, opts CommitPushOptions) error {
	var base string
	if opts.RebaseOnReject {
		// An unborn HEAD leaves base empty, and no rebase runs.
		base, _ = CurrentCommitSHA(r)
	}

	fmt.Printf("  - Committing changes... ")
	if err := r.Run(gitcmd.CmdGit, gitcmd.CommitArgs(commitMessage)...); err != nil {
		if opts.TolerateNothingToCommit && isNothingToCommitExit(err) {
			// Decided 2026-08-04 (43ff9a2): nothing committed means nothing to publish;
			// pushing anyway fails a behind-remote branch as non-fast-forward.
			fmt.Println("[WARN] Nothing to commit, skipping commit and push...")
			return nil
		}
		fmt.Println("FAILED")
		return fmt.Errorf("failed to commit: %w", err)
	}
	fmt.Println("Done")

	pushArgs := gitcmd.PushArgs(gitcmd.RefOrigin, branch)
	if opts.SetUpstream {
		pushArgs = gitcmd.PushUpstreamArgs(gitcmd.RefOrigin, branch)
	}
	var beforeRetry func() error
	if opts.RebaseOnReject {
		beforeRetry = func() error { return rebaseIfBehind(r, branch, base) }
	}
	return pushWithRetry(ctx, r, pushArgs, opts.PushAttempts, beforeRetry)
}

// rebaseIfBehind replays the local commit onto the remote branch when the remote has
// only moved forward from base, the commit's parent. Asking git about the histories,
// rather than reading the push's error text, works in any locale and leaves an auth
// failure alone. It returns an error only when a retry cannot help; when it cannot tell
// whether the checkout is behind, the plain retry decides.
func rebaseIfBehind(r gitcmd.Runner, branch, base string) error {
	// The push sends refs/heads/<branch>, so rebasing any other HEAD cannot help it.
	head, err := r.Output(gitcmd.CmdGit, gitcmd.CurrentBranchArgs()...)
	if err != nil || strings.TrimSpace(string(head)) != branch {
		return nil
	}
	// FETCH_HEAD, not origin/<branch>: a checkout's refspec may not track the branch.
	if err := RunStep(r, "Fetching remote branch", gitcmd.CmdGit, gitcmd.FetchArgs(gitcmd.RefOrigin, branch)...); err != nil {
		return nil
	}
	_, err = r.Output(gitcmd.CmdGit, gitcmd.IsAncestorArgs(gitcmd.RefFetchHead, gitcmd.RefHead)...)
	if code, ok := gitcmd.ExitCodeOf(err); !ok || code != 1 || base == "" {
		return nil
	}
	// Decided 2026-11-03: exit 1 above also means the remote was rewritten, and rebasing
	// then replays commits a force-push removed, without any force on our side.
	if _, err := r.Output(gitcmd.CmdGit, gitcmd.IsAncestorArgs(base, gitcmd.RefFetchHead)...); err != nil {
		return fmt.Errorf("remote %s no longer contains %s: its history was rewritten, or this job holds other unpushed commits; not rebasing", branch, base)
	}
	if err := RunStep(r, "Rebasing onto remote branch", gitcmd.CmdGit, gitcmd.RebaseArgs(gitcmd.RefFetchHead)...); err != nil {
		// A dirty tree refuses before starting, which leaves nothing to abort.
		if abortErr := RunStep(r, "Aborting rebase", gitcmd.CmdGit, gitcmd.RebaseAbortArgs()...); abortErr != nil {
			return fmt.Errorf("rebase onto %s: %w (no rebase to abort, or abort failed: %v)", branch, err, abortErr)
		}
		return fmt.Errorf("rebase onto %s: %w", branch, err)
	}
	return nil
}

// pushWithRetry runs the push up to attempts times with linear backoff, honoring
// ctx while it waits. beforeRetry, when set, runs before each retry.
func pushWithRetry(ctx context.Context, r gitcmd.Runner, pushArgs []string, attempts int, beforeRetry func() error) error {
	var err error
	for i := range max(attempts, 1) {
		if i > 0 {
			select {
			case <-ctx.Done():
				return fmt.Errorf("%w: %w", ErrPushAfterCommit, ctx.Err())
			case <-time.After(pushRetryDelay * time.Duration(i)):
			}
			if beforeRetry != nil {
				if rebaseErr := beforeRetry(); rebaseErr != nil {
					return fmt.Errorf("%w: %w", ErrPushAfterCommit, rebaseErr)
				}
			}
		}
		if err = RunStep(r, "Pushing changes", gitcmd.CmdGit, pushArgs...); err == nil {
			return nil
		}
	}
	return fmt.Errorf("%w: %w", ErrPushAfterCommit, err)
}

// CurrentCommitSHA retrieves the current HEAD commit SHA.
func CurrentCommitSHA(r gitcmd.Runner) (string, error) {
	out, err := r.Output(gitcmd.CmdGit, gitcmd.RevParseArgs("HEAD")...)
	if err != nil {
		return "", fmt.Errorf("get commit SHA: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// executeGitAdd executes the git add command for a specific pattern.
func executeGitAdd(r gitcmd.Runner, pattern string) error {
	return r.Run(gitcmd.CmdGit, gitcmd.AddArgs(pattern)...)
}
