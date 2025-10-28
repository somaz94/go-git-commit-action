package gitcmd

// Git command constants
const (
	// Base command
	CmdGit = "git"
)

// Git subcommands
const (
	SubCmdConfig   = "config"
	SubCmdCommit   = "commit"
	SubCmdPush     = "push"
	SubCmdFetch    = "fetch"
	SubCmdCheckout = "checkout"
	SubCmdTag      = "tag"
	SubCmdStatus   = "status"
	SubCmdAdd      = "add"
	SubCmdStash    = "stash"
	SubCmdReset    = "reset"
	SubCmdRevParse = "rev-parse"
	SubCmdLsRemote = "ls-remote"
	SubCmdDiff     = "diff"
	SubCmdRevList  = "rev-list"
)

// Git global options
const (
	OptGlobal     = "--global"
	OptAdd        = "--add"
	OptList       = "--list"
	OptForce      = "-f"
	OptHard       = "--hard"
	OptUpstream   = "-u"
	OptPorcelain  = "--porcelain"
	OptVerify     = "--verify"
	OptHeads      = "--heads"
	OptTags       = "--tags"
	OptNameOnly   = "--name-only"
	OptNameStatus = "--name-status"
)

// Git config specific options
const (
	ConfigSafeDirectory = "safe.directory"
	ConfigUserEmail     = "user.email"
	ConfigUserName      = "user.name"
)

// Git commit options
const (
	OptMessage  = "-m"
	OptAnnotate = "-a"
	OptDelete   = "-d"
)

// Git stash options
const (
	StashPush         = "push"
	StashOptUntracked = "-u"
)

// Common paths
const (
	PathApp             = "/app"
	PathGitHubWorkspace = "/github/workspace"
)

// Git references
const (
	RefOrigin = "origin"
	RefTags   = "refs/tags/"
)

// BuildArgs is a helper function to construct git command arguments.
// It provides a fluent interface for building command arguments.
type ArgsBuilder struct {
	args []string
}

// NewArgsBuilder creates a new arguments builder.
func NewArgsBuilder() *ArgsBuilder {
	return &ArgsBuilder{
		args: make([]string, 0),
	}
}

// Add adds one or more arguments to the builder.
func (b *ArgsBuilder) Add(args ...string) *ArgsBuilder {
	b.args = append(b.args, args...)
	return b
}

// Build returns the constructed arguments slice.
func (b *ArgsBuilder) Build() []string {
	return b.args
}

// Common git command builders for convenience

// ConfigSafeDirArgs builds arguments for setting safe directory.
func ConfigSafeDirArgs(path string) []string {
	return NewArgsBuilder().
		Add(SubCmdConfig, OptGlobal, OptAdd, ConfigSafeDirectory, path).
		Build()
}

// ConfigUserEmailArgs builds arguments for setting user email.
func ConfigUserEmailArgs(email string) []string {
	return NewArgsBuilder().
		Add(SubCmdConfig, OptGlobal, ConfigUserEmail, email).
		Build()
}

// ConfigUserNameArgs builds arguments for setting user name.
func ConfigUserNameArgs(name string) []string {
	return NewArgsBuilder().
		Add(SubCmdConfig, OptGlobal, ConfigUserName, name).
		Build()
}

// ConfigListArgs builds arguments for listing git configuration.
func ConfigListArgs() []string {
	return NewArgsBuilder().
		Add(SubCmdConfig, OptGlobal, OptList).
		Build()
}

// CommitArgs builds arguments for committing changes.
func CommitArgs(message string) []string {
	return NewArgsBuilder().
		Add(SubCmdCommit, OptMessage, message).
		Build()
}

// PushArgs builds arguments for pushing to remote.
func PushArgs(remote, branch string) []string {
	return NewArgsBuilder().
		Add(SubCmdPush, remote, branch).
		Build()
}

// PushUpstreamArgs builds arguments for pushing with upstream.
func PushUpstreamArgs(remote, branch string) []string {
	return NewArgsBuilder().
		Add(SubCmdPush, OptUpstream, remote, branch).
		Build()
}

