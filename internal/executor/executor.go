package executor

import (
	"io"
	"os/exec"
)

// CommandExecutor defines the interface for executing system commands.
// This interface allows for dependency injection and makes the code testable
// by enabling the use of mock implementations in tests.
type CommandExecutor interface {
	// Execute runs a command and returns an error if it fails.
	Execute(name string, args ...string) error

	// ExecuteWithOutput runs a command and returns its combined stdout/stderr output.
	ExecuteWithOutput(name string, args ...string) ([]byte, error)

	// ExecuteWithStreams runs a command with custom stdout/stderr streams.
	ExecuteWithStreams(name string, args []string, stdout, stderr io.Writer) error
}

// RealExecutor is the production implementation of CommandExecutor
// that executes actual system commands.
type RealExecutor struct{}

// NewRealExecutor creates a new RealExecutor instance.
func NewRealExecutor() *RealExecutor {
	return &RealExecutor{}
}

// Execute runs a command and returns an error if it fails.
func (e *RealExecutor) Execute(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}

// ExecuteWithOutput runs a command and returns its combined stdout/stderr output.
func (e *RealExecutor) ExecuteWithOutput(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.CombinedOutput()
}

// ExecuteWithStreams runs a command with custom stdout/stderr streams.
func (e *RealExecutor) ExecuteWithStreams(name string, args []string, stdout, stderr io.Writer) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}
