package shared

import (
	"os/exec"
	"strings"
	"testing"
)

func TestIsNothingToCommitExit(t *testing.T) {
	// A real exit-code-1 failure (git's "nothing to commit" signal).
	err1 := exec.Command("sh", "-c", "exit 1").Run()
	if err1 == nil {
		t.Fatal("expected error from exit 1")
	}
	if !isNothingToCommitExit(err1) {
		t.Error("isNothingToCommitExit() = false for exit code 1, want true")
	}

	// A different non-zero exit code must not be treated as nothing-to-commit.
	err2 := exec.Command("sh", "-c", "exit 2").Run()
	if err2 == nil {
		t.Fatal("expected error from exit 2")
	}
	if isNothingToCommitExit(err2) {
		t.Error("isNothingToCommitExit() = true for exit code 2, want false")
	}

	// A nil error is not a nothing-to-commit exit.
	if isNothingToCommitExit(nil) {
		t.Error("isNothingToCommitExit(nil) = true, want false")
	}
}

func TestStageFiles_EmptyPattern(t *testing.T) {
	// strings.Fields("") returns empty slice, so no git add is called
	// This should succeed with no operations
	err := StageFiles("")
	if err != nil {
		t.Errorf("StageFiles('') error = %v, want nil", err)
	}
}

func TestStageFiles_WhitespaceOnly(t *testing.T) {
	err := StageFiles("   ")
	if err != nil {
		t.Errorf("StageFiles('   ') error = %v, want nil", err)
	}
}

func TestFieldsSplitting(t *testing.T) {
	// Verify the splitting logic matches expectations
	tests := []struct {
		input string
		want  int
	}{
		{"", 0},
		{".", 1},
		{"file1.txt file2.txt", 2},
		{"  file1.txt  file2.txt  file3.txt  ", 3},
		{"single", 1},
	}

	for _, tt := range tests {
		fields := strings.Fields(tt.input)
		if len(fields) != tt.want {
			t.Errorf("Fields(%q) = %d fields, want %d", tt.input, len(fields), tt.want)
		}
	}
}