// FetchArgs builds arguments for fetching from remote.
func FetchArgs(remote, branch string) []string {
	return NewArgsBuilder().
		Add(SubCmdFetch, remote, branch).
		Build()
}

// CheckoutArgs builds arguments for checking out a branch.
func CheckoutArgs(branch string) []string {
	return NewArgsBuilder().
		Add(SubCmdCheckout, branch).
		Build()
}

// CheckoutNewBranchArgs builds arguments for creating and checking out a new branch.
func CheckoutNewBranchArgs(branch string) []string {
	return NewArgsBuilder().
		Add(SubCmdCheckout, "-b", branch).
		Build()
}

// StatusPorcelainArgs builds arguments for getting status in porcelain format.
func StatusPorcelainArgs() []string {
	return NewArgsBuilder().
		Add(SubCmdStatus, OptPorcelain).
		Build()
}

// AddArgs builds arguments for adding files.
func AddArgs(pattern string) []string {
	return NewArgsBuilder().
		Add(SubCmdAdd, pattern).
		Build()
}

// TagCreateArgs builds arguments for creating a tag.
func TagCreateArgs(tagName string, force bool) []string {
	builder := NewArgsBuilder().Add(SubCmdTag)
	if force {
		builder.Add(OptForce)
	}
	return builder.Add(tagName).Build()
}

// TagCreateAnnotatedArgs builds arguments for creating an annotated tag.
func TagCreateAnnotatedArgs(tagName, message string, force bool) []string {
	builder := NewArgsBuilder().Add(SubCmdTag)
	if force {
		builder.Add(OptForce)
	}
	return builder.Add(OptAnnotate, tagName, OptMessage, message).Build()
}

// TagDeleteArgs builds arguments for deleting a tag.
func TagDeleteArgs(tagName string) []string {
	return NewArgsBuilder().
		Add(SubCmdTag, OptDelete, tagName).
		Build()
}

// PushTagArgs builds arguments for pushing a tag.
func PushTagArgs(tagName string, force bool) []string {
	builder := NewArgsBuilder().Add(SubCmdPush)
	if force {
		builder.Add(OptForce)
	}
	return builder.Add(RefOrigin, tagName).Build()
}

// DeleteRemoteTagArgs builds arguments for deleting a remote tag.
func DeleteRemoteTagArgs(tagName string) []string {
	return NewArgsBuilder().
		Add(SubCmdPush, RefOrigin, ":"+RefTags+tagName).
		Build()
}

// FetchTagsArgs builds arguments for fetching tags.
func FetchTagsArgs() []string {
	return NewArgsBuilder().
		Add(SubCmdFetch, OptTags, OptForce, RefOrigin).
		Build()
}

// RevParseArgs builds arguments for rev-parse command.
func RevParseArgs(ref string) []string {
	return NewArgsBuilder().
		Add(SubCmdRevParse, OptVerify, ref).
		Build()
}

// LsRemoteHeadsArgs builds arguments for listing remote heads.
func LsRemoteHeadsArgs(remote, branch string) []string {
	return NewArgsBuilder().
		Add(SubCmdLsRemote, OptHeads, remote, branch).
		Build()
}

// ResetHardArgs builds arguments for hard reset.
func ResetHardArgs(ref string) []string {
	return NewArgsBuilder().
		Add(SubCmdReset, OptHard, ref).
		Build()
}

// StashPushArgs builds arguments for stash push.
func StashPushArgs() []string {
	return NewArgsBuilder().
		Add(SubCmdStash, StashPush, StashOptUntracked).
		Build()
}

// DiffNameOnlyArgs builds arguments for diff with name only.
func DiffNameOnlyArgs(base, head string) []string {
	return NewArgsBuilder().
		Add(SubCmdDiff, base+"..."+head, OptNameOnly).
		Build()
}

// DiffNameStatusArgs builds arguments for diff with name status.
func DiffNameStatusArgs(base, head string) []string {
	return NewArgsBuilder().
		Add(SubCmdDiff, base+".."+head, OptNameStatus).
		Build()
}

// RevListArgs builds arguments for rev-list command.
func RevListArgs(ref string) []string {
	return NewArgsBuilder().
		Add(SubCmdRevList, "-n1", ref).
		Build()
}
