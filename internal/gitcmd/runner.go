package gitcmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// Runner executes external commands. It is the single seam between the git
// packages and the operating system: production code uses ExecRunner, while
// tests substitute a fake so that command sequences can be asserted without a
// real git repository.
type Runner interface {
	// Run executes the command and streams its stdout/stderr to the runner's
	// configured writers. It returns an error wrapping the exit status when the
	// command runs but exits non-zero.
	Run(name string, args ...string) error

	// Output executes the command and returns its stdout. Stderr is not
	// captured, matching the semantics of exec.Cmd.Output.
	Output(name string, args ...string) ([]byte, error)
}

// ExecRunner is the production Runner, backed by os/exec.
type ExecRunner struct {
	// Stdout and Stderr receive the streamed output of Run. A nil writer
	// discards that stream.
	Stdout io.Writer
	Stderr io.Writer
}

// NewExecRunner returns an ExecRunner that streams to the process stdout and
// stderr, matching the behavior the action had before the Runner seam existed.
func NewExecRunner() *ExecRunner {
	return &ExecRunner{Stdout: os.Stdout, Stderr: os.Stderr}
}

// Run executes the command, streaming output to the configured writers.
func (r *ExecRunner) Run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = r.Stdout
	cmd.Stderr = r.Stderr
	return cmd.Run()
}

// Output executes the command and returns its stdout.
//
// Stderr is deliberately left unset: exec.Cmd.Output rejects a command whose
// Stderr is already assigned, so the streaming writers do not apply here.
func (r *ExecRunner) Output(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).Output()
}

// ExitError reports that a command ran to completion but exited non-zero.
// A fake Runner returns this so that exit-code-sensitive logic (such as the
// "nothing to commit" check) is reachable in tests without spawning a process.
type ExitError struct {
	// Code is the process exit status.
	Code int
	// Err is the underlying cause, if any.
	Err error
}

// Error implements the error interface.
func (e *ExitError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("exit status %d: %v", e.Code, e.Err)
	}
	return fmt.Sprintf("exit status %d", e.Code)
}

// Unwrap exposes the underlying cause to errors.Is / errors.As.
func (e *ExitError) Unwrap() error { return e.Err }

// ExitCode returns the process exit status.
func (e *ExitError) ExitCode() int { return e.Code }

// ExitCodeOf reports the exit status carried by err and whether err represents
// a completed-but-failed command. It recognizes both the real *exec.ExitError
// produced by ExecRunner and the *ExitError produced by fakes, so callers can
// branch on an exit code without caring which Runner they were given.
func ExitCodeOf(err error) (int, bool) {
	if err == nil {
		return 0, false
	}

	var gitErr *ExitError
	if errors.As(err, &gitErr) {
		return gitErr.Code, true
	}

	var execErr *exec.ExitError
	if errors.As(err, &execErr) {
		return execErr.ExitCode(), true
	}

	return 0, false
}
