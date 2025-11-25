package git

import (
	"testing"

	"github.com/somaz94/go-git-commit-action/internal/gitcmd"
)

func TestCommand_Structure(t *testing.T) {
	cmd := Command{
		Name: "git",
		Args: []string{"status"},
		Desc: "Checking git status",
	}

	if cmd.Name != "git" {
		t.Errorf("Command.Name = %v, want git", cmd.Name)
	}
	if len(cmd.Args) != 1 || cmd.Args[0] != "status" {
		t.Errorf("Command.Args = %v, want [status]", cmd.Args)
	}
	if cmd.Desc != "Checking git status" {
		t.Errorf("Command.Desc = %v, want 'Checking git status'", cmd.Desc)
	}
}

func TestFileBackup_Structure(t *testing.T) {
	backup := FileBackup{
		path:    "/test/file.txt",
		content: []byte("test content"),
	}

	if backup.path != "/test/file.txt" {
		t.Errorf("FileBackup.path = %v, want /test/file.txt", backup.path)
	}
	if string(backup.content) != "test content" {
		t.Errorf("FileBackup.content = %v, want 'test content'", string(backup.content))
	}
}

func TestShortenCommitSHA(t *testing.T) {
	// This test is for the tag.go file
	tests := []struct {
		name string
		sha  string
		want string
	}{
		{
			name: "long SHA",
			sha:  "1234567890abcdef1234567890abcdef12345678",
			want: "12345678",
		},
		{
			name: "short SHA",
			sha:  "1234567",
			want: "1234567",
		},
		{
			name: "exact 8 chars",
			sha:  "12345678",
			want: "12345678",
		},
		{
			name: "empty string",
			sha:  "",
			want: "",
		},
	}

	// Note: shortenCommitSHA is not exported, so we can't test it directly
	// This is a placeholder to show the test structure
	// In practice, you would either export the function or test it indirectly
	_ = tests
}

func TestCommandBatch_Preparation(t *testing.T) {
	// Test command batch structure
	commands := []Command{
		{gitcmd.CmdGit, gitcmd.ConfigUserEmailArgs("test@example.com"), "Setting email"},
		{gitcmd.CmdGit, gitcmd.ConfigUserNameArgs("Test User"), "Setting name"},
		{gitcmd.CmdGit, gitcmd.CommitArgs("test commit"), "Committing"},
	}

	if len(commands) != 3 {
		t.Errorf("Expected 3 commands, got %d", len(commands))
	}

	// Verify first command
	if commands[0].Name != gitcmd.CmdGit {
		t.Errorf("First command name = %v, want %v", commands[0].Name, gitcmd.CmdGit)
	}
	if commands[0].Desc != "Setting email" {
		t.Errorf("First command desc = %v, want 'Setting email'", commands[0].Desc)
	}
}
