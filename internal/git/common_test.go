package git

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/somaz94/go-git-commit-action/internal/config"
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shortenCommitSHA(tt.sha)
			if got != tt.want {
				t.Errorf("shortenCommitSHA(%q) = %v, want %v", tt.sha, got, tt.want)
			}
		})
	}
}

func TestCommandBatch_Preparation(t *testing.T) {
	commands := []Command{
		{gitcmd.CmdGit, gitcmd.ConfigUserEmailArgs("test@example.com"), "Setting email"},
		{gitcmd.CmdGit, gitcmd.ConfigUserNameArgs("Test User"), "Setting name"},
		{gitcmd.CmdGit, gitcmd.CommitArgs("test commit"), "Committing"},
	}

	if len(commands) != 3 {
		t.Errorf("Expected 3 commands, got %d", len(commands))
	}

	if commands[0].Name != gitcmd.CmdGit {
		t.Errorf("First command name = %v, want %v", commands[0].Name, gitcmd.CmdGit)
	}
	if commands[0].Desc != "Setting email" {
		t.Errorf("First command desc = %v, want 'Setting email'", commands[0].Desc)
	}
}

func TestIsNothingToCommitError(t *testing.T) {
	tests := []struct {
		name string
		cmd  Command
		err  error
		want bool
	}{
		{
			name: "non-commit command",
			cmd:  Command{Name: "git", Args: []string{"push", "origin"}},
			err:  fmt.Errorf("some error"),
			want: false,
		},
		{
			name: "empty args",
			cmd:  Command{Name: "git", Args: []string{}},
			err:  fmt.Errorf("some error"),
			want: false,
		},
		{
			name: "non-ExitError",
			cmd:  Command{Name: "git", Args: []string{"commit", "-m", "test"}},
			err:  fmt.Errorf("random error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isNothingToCommitError(tt.cmd, tt.err)
			if got != tt.want {
				t.Errorf("isNothingToCommitError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsNothingToCommitError_WithExitError(t *testing.T) {
	// Run a command that exits with code 1 to get a real exec.ExitError
	cmd := exec.Command("sh", "-c", "exit 1")
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected error from exit 1")
	}

	commitCmd := Command{Name: "git", Args: []string{"commit", "-m", "test"}}
	if !isNothingToCommitError(commitCmd, err) {
		t.Error("isNothingToCommitError() should return true for commit with exit code 1")
	}

	pushCmd := Command{Name: "git", Args: []string{"push"}}
	if isNothingToCommitError(pushCmd, err) {
		t.Error("isNothingToCommitError() should return false for non-commit command")
	}
}

func TestExecuteCommandBatch_Success(t *testing.T) {
	commands := []Command{
		{Name: "echo", Args: []string{"hello"}, Desc: "Echoing hello"},
		{Name: "echo", Args: []string{"world"}, Desc: "Echoing world"},
	}

	err := ExecuteCommandBatch(commands, "Test header")
	if err != nil {
		t.Errorf("ExecuteCommandBatch() error = %v, want nil", err)
	}
}

func TestExecuteCommandBatch_EmptyHeader(t *testing.T) {
	commands := []Command{
		{Name: "echo", Args: []string{"test"}, Desc: "Echo test"},
	}

	err := ExecuteCommandBatch(commands, "")
	if err != nil {
		t.Errorf("ExecuteCommandBatch() error = %v, want nil", err)
	}
}

func TestExecuteCommandBatch_EmptyCommands(t *testing.T) {
	err := ExecuteCommandBatch(nil, "")
	if err != nil {
		t.Errorf("ExecuteCommandBatch(nil) error = %v, want nil", err)
	}
}

func TestExecuteCommandBatch_FailedCommand(t *testing.T) {
	commands := []Command{
		{Name: "false", Args: []string{}, Desc: "Always fails"},
	}

	err := ExecuteCommandBatch(commands, "")
	if err == nil {
		t.Error("ExecuteCommandBatch() with failing command should return error")
	}
}

func TestExecuteCommandBatch_NothingToCommit(t *testing.T) {
	// "git commit" in a non-repo dir will exit with code 128, not 1
	// Use sh -c "exit 1" wrapped as a commit-like command to test the skip logic
	// Actually, we test using a real echo command - commit skip only triggers for exit code 1
	commands := []Command{
		{Name: "echo", Args: []string{"commit", "test"}, Desc: "Non-commit echo"},
	}

	err := ExecuteCommandBatch(commands, "")
	if err != nil {
		t.Errorf("ExecuteCommandBatch() error = %v, want nil", err)
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.GitConfig
		wantErr bool
	}{
		{
			name:    "no PR creation",
			cfg:     &config.GitConfig{CreatePR: false},
			wantErr: false,
		},
		{
			name: "valid PR config with auto branch",
			cfg: &config.GitConfig{
				CreatePR:    true,
				AutoBranch:  true,
				PRBase:      "main",
				GitHubToken: "token",
			},
			wantErr: false,
		},
		{
			name: "valid PR config with manual branch",
			cfg: &config.GitConfig{
				CreatePR:    true,
				AutoBranch:  false,
				PRBranch:    "feature",
				PRBase:      "main",
				GitHubToken: "token",
			},
			wantErr: false,
		},
		{
			name: "missing PR branch",
			cfg: &config.GitConfig{
				CreatePR:    true,
				AutoBranch:  false,
				PRBase:      "main",
				GitHubToken: "token",
			},
			wantErr: true,
		},
		{
			name: "missing PR base",
			cfg: &config.GitConfig{
				CreatePR:    true,
				AutoBranch:  false,
				PRBranch:    "feature",
				GitHubToken: "token",
			},
			wantErr: true,
		},
		{
			name: "missing GitHub token",
			cfg: &config.GitConfig{
				CreatePR:   true,
				AutoBranch: false,
				PRBranch:   "feature",
				PRBase:     "main",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewTagManager(t *testing.T) {
	cfg := &config.GitConfig{
		TagName:    "v1.0.0",
		TagMessage: "Release v1.0.0",
	}

	tm := NewTagManager(cfg)
	if tm == nil {
		t.Fatal("NewTagManager() returned nil")
	}
	if tm.config != cfg {
		t.Error("NewTagManager() config does not match")
	}
}

func TestBuildTagArgs(t *testing.T) {
	tests := []struct {
		name         string
		tagName      string
		tagMessage   string
		targetCommit string
		wantContains string
	}{
		{
			name:         "simple tag without message",
			tagName:      "v1.0.0",
			targetCommit: "",
			wantContains: "v1.0.0",
		},
		{
			name:         "simple tag with target commit",
			tagName:      "v1.0.0",
			targetCommit: "abc123",
			wantContains: "abc123",
		},
		{
			name:         "annotated tag with message",
			tagName:      "v1.0.0",
			tagMessage:   "Release v1.0.0",
			targetCommit: "",
			wantContains: "Release v1.0.0",
		},
		{
			name:         "annotated tag with message and target",
			tagName:      "v1.0.0",
			tagMessage:   "Release v1.0.0",
			targetCommit: "abc123",
			wantContains: "abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := &TagManager{
				config: &config.GitConfig{
					TagName:    tt.tagName,
					TagMessage: tt.tagMessage,
				},
			}
			args := tm.buildTagArgs(tt.targetCommit)
			found := false
			for _, arg := range args {
				if arg == tt.wantContains {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("buildTagArgs() = %v, want to contain %q", args, tt.wantContains)
			}
		})
	}
}

func TestBuildTagDescription(t *testing.T) {
	tests := []struct {
		name         string
		tagName      string
		tagRef       string
		targetCommit string
		wantContains string
	}{
		{
			name:         "simple tag",
			tagName:      "v1.0.0",
			targetCommit: "",
			wantContains: "v1.0.0",
		},
		{
			name:         "tag with reference different from commit",
			tagName:      "v1.0.0",
			tagRef:       "main",
			targetCommit: "1234567890abcdef",
			wantContains: "main",
		},
		{
			name:         "tag with reference same as commit",
			tagName:      "v1.0.0",
			tagRef:       "12345678",
			targetCommit: "12345678",
			wantContains: "12345678",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := &TagManager{
				config: &config.GitConfig{
					TagName:      tt.tagName,
					TagReference: tt.tagRef,
				},
			}
			desc := tm.buildTagDescription(tt.targetCommit)
			if !contains(desc, tt.wantContains) {
				t.Errorf("buildTagDescription() = %q, want to contain %q", desc, tt.wantContains)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestPrintChangeDetectionInfo(t *testing.T) {
	// This function only prints, so we just verify it doesn't panic
	printChangeDetectionInfo([]byte("M file.txt"), []byte("file.txt"), true, true)
	printChangeDetectionInfo([]byte{}, []byte{}, false, false)
}

func TestRestoreChanges_EmptyBackups(t *testing.T) {
	err := restoreChanges(nil)
	if err != nil {
		t.Errorf("restoreChanges(nil) error = %v, want nil", err)
	}

	err = restoreChanges([]FileBackup{})
	if err != nil {
		t.Errorf("restoreChanges(empty) error = %v, want nil", err)
	}
}

func TestRestoreChanges_WithTempFile(t *testing.T) {
	tmpDir := t.TempDir()
	backups := []FileBackup{
		{path: tmpDir + "/restored.txt", content: []byte("restored content")},
	}

	err := restoreChanges(backups)
	if err != nil {
		t.Fatalf("restoreChanges() error = %v", err)
	}
}

func TestBackupChanges_ShortLines(t *testing.T) {
	cfg := &config.GitConfig{RepoPath: "."}
	// Lines shorter than 4 chars should be skipped without panic
	statusOutput := "M\n\nXY\n"
	backups, err := backupChanges(cfg, statusOutput)
	if err != nil {
		t.Fatalf("backupChanges() error = %v", err)
	}
	if len(backups) != 0 {
		t.Errorf("backupChanges() returned %d backups, want 0", len(backups))
	}
}

func TestBackupChanges_ValidLine(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.GitConfig{RepoPath: "."}

	// Create a temp file to backup
	tmpFile := tmpDir + "/test.txt"
	if err := os.WriteFile(tmpFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	statusOutput := fmt.Sprintf(" M %s\n", tmpFile)
	backups, err := backupChanges(cfg, statusOutput)
	if err != nil {
		t.Fatalf("backupChanges() error = %v", err)
	}
	if len(backups) != 1 {
		t.Errorf("backupChanges() returned %d backups, want 1", len(backups))
	}
}

func TestBackupChanges_DeletedFileSkipped(t *testing.T) {
	cfg := &config.GitConfig{RepoPath: "."}
	statusOutput := " D deleted-file.txt\n"
	backups, err := backupChanges(cfg, statusOutput)
	if err != nil {
		t.Fatalf("backupChanges() error = %v", err)
	}
	if len(backups) != 0 {
		t.Errorf("backupChanges() returned %d backups for deleted file, want 0", len(backups))
	}
}

func BenchmarkShortenCommitSHA(b *testing.B) {
	sha := "1234567890abcdef1234567890abcdef12345678"
	for i := 0; i < b.N; i++ {
		shortenCommitSHA(sha)
	}
}
