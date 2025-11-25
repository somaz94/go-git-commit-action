package executor

import (
	"fmt"
	"io"
	"strings"
)

// MockExecutor is a test implementation of CommandExecutor
// that records all commands executed and can return predefined outputs.
type MockExecutor struct {
	// Commands records all executed commands
	Commands []ExecutedCommand

	// Outputs maps command patterns to their outputs
	Outputs map[string][]byte

	// Errors maps command patterns to their errors
	Errors map[string]error

	// StreamOutputs maps command patterns to their stream outputs
	StreamOutputs map[string]string
}

// ExecutedCommand represents a command that was executed.
type ExecutedCommand struct {
	Name string
	Args []string
}

// NewMockExecutor creates a new MockExecutor instance.
func NewMockExecutor() *MockExecutor {
	return &MockExecutor{
		Commands:      make([]ExecutedCommand, 0),
		Outputs:       make(map[string][]byte),
		Errors:        make(map[string]error),
		StreamOutputs: make(map[string]string),
	}
}

// Execute records the command and returns a predefined error if set.
func (m *MockExecutor) Execute(name string, args ...string) error {
	m.Commands = append(m.Commands, ExecutedCommand{Name: name, Args: args})

	key := m.buildKey(name, args...)
	if err, ok := m.Errors[key]; ok {
		return err
	}

	return nil
}

// ExecuteWithOutput records the command and returns predefined output if set.
func (m *MockExecutor) ExecuteWithOutput(name string, args ...string) ([]byte, error) {
	m.Commands = append(m.Commands, ExecutedCommand{Name: name, Args: args})

	key := m.buildKey(name, args...)

	// Check for error first
	if err, ok := m.Errors[key]; ok {
		return nil, err
	}

	// Return predefined output if available
	if output, ok := m.Outputs[key]; ok {
		return output, nil
	}

	return []byte{}, nil
}

// ExecuteWithStreams records the command and writes predefined output to streams.
func (m *MockExecutor) ExecuteWithStreams(name string, args []string, stdout, stderr io.Writer) error {
	m.Commands = append(m.Commands, ExecutedCommand{Name: name, Args: args})

	key := m.buildKey(name, args...)

	// Check for error first
	if err, ok := m.Errors[key]; ok {
		if stderr != nil {
			fmt.Fprintf(stderr, "mock error: %v", err)
		}
		return err
	}

	// Write predefined output if available
	if output, ok := m.StreamOutputs[key]; ok {
		if stdout != nil {
			fmt.Fprint(stdout, output)
		}
	}

	return nil
}

// SetOutput sets a predefined output for a specific command.
func (m *MockExecutor) SetOutput(output []byte, name string, args ...string) {
	key := m.buildKey(name, args...)
	m.Outputs[key] = output
}

// SetError sets a predefined error for a specific command.
func (m *MockExecutor) SetError(err error, name string, args ...string) {
	key := m.buildKey(name, args...)
	m.Errors[key] = err
}

// SetStreamOutput sets a predefined stream output for a specific command.
func (m *MockExecutor) SetStreamOutput(output string, name string, args ...string) {
	key := m.buildKey(name, args...)
	m.StreamOutputs[key] = output
}

// GetExecutedCommands returns all executed commands.
func (m *MockExecutor) GetExecutedCommands() []ExecutedCommand {
	return m.Commands
}

// GetLastCommand returns the last executed command, or nil if no commands were executed.
func (m *MockExecutor) GetLastCommand() *ExecutedCommand {
	if len(m.Commands) == 0 {
		return nil
	}
	return &m.Commands[len(m.Commands)-1]
}

// Reset clears all recorded commands and outputs.
func (m *MockExecutor) Reset() {
	m.Commands = make([]ExecutedCommand, 0)
	m.Outputs = make(map[string][]byte)
	m.Errors = make(map[string]error)
	m.StreamOutputs = make(map[string]string)
}

// CommandExecuted checks if a specific command was executed.
func (m *MockExecutor) CommandExecuted(name string, args ...string) bool {
	for _, cmd := range m.Commands {
		if cmd.Name == name && m.argsMatch(cmd.Args, args) {
			return true
		}
	}
	return false
}

// buildKey creates a unique key for a command and its arguments.
func (m *MockExecutor) buildKey(name string, args ...string) string {
	return name + " " + strings.Join(args, " ")
}

// argsMatch checks if two argument slices match.
func (m *MockExecutor) argsMatch(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
