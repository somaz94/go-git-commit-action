package gitcmd

import (
	"reflect"
	"testing"
)

func TestConfigSafeDirArgs(t *testing.T) {
	args := ConfigSafeDirArgs("/test/path")
	expected := []string{SubCmdConfig, OptGlobal, OptAdd, ConfigSafeDirectory, "/test/path"}

	if !reflect.DeepEqual(args, expected) {
		t.Errorf("ConfigSafeDirArgs() = %v, want %v", args, expected)
	}
}

func TestConfigUserEmailArgs(t *testing.T) {
	args := ConfigUserEmailArgs("test@example.com")
	expected := []string{SubCmdConfig, OptGlobal, ConfigUserEmail, "test@example.com"}

	if !reflect.DeepEqual(args, expected) {
		t.Errorf("ConfigUserEmailArgs() = %v, want %v", args, expected)
	}
}

func TestConfigUserNameArgs(t *testing.T) {
	args := ConfigUserNameArgs("Test User")
	expected := []string{SubCmdConfig, OptGlobal, ConfigUserName, "Test User"}

	if !reflect.DeepEqual(args, expected) {
		t.Errorf("ConfigUserNameArgs() = %v, want %v", args, expected)
	}
}

func TestConfigListArgs(t *testing.T) {
	args := ConfigListArgs()
	expected := []string{SubCmdConfig, OptGlobal, OptList}

	if !reflect.DeepEqual(args, expected) {
		t.Errorf("ConfigListArgs() = %v, want %v", args, expected)
	}
}

func TestCommitArgs(t *testing.T) {
	args := CommitArgs("test commit message")
	expected := []string{SubCmdCommit, OptMessage, "test commit message"}

	if !reflect.DeepEqual(args, expected) {
		t.Errorf("CommitArgs() = %v, want %v", args, expected)
	}
}

func TestPushArgs(t *testing.T) {
	args := PushArgs("origin", "main")
	expected := []string{SubCmdPush, "origin", "main"}

	if !reflect.DeepEqual(args, expected) {
		t.Errorf("PushArgs() = %v, want %v", args, expected)
	}
}

func TestPushUpstreamArgs(t *testing.T) {
	args := PushUpstreamArgs("origin", "feature-branch")
	expected := []string{SubCmdPush, OptUpstream, "origin", "feature-branch"}

	if !reflect.DeepEqual(args, expected) {
		t.Errorf("PushUpstreamArgs() = %v, want %v", args, expected)
	}
}

func TestFetchArgs(t *testing.T) {
	args := FetchArgs("origin", "main")
	expected := []string{SubCmdFetch, "origin", "main"}

	if !reflect.DeepEqual(args, expected) {
		t.Errorf("FetchArgs() = %v, want %v", args, expected)
	}
}

func TestCheckoutArgs(t *testing.T) {
	args := CheckoutArgs("develop")
	expected := []string{SubCmdCheckout, "develop"}

	if !reflect.DeepEqual(args, expected) {
		t.Errorf("CheckoutArgs() = %v, want %v", args, expected)
	}
}

func TestCheckoutNewBranchArgs(t *testing.T) {
	args := CheckoutNewBranchArgs("new-feature")
	expected := []string{SubCmdCheckout, "-b", "new-feature"}

	if !reflect.DeepEqual(args, expected) {
		t.Errorf("CheckoutNewBranchArgs() = %v, want %v", args, expected)
	}
}

func TestStatusPorcelainArgs(t *testing.T) {
	args := StatusPorcelainArgs()
	expected := []string{SubCmdStatus, OptPorcelain}

	if !reflect.DeepEqual(args, expected) {
		t.Errorf("StatusPorcelainArgs() = %v, want %v", args, expected)
	}
}

func TestAddArgs(t *testing.T) {
	args := AddArgs("*.go")
	expected := []string{SubCmdAdd, "*.go"}

	if !reflect.DeepEqual(args, expected) {
		t.Errorf("AddArgs() = %v, want %v", args, expected)
	}
}

func TestTagCreateArgs(t *testing.T) {
	tests := []struct {
		name    string
		tagName string
		force   bool
		want    []string
	}{
		{
			name:    "without force",
			tagName: "v1.0.0",
			force:   false,
			want:    []string{SubCmdTag, "v1.0.0"},
		},
		{
			name:    "with force",
			tagName: "v1.0.0",
			force:   true,
			want:    []string{SubCmdTag, OptForce, "v1.0.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := TagCreateArgs(tt.tagName, tt.force)
			if !reflect.DeepEqual(args, tt.want) {
				t.Errorf("TagCreateArgs() = %v, want %v", args, tt.want)
			}
		})
	}
}

func TestTagCreateAnnotatedArgs(t *testing.T) {
	tests := []struct {
		name    string
		tagName string
		message string
		force   bool
		want    []string
	}{
		{
			name:    "without force",
			tagName: "v1.0.0",
			message: "Release 1.0.0",
			force:   false,
			want:    []string{SubCmdTag, OptAnnotate, "v1.0.0", OptMessage, "Release 1.0.0"},
		},
		{
			name:    "with force",
			tagName: "v1.0.0",
			message: "Release 1.0.0",
			force:   true,
			want:    []string{SubCmdTag, OptForce, OptAnnotate, "v1.0.0", OptMessage, "Release 1.0.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := TagCreateAnnotatedArgs(tt.tagName, tt.message, tt.force)
			if !reflect.DeepEqual(args, tt.want) {
				t.Errorf("TagCreateAnnotatedArgs() = %v, want %v", args, tt.want)
			}
		})
	}
}

