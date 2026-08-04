package git

import (
	"context"
	"strings"
	"testing"

	"github.com/somaz94/go-git-commit-action/internal/config"
	"github.com/somaz94/go-git-commit-action/internal/gitcmd"
	"github.com/somaz94/go-git-commit-action/internal/output"
)

// tagConfig returns a config carrying a tag name and a single retry.
func tagConfig(tagName string) *config.GitConfig {
	cfg := baseConfig()
	cfg.TagName = tagName
	return cfg
}

func TestHandleGitTag_CreatesLightweightTag(t *testing.T) {
	f := gitcmd.NewFakeRunner()
	tm := NewTagManagerWithRunner(tagConfig("v1.2.3"), f)
	result := output.NewResult()

	if err := tm.HandleGitTag(context.Background(), result); err != nil {
		t.Fatalf("HandleGitTag() error = %v, want nil", err)
	}

	assertSequence(t, f.Keys(), []string{
		key(gitcmd.FetchTagsArgs()),
		key(gitcmd.TagCreateArgs("v1.2.3", true)),
		key(gitcmd.PushTagArgs("v1.2.3", true)),
	})
	if got := result.Get(output.KeyTagName); got != "v1.2.3" {
		t.Errorf("tag_name output = %q, want %q", got, "v1.2.3")
	}
}

func TestHandleGitTag_CreatesAnnotatedTag(t *testing.T) {
	cfg := tagConfig("v2.0.0")
	cfg.TagMessage = "release 2.0.0"
	f := gitcmd.NewFakeRunner()
	tm := NewTagManagerWithRunner(cfg, f)

	if err := tm.HandleGitTag(context.Background(), output.NewResult()); err != nil {
		t.Fatalf("HandleGitTag() error = %v, want nil", err)
	}

	want := key(gitcmd.TagCreateAnnotatedArgs("v2.0.0", "release 2.0.0", true))
	if !f.Ran(want) {
		t.Errorf("Keys() = %v, want it to contain %q", f.Keys(), want)
	}
}

func TestHandleGitTag_DeletesTag(t *testing.T) {
	cfg := tagConfig("v1.0.0")
	cfg.DeleteTag = true
	f := gitcmd.NewFakeRunner()
	tm := NewTagManagerWithRunner(cfg, f)
	result := output.NewResult()

	if err := tm.HandleGitTag(context.Background(), result); err != nil {
		t.Fatalf("HandleGitTag() error = %v, want nil", err)
	}

	assertSequence(t, f.Keys(), []string{
		key(gitcmd.FetchTagsArgs()),
		key(gitcmd.TagDeleteArgs("v1.0.0")),
		key(gitcmd.DeleteRemoteTagArgs("v1.0.0")),
	})
	// Deletion must not advertise a tag_name output.
	if got := result.Get(output.KeyTagName); got != "" {
		t.Errorf("tag_name output = %q, want empty on the delete path", got)
	}
	if f.Ran(key(gitcmd.TagCreateArgs("v1.0.0", true))) {
		t.Error("a tag was created on the delete path, want none")
	}
}

func TestHandleGitTag_FetchFailureAborts(t *testing.T) {
	f := gitcmd.NewFakeRunner().
		Stub(key(gitcmd.FetchTagsArgs()), gitcmd.FakeResult{Err: gitcmd.Fail(128)})
	tm := NewTagManagerWithRunner(tagConfig("v1.0.0"), f)

	if err := tm.HandleGitTag(context.Background(), output.NewResult()); err == nil {
		t.Fatal("HandleGitTag() error = nil, want the fetch failure")
	}
	if f.Ran(key(gitcmd.TagCreateArgs("v1.0.0", true))) {
		t.Error("tag creation ran after a fetch failure, want it skipped")
	}
}

func TestHandleGitTag_ResolvesTagReference(t *testing.T) {
	const sha = "0123456789abcdef0123456789abcdef01234567"
	cfg := tagConfig("v3.0.0")
	cfg.TagReference = "release-branch"
	f := gitcmd.NewFakeRunner().
		Stub(key(gitcmd.RevListArgs("release-branch")), gitcmd.FakeResult{Stdout: sha + "\n"})
	tm := NewTagManagerWithRunner(cfg, f)

	if err := tm.HandleGitTag(context.Background(), output.NewResult()); err != nil {
		t.Fatalf("HandleGitTag() error = %v, want nil", err)
	}

	assertSequence(t, f.Keys(), []string{
		key(gitcmd.RevParseArgs("release-branch")),
		key(gitcmd.RevListArgs("release-branch")),
	})
	// The resolved commit must be appended to the tag command.
	wantTag := key(append(gitcmd.TagCreateArgs("v3.0.0", true), sha))
	if !f.Ran(wantTag) {
		t.Errorf("Keys() = %v, want it to contain %q", f.Keys(), wantTag)
	}
}

