package gitcmd

// Git command constants
const (
	// Base command
	CmdGit = "git"
)

// Git subcommands
const (
	SubCmdConfig      = "config"
	SubCmdCommit      = "commit"
	SubCmdPush        = "push"
	SubCmdFetch       = "fetch"
	SubCmdCheckout    = "checkout"
	SubCmdTag         = "tag"
	SubCmdStatus      = "status"
	SubCmdAdd         = "add"
	SubCmdStash       = "stash"
	SubCmdReset       = "reset"
	SubCmdRevParse    = "rev-parse"
	SubCmdLsRemote    = "ls-remote"
	SubCmdDiff        = "diff"
	SubCmdRevList     = "rev-list"
	SubCmdRemote      = "remote"
	SubCmdRebase      = "rebase"
	SubCmdMergeBase   = "merge-base"
	SubCmdSymbolicRef = "symbolic-ref"
)

// Git global options
const (
	OptShowScope        = "--show-scope"
	OptGetRegexp        = "--get-regexp"
	OptGet              = "--get"
	OptForce            = "-f"
	OptQuiet            = "-q"
	OptAbort            = "--abort"
	OptIsAncestor       = "--is-ancestor"
	OptHard             = "--hard"
	OptUpstream         = "-u"
	OptPorcelain        = "--porcelain"
	OptVerify           = "--verify"
	OptHeads            = "--heads"
	OptTags             = "--tags"
	OptNameOnly         = "--name-only"
	OptCached           = "--cached"
	OptSymbolicFullName = "--symbolic-full-name"
	OptNameStatus       = "--name-status"
	OptDeleteRemote     = "--delete"
	OptSetURL           = "set-url"
)

// Git config specific options
const (
	ConfigSafeDirectory = "safe.directory"
	ConfigUserEmail     = "user.email"
	ConfigUserName      = "user.name"
)

// Git commit and tag options
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
	RefOrigin    = "origin"
	RefHead      = "HEAD"
	RefFetchHead = "FETCH_HEAD"
	RefTags      = "refs/tags/"
	RefHeads     = "refs/heads/"
)

// ArgsBuilder constructs git command arguments through a fluent interface.
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

// ConfigListArgs builds arguments for listing the config keys the action sets,
// with the scope each value resolves from.
func ConfigListArgs() []string {
	return NewArgsBuilder().
		Add(SubCmdConfig, OptShowScope, OptGetRegexp, `^(safe\.directory|user\.email|user\.name)$`).
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

// StatusPorcelainArgs builds arguments for getting status in porcelain format,
// limited to pathspecs when any are given.
func StatusPorcelainArgs(pathspecs ...string) []string {
	builder := NewArgsBuilder().Add(SubCmdStatus, OptPorcelain)
	if len(pathspecs) > 0 {
		builder.Add("--").Add(pathspecs...)
	}
	return builder.Build()
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

// RebaseArgs builds arguments for rebasing onto ref. Deliberately no --autostash: when
// the stash cannot be reapplied git still exits 0 and leaves conflict markers in the tree.
func RebaseArgs(ref string) []string {
	return NewArgsBuilder().
		Add(SubCmdRebase, ref).
		Build()
}

// StagedNamesArgs builds arguments for listing the paths staged for the next commit.
func StagedNamesArgs() []string {
	return NewArgsBuilder().
		Add(SubCmdDiff, OptCached, OptNameOnly).
		Build()
}

// CurrentBranchArgs builds arguments for naming the ref HEAD is on, in full. It exits
// non-zero on a detached HEAD. Not --short: with a tag of the same name that prints
// heads/<branch>, so a comparison with the branch input would falsely fail.
func CurrentBranchArgs() []string {
	return NewArgsBuilder().
		Add(SubCmdSymbolicRef, OptQuiet, RefHead).
		Build()
}

// FullRefArgs builds arguments for resolving a name to its full ref, as a push would.
func FullRefArgs(name string) []string {
	return NewArgsBuilder().
		Add(SubCmdRevParse, OptSymbolicFullName, name).
		Build()
}

// BranchRef returns the full ref CurrentBranchArgs prints for branch.
func BranchRef(branch string) string {
	return RefHeads + branch
}

// RebaseAbortArgs builds arguments for abandoning a rebase in progress.
func RebaseAbortArgs() []string {
	return NewArgsBuilder().
		Add(SubCmdRebase, OptAbort).
		Build()
}

// IsAncestorArgs builds arguments for testing whether ancestor is contained in
// descendant: exit 0 if it is, exit 1 if it is not.
func IsAncestorArgs(ancestor, descendant string) []string {
	return NewArgsBuilder().
		Add(SubCmdMergeBase, OptIsAncestor, ancestor, descendant).
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
// Two dots compare the tips directly: three dots need a merge base, which a
// shallow Actions checkout usually lacks ("no merge base", exit 128).
func DiffNameOnlyArgs(base, head string) []string {
	return NewArgsBuilder().
		Add(SubCmdDiff, base+".."+head, OptNameOnly).
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

// ConfigGetArgs builds arguments for getting a config value.
func ConfigGetArgs(key string) []string {
	return NewArgsBuilder().
		Add(SubCmdConfig, OptGet, key).
		Build()
}

// RemoteSetURLArgs builds arguments for setting a remote URL.
func RemoteSetURLArgs(remote, url string) []string {
	return NewArgsBuilder().
		Add(SubCmdRemote, OptSetURL, remote, url).
		Build()
}

// PushDeleteBranchArgs builds arguments for deleting a remote branch.
func PushDeleteBranchArgs(remote, branch string) []string {
	return NewArgsBuilder().
		Add(SubCmdPush, remote, OptDeleteRemote, branch).
		Build()
}