func TestTagDeleteArgs(t *testing.T) {
	args := TagDeleteArgs("v1.0.0")
	expected := []string{SubCmdTag, OptDelete, "v1.0.0"}

	if !reflect.DeepEqual(args, expected) {
		t.Errorf("TagDeleteArgs() = %v, want %v", args, expected)
	}
}

func TestPushTagArgs(t *testing.T) {
	tests := []struct {
		name    string
		tagName string
		force   bool
		want    []string
	}{
		{
			name:    "without force",
			tagName: "v1.0.0",
			force:   false,
			want:    []string{SubCmdPush, RefOrigin, "v1.0.0"},
		},
		{
			name:    "with force",
			tagName: "v1.0.0",
			force:   true,
			want:    []string{SubCmdPush, OptForce, RefOrigin, "v1.0.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := PushTagArgs(tt.tagName, tt.force)
			if !reflect.DeepEqual(args, tt.want) {
				t.Errorf("PushTagArgs() = %v, want %v", args, tt.want)
			}
		})
	}
}

func TestDeleteRemoteTagArgs(t *testing.T) {
	args := DeleteRemoteTagArgs("v1.0.0")
	expected := []string{SubCmdPush, RefOrigin, ":refs/tags/v1.0.0"}

	if !reflect.DeepEqual(args, expected) {
		t.Errorf("DeleteRemoteTagArgs() = %v, want %v", args, expected)
	}
}

func TestFetchTagsArgs(t *testing.T) {
	args := FetchTagsArgs()
	expected := []string{SubCmdFetch, OptTags, OptForce, RefOrigin}

	if !reflect.DeepEqual(args, expected) {
		t.Errorf("FetchTagsArgs() = %v, want %v", args, expected)
	}
}

func TestRevParseArgs(t *testing.T) {
	args := RevParseArgs("HEAD")
	expected := []string{SubCmdRevParse, OptVerify, "HEAD"}

	if !reflect.DeepEqual(args, expected) {
		t.Errorf("RevParseArgs() = %v, want %v", args, expected)
	}
}

func TestLsRemoteHeadsArgs(t *testing.T) {
	args := LsRemoteHeadsArgs("origin", "main")
	expected := []string{SubCmdLsRemote, OptHeads, "origin", "main"}

	if !reflect.DeepEqual(args, expected) {
		t.Errorf("LsRemoteHeadsArgs() = %v, want %v", args, expected)
	}
}

func TestResetHardArgs(t *testing.T) {
	args := ResetHardArgs("origin/main")
	expected := []string{SubCmdReset, OptHard, "origin/main"}

	if !reflect.DeepEqual(args, expected) {
		t.Errorf("ResetHardArgs() = %v, want %v", args, expected)
	}
}

func TestStashPushArgs(t *testing.T) {
	args := StashPushArgs()
	expected := []string{SubCmdStash, StashPush, StashOptUntracked}

	if !reflect.DeepEqual(args, expected) {
		t.Errorf("StashPushArgs() = %v, want %v", args, expected)
	}
}

func TestDiffNameOnlyArgs(t *testing.T) {
	args := DiffNameOnlyArgs("main", "develop")
	expected := []string{SubCmdDiff, "main...develop", OptNameOnly}

	if !reflect.DeepEqual(args, expected) {
		t.Errorf("DiffNameOnlyArgs() = %v, want %v", args, expected)
	}
}

func TestDiffNameStatusArgs(t *testing.T) {
	args := DiffNameStatusArgs("main", "develop")
	expected := []string{SubCmdDiff, "main..develop", OptNameStatus}

	if !reflect.DeepEqual(args, expected) {
		t.Errorf("DiffNameStatusArgs() = %v, want %v", args, expected)
	}
}

func TestRevListArgs(t *testing.T) {
	args := RevListArgs("v1.0.0")
	expected := []string{SubCmdRevList, "-n1", "v1.0.0"}

	if !reflect.DeepEqual(args, expected) {
		t.Errorf("RevListArgs() = %v, want %v", args, expected)
	}
}

func TestArgsBuilder(t *testing.T) {
	// Test the builder pattern
	builder := NewArgsBuilder()
	args := builder.
		Add(SubCmdConfig).
		Add(OptGlobal).
		Add("user.name").
		Add("Test User").
		Build()

	expected := []string{SubCmdConfig, OptGlobal, "user.name", "Test User"}

	if !reflect.DeepEqual(args, expected) {
		t.Errorf("ArgsBuilder.Build() = %v, want %v", args, expected)
	}
}

func TestArgsBuilderMultipleAdds(t *testing.T) {
	// Test adding multiple arguments at once
	builder := NewArgsBuilder()
	args := builder.
		Add(SubCmdPush, OptForce, RefOrigin, "main").
		Build()

	expected := []string{SubCmdPush, OptForce, RefOrigin, "main"}

	if !reflect.DeepEqual(args, expected) {
		t.Errorf("ArgsBuilder with multiple Add() = %v, want %v", args, expected)
	}
}
