package gitcmd

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

func TestExecRunner_Run_Success(t *testing.T) {
	var stdout, stderr bytes.Buffer
	r := &ExecRunner{Stdout: &stdout, Stderr: &stderr}

	if err := r.Run("echo", "hello"); err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if got := strings.TrimSpace(stdout.String()); got != "hello" {
		t.Errorf("stdout = %q, want %q", got, "hello")
	}
}

func TestExecRunner_Run_StreamsStderr(t *testing.T) {
	var stdout, stderr bytes.Buffer
	r := &ExecRunner{Stdout: &stdout, Stderr: &stderr}

	if err := r.Run("sh", "-c", "echo oops >&2"); err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if got := strings.TrimSpace(stderr.String()); got != "oops" {
		t.Errorf("stderr = %q, want %q", got, "oops")
	}
}

func TestExecRunner_Run_ExitError(t *testing.T) {
	r := &ExecRunner{}

	err := r.Run("sh", "-c", "exit 3")
	if err == nil {
		t.Fatal("Run() error = nil, want exit failure")
	}
	code, ok := ExitCodeOf(err)
	if !ok || code != 3 {
		t.Errorf("ExitCodeOf() = (%d, %v), want (3, true)", code, ok)
	}
}

func TestExecRunner_Output(t *testing.T) {
	r := NewExecRunner()

	out, err := r.Output("echo", "captured")
	if err != nil {
		t.Fatalf("Output() error = %v, want nil", err)
	}
	if got := strings.TrimSpace(string(out)); got != "captured" {
		t.Errorf("Output() = %q, want %q", got, "captured")
	}
}

// Output must not inherit the streaming Stderr writer: exec.Cmd.Output rejects
// a command whose Stderr is already assigned.
func TestExecRunner_Output_IgnoresStderrWriter(t *testing.T) {
	var stderr bytes.Buffer
	r := &ExecRunner{Stdout: &bytes.Buffer{}, Stderr: &stderr}

	if _, err := r.Output("echo", "ok"); err != nil {
		t.Fatalf("Output() error = %v, want nil", err)
	}
}

func TestExecRunner_Output_ExitError(t *testing.T) {
	r := NewExecRunner()

	if _, err := r.Output("sh", "-c", "exit 4"); err == nil {
		t.Fatal("Output() error = nil, want exit failure")
	} else if code, ok := ExitCodeOf(err); !ok || code != 4 {
		t.Errorf("ExitCodeOf() = (%d, %v), want (4, true)", code, ok)
	}
}

func TestExitError(t *testing.T) {
	cause := errors.New("underlying")
	e := &ExitError{Code: 1, Err: cause}

	if e.ExitCode() != 1 {
		t.Errorf("ExitCode() = %d, want 1", e.ExitCode())
	}
	if !errors.Is(e, cause) {
		t.Error("errors.Is() = false, want the wrapped cause to be reachable")
	}
	if !strings.Contains(e.Error(), "underlying") {
		t.Errorf("Error() = %q, want it to mention the cause", e.Error())
	}

	bare := &ExitError{Code: 2}
	if got := bare.Error(); got != "exit status 2" {
		t.Errorf("Error() = %q, want %q", got, "exit status 2")
	}
}

func TestExitCodeOf(t *testing.T) {
	realExit := exec.Command("sh", "-c", "exit 1").Run()

	tests := []struct {
		name     string
		err      error
		wantCode int
		wantOK   bool
	}{
		{"nil error", nil, 0, false},
		{"plain error", errors.New("boom"), 0, false},
		{"fake exit error", Fail(1), 1, true},
		{"fake exit error, other code", Fail(128), 128, true},
		{"real exec exit error", realExit, 1, true},
		{"wrapped fake exit error", fmt.Errorf("commit: %w", Fail(1)), 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, ok := ExitCodeOf(tt.err)
			if code != tt.wantCode || ok != tt.wantOK {
				t.Errorf("ExitCodeOf() = (%d, %v), want (%d, %v)", code, ok, tt.wantCode, tt.wantOK)
			}
		})
	}
}