func TestHandleGitTag_InvalidReferenceFails(t *testing.T) {
	cfg := tagConfig("v1.0.0")
	cfg.TagReference = "no-such-ref"
	f := gitcmd.NewFakeRunner().
		Stub(key(gitcmd.RevParseArgs("no-such-ref")), gitcmd.FakeResult{Err: gitcmd.Fail(128)})
	tm := NewTagManagerWithRunner(cfg, f)

	err := tm.HandleGitTag(context.Background(), output.NewResult())
	if err == nil {
		t.Fatal("HandleGitTag() error = nil, want the reference verification failure")
	}
	if !strings.Contains(err.Error(), "no-such-ref") {
		t.Errorf("error = %q, want it to name the bad reference", err.Error())
	}
}

func TestHandleGitTag_MalformedSHAFails(t *testing.T) {
	cfg := tagConfig("v1.0.0")
	cfg.TagReference = "weird-ref"
	f := gitcmd.NewFakeRunner().
		Stub(key(gitcmd.RevListArgs("weird-ref")), gitcmd.FakeResult{Stdout: "not-a-sha\n"})
	tm := NewTagManagerWithRunner(cfg, f)

	err := tm.HandleGitTag(context.Background(), output.NewResult())
	if err == nil {
		t.Fatal("HandleGitTag() error = nil, want a malformed SHA to be rejected")
	}
	if !strings.Contains(err.Error(), "invalid commit SHA") {
		t.Errorf("error = %q, want it to report an invalid SHA", err.Error())
	}
}

func TestHandleGitTag_RevListFailureFails(t *testing.T) {
	cfg := tagConfig("v1.0.0")
	cfg.TagReference = "some-ref"
	f := gitcmd.NewFakeRunner().
		Stub(key(gitcmd.RevListArgs("some-ref")), gitcmd.FakeResult{Err: gitcmd.Fail(128)})
	tm := NewTagManagerWithRunner(cfg, f)

	if err := tm.HandleGitTag(context.Background(), output.NewResult()); err == nil {
		t.Fatal("HandleGitTag() error = nil, want the rev-list failure")
	}
}

// An annotated tag pointing at a resolved commit must place the commit before
// the -m flag, which is where git expects it.
func TestHandleGitTag_AnnotatedWithReference(t *testing.T) {
	const sha = "89abcdef0123456789abcdef0123456789abcdef"
	cfg := tagConfig("v4.0.0")
	cfg.TagMessage = "annotated"
	cfg.TagReference = "main"
	f := gitcmd.NewFakeRunner().
		Stub(key(gitcmd.RevListArgs("main")), gitcmd.FakeResult{Stdout: sha + "\n"})
	tm := NewTagManagerWithRunner(cfg, f)

	if err := tm.HandleGitTag(context.Background(), output.NewResult()); err != nil {
		t.Fatalf("HandleGitTag() error = %v, want nil", err)
	}

	var tagCall string
	for _, k := range f.Keys() {
		if strings.Contains(k, "tag") && strings.Contains(k, "annotated") {
			tagCall = k
		}
	}
	if tagCall == "" {
		t.Fatalf("Keys() = %v, want an annotated tag command", f.Keys())
	}
	shaPos := strings.Index(tagCall, sha)
	msgPos := strings.Index(tagCall, gitcmd.OptMessage)
	if shaPos < 0 || msgPos < 0 || shaPos > msgPos {
		t.Errorf("tag command = %q, want the commit SHA to precede %q", tagCall, gitcmd.OptMessage)
	}
}

func TestHandleGitTag_RetriesOnFailure(t *testing.T) {
	cfg := tagConfig("v1.0.0")
	cfg.RetryCount = 3
	attempts := 0
	f := gitcmd.NewFakeRunner()
	f.Handler = func(name string, args []string) (string, error) {
		if len(args) > 0 && args[0] == gitcmd.SubCmdFetch {
			attempts++
			if attempts < 2 {
				return "", gitcmd.Fail(128)
			}
		}
		return "", nil
	}
	tm := NewTagManagerWithRunner(cfg, f)

	if err := tm.HandleGitTag(context.Background(), output.NewResult()); err != nil {
		t.Fatalf("HandleGitTag() error = %v, want the retry to succeed", err)
	}
	if attempts != 2 {
		t.Errorf("fetch attempts = %d, want 2 (one failure then one success)", attempts)
	}
}

func TestNewTagManager_DefaultsToExecRunner(t *testing.T) {
	tm := NewTagManager(tagConfig("v1.0.0"))
	if tm.runner == nil {
		t.Error("NewTagManager() runner = nil, want a default ExecRunner")
	}
}
