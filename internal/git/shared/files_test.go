package shared

import (
	"strings"
	"testing"
)

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
