package executor

import (
	"bytes"
	"errors"
	"testing"
)

func TestRealExecutor_Execute(t *testing.T) {
	executor := NewRealExecutor()

	// Test successful command
	err := executor.Execute("echo", "hello")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test failing command
	err = executor.Execute("false")
	if err == nil {
		t.Error("Expected error for 'false' command")
	}
}

func TestRealExecutor_ExecuteWithOutput(t *testing.T) {
	executor := NewRealExecutor()

	// Test command with output
	output, err := executor.ExecuteWithOutput("echo", "test output")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(output) == 0 {
		t.Error("Expected output, got none")
	}
}

func TestRealExecutor_ExecuteWithStreams(t *testing.T) {
	executor := NewRealExecutor()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// Test successful command with streams
	err := executor.ExecuteWithStreams("echo", []string{"hello"}, &stdout, &stderr)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if stdout.Len() == 0 {
		t.Error("Expected stdout output, got none")
	}

	// Test failing command with streams
	stdout.Reset()
	stderr.Reset()
	err = executor.ExecuteWithStreams("false", []string{}, &stdout, &stderr)
	if err == nil {
		t.Error("Expected error for 'false' command")
	}
}

func TestMockExecutor_Execute(t *testing.T) {
	mock := NewMockExecutor()

	// Test successful execution
	err := mock.Execute("git", "status")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify command was recorded
	if len(mock.GetExecutedCommands()) != 1 {
		t.Errorf("Expected 1 command, got %d", len(mock.GetExecutedCommands()))
	}

	cmd := mock.GetLastCommand()
	if cmd.Name != "git" {
		t.Errorf("Expected command name 'git', got '%s'", cmd.Name)
	}
}

func TestMockExecutor_ExecuteWithError(t *testing.T) {
	mock := NewMockExecutor()
	expectedErr := errors.New("test error")

	// Set up error for specific command
	mock.SetError(expectedErr, "git", "push")

	// Execute command
	err := mock.Execute("git", "push")
	if err != expectedErr {
		t.Errorf("Expected error '%v', got '%v'", expectedErr, err)
	}
}

func TestMockExecutor_ExecuteWithOutput(t *testing.T) {
	mock := NewMockExecutor()
	expectedOutput := []byte("test output")

	// Set up output for specific command
	mock.SetOutput(expectedOutput, "git", "log")

	// Execute command
	output, err := mock.ExecuteWithOutput("git", "log")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if string(output) != string(expectedOutput) {
		t.Errorf("Expected output '%s', got '%s'", expectedOutput, output)
	}

	// Test with error
	mock.Reset()
	expectedErr := errors.New("output error")
	mock.SetError(expectedErr, "git", "status")

	_, err = mock.ExecuteWithOutput("git", "status")
	if err != expectedErr {
		t.Errorf("Expected error '%v', got '%v'", expectedErr, err)
	}

	// Test command without configured output
	mock.Reset()
	output, err = mock.ExecuteWithOutput("git", "branch")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(output) != 0 {
		t.Errorf("Expected empty output, got '%s'", output)
	}
}

func TestMockExecutor_ExecuteWithStreams(t *testing.T) {
	mock := NewMockExecutor()
	var stdout bytes.Buffer

	expectedOutput := "stream output"
	mock.SetStreamOutput(expectedOutput, "git", "diff")

	// Execute command with stream output
	err := mock.ExecuteWithStreams("git", []string{"diff"}, &stdout, nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if stdout.String() != expectedOutput {
		t.Errorf("Expected output '%s', got '%s'", expectedOutput, stdout.String())
	}

	// Test with error
	mock.Reset()
	stdout.Reset()
	expectedErr := errors.New("stream error")
	mock.SetError(expectedErr, "git", "push")

	err = mock.ExecuteWithStreams("git", []string{"push"}, &stdout, nil)
	if err != expectedErr {
		t.Errorf("Expected error '%v', got '%v'", expectedErr, err)
	}

	// Test command without configured stream output
	mock.Reset()
	stdout.Reset()
	err = mock.ExecuteWithStreams("git", []string{"status"}, &stdout, nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if stdout.Len() != 0 {
		t.Errorf("Expected empty stdout, got '%s'", stdout.String())
	}
}

func TestMockExecutor_CommandExecuted(t *testing.T) {
	mock := NewMockExecutor()

	// Execute some commands
	mock.Execute("git", "status")
	mock.Execute("git", "add", ".")
	mock.Execute("git", "commit", "-m", "test")

	// Test CommandExecuted
	if !mock.CommandExecuted("git", "status") {
		t.Error("Expected 'git status' to be executed")
	}

	if !mock.CommandExecuted("git", "add", ".") {
		t.Error("Expected 'git add .' to be executed")
	}

	if mock.CommandExecuted("git", "push") {
		t.Error("Expected 'git push' to NOT be executed")
	}
}

func TestMockExecutor_Reset(t *testing.T) {
	mock := NewMockExecutor()

	// Execute some commands and set outputs
	mock.Execute("git", "status")
	mock.SetOutput([]byte("output"), "git", "log")
	mock.SetError(errors.New("error"), "git", "push")

	// Reset
	mock.Reset()

	// Verify everything is cleared
	if len(mock.GetExecutedCommands()) != 0 {
		t.Error("Expected no commands after reset")
	}

	if len(mock.Outputs) != 0 {
		t.Error("Expected no outputs after reset")
	}

	if len(mock.Errors) != 0 {
		t.Error("Expected no errors after reset")
	}
}

func TestMockExecutor_GetLastCommand(t *testing.T) {
	mock := NewMockExecutor()

	// Test with no commands
	cmd := mock.GetLastCommand()
	if cmd != nil {
		t.Errorf("Expected nil for empty command list, got %v", cmd)
	}

	// Execute commands
	mock.Execute("git", "status")
	mock.Execute("git", "add", ".")

	// Test last command
	cmd = mock.GetLastCommand()
	if cmd == nil {
		t.Fatal("Expected command, got nil")
	}
	if cmd.Name != "git" {
		t.Errorf("Expected command name 'git', got '%s'", cmd.Name)
	}
	if len(cmd.Args) != 2 || cmd.Args[0] != "add" || cmd.Args[1] != "." {
		t.Errorf("Expected args ['add', '.'], got %v", cmd.Args)
	}
}

func TestMockExecutor_MultipleCommands(t *testing.T) {
	mock := NewMockExecutor()

	// Execute multiple commands
	commands := []struct {
		name string
		args []string
	}{
		{"git", []string{"init"}},
		{"git", []string{"add", "."}},
		{"git", []string{"commit", "-m", "test"}},
	}

	for _, cmd := range commands {
		err := mock.Execute(cmd.name, cmd.args...)
		if err != nil {
			t.Errorf("Unexpected error for command '%s %v': %v", cmd.name, cmd.args, err)
		}
	}

	// Verify all commands were recorded
	executed := mock.GetExecutedCommands()
	if len(executed) != len(commands) {
		t.Errorf("Expected %d commands, got %d", len(commands), len(executed))
	}

	// Verify each command
	for i, cmd := range commands {
		if !mock.CommandExecuted(cmd.name, cmd.args...) {
			t.Errorf("Command '%s %v' was not recorded", cmd.name, cmd.args)
		}

		if executed[i].Name != cmd.name {
			t.Errorf("Command %d: expected name '%s', got '%s'", i, cmd.name, executed[i].Name)
		}
	}
}