func TestFakeRunner_RecordsCalls(t *testing.T) {
	f := NewFakeRunner()

	if err := f.Run("git", "status", "--porcelain"); err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if _, err := f.Output("git", "rev-parse", "HEAD"); err != nil {
		t.Fatalf("Output() error = %v, want nil", err)
	}

	want := []string{"git status --porcelain", "git rev-parse HEAD"}
	got := f.Keys()
	if len(got) != len(want) {
		t.Fatalf("Keys() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Keys()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
	if !f.Ran("git status --porcelain") {
		t.Error("Ran() = false for a recorded command, want true")
	}
	if f.Ran("git push origin main") {
		t.Error("Ran() = true for a command never issued, want false")
	}
}

func TestFakeRunner_StubbedResult(t *testing.T) {
	f := NewFakeRunner().
		Stub("git rev-parse HEAD", FakeResult{Stdout: "abc123\n"}).
		Stub("git commit -m msg", FakeResult{Err: Fail(1)})

	out, err := f.Output("git", "rev-parse", "HEAD")
	if err != nil {
		t.Fatalf("Output() error = %v, want nil", err)
	}
	if got := strings.TrimSpace(string(out)); got != "abc123" {
		t.Errorf("Output() = %q, want %q", got, "abc123")
	}

	err = f.Run("git", "commit", "-m", "msg")
	if code, ok := ExitCodeOf(err); !ok || code != 1 {
		t.Errorf("ExitCodeOf() = (%d, %v), want (1, true)", code, ok)
	}
}

func TestFakeRunner_Default(t *testing.T) {
	f := &FakeRunner{Default: FakeResult{Err: Fail(128)}}

	if err := f.Run("git", "anything"); err == nil {
		t.Error("Run() error = nil, want the default failure")
	}
	// Output propagates the error and returns no bytes.
	out, err := f.Output("git", "anything")
	if err == nil {
		t.Error("Output() error = nil, want the default failure")
	}
	if out != nil {
		t.Errorf("Output() = %v, want nil on error", out)
	}
}

func TestFakeRunner_Handler(t *testing.T) {
	// A stateful fake: status reports a change only after an add has run.
	f := NewFakeRunner()
	added := false
	f.Handler = func(name string, args []string) (string, error) {
		if len(args) > 0 && args[0] == "add" {
			added = true
			return "", nil
		}
		if added {
			return " M file.txt\n", nil
		}
		return "", nil
	}

	if out, _ := f.Output("git", "status", "--porcelain"); len(out) != 0 {
		t.Errorf("status before add = %q, want empty", out)
	}
	if err := f.Run("git", "add", "."); err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if out, _ := f.Output("git", "status", "--porcelain"); len(out) == 0 {
		t.Error("status after add = empty, want a change")
	}
}

func TestFakeRunner_Reset(t *testing.T) {
	f := NewFakeRunner().Stub("git status", FakeResult{Stdout: "x"})
	_ = f.Run("git", "status")

	f.Reset()

	if len(f.Calls()) != 0 {
		t.Errorf("Calls() = %v after Reset, want empty", f.Calls())
	}
	// Stubs survive a Reset.
	out, _ := f.Output("git", "status")
	if string(out) != "x" {
		t.Errorf("Output() = %q after Reset, want the stub to survive", out)
	}
}

func TestCall_Key(t *testing.T) {
	tests := []struct {
		name string
		call Call
		want string
	}{
		{"no args", Call{Name: "git"}, "git"},
		{"one arg", Call{Name: "git", Args: []string{"status"}}, "git status"},
		{"many args", Call{Name: "git", Args: []string{"commit", "-m", "msg"}}, "git commit -m msg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.call.Key(); got != tt.want {
				t.Errorf("Key() = %q, want %q", got, tt.want)
			}
		})
	}
}

// The recorded args must be a copy: mutating the caller's slice afterwards
// must not rewrite history.
func TestFakeRunner_CopiesArgs(t *testing.T) {
	f := NewFakeRunner()
	args := []string{"status", "--porcelain"}

	_ = f.Run("git", args...)
	args[0] = "mutated"

	if got := f.Keys()[0]; got != "git status --porcelain" {
		t.Errorf("Keys()[0] = %q, want the args recorded at call time", got)
	}
}
