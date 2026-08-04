package gitcmd

import (
	"strings"
	"sync"
)

// Call is a single command invocation recorded by FakeRunner.
type Call struct {
	Name string
	Args []string
}

// Key renders the call as a single space-joined string ("git status --porcelain"),
// the same form FakeRunner uses to look up canned results.
func (c Call) Key() string {
	if len(c.Args) == 0 {
		return c.Name
	}
	return c.Name + " " + strings.Join(c.Args, " ")
}

// FakeResult is the canned outcome of one command.
type FakeResult struct {
	// Stdout is returned by Output. Run ignores it.
	Stdout string
	// Err is returned by both Run and Output. Use Fail to build an error that
	// carries a specific exit code.
	Err error
}

// Fail returns an error representing a command that ran and exited with code.
func Fail(code int) error {
	return &ExitError{Code: code}
}

// FakeRunner is a test double for Runner. It records every invocation and
// returns canned results, letting tests assert the exact git command sequence a
// code path emits without touching a real repository.
//
// The zero value is usable: every command succeeds with empty output.
type FakeRunner struct {
	mu    sync.Mutex
	calls []Call

	// Handler, when non-nil, takes precedence over Results and Default and
	// decides the outcome of every call. Use it for stateful fakes (for
	// example, a status that changes after an add).
	Handler func(name string, args []string) (string, error)

	// Results maps a Call.Key to its canned outcome. A command with no entry
	// falls back to Default.
	Results map[string]FakeResult

	// Default applies to any command not present in Results.
	Default FakeResult
}

// NewFakeRunner returns a FakeRunner with an initialized Results map.
func NewFakeRunner() *FakeRunner {
	return &FakeRunner{Results: make(map[string]FakeResult)}
}

// Stub registers a canned result for the given command key and returns the
// runner so calls can be chained.
func (f *FakeRunner) Stub(key string, res FakeResult) *FakeRunner {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.Results == nil {
		f.Results = make(map[string]FakeResult)
	}
	f.Results[key] = res
	return f
}

// Run records the call and returns the canned error for it.
func (f *FakeRunner) Run(name string, args ...string) error {
	_, err := f.resolve(name, args)
	return err
}

// Output records the call and returns the canned stdout and error for it.
func (f *FakeRunner) Output(name string, args ...string) ([]byte, error) {
	out, err := f.resolve(name, args)
	if err != nil {
		return nil, err
	}
	return []byte(out), nil
}

// resolve records the call and computes its outcome.
func (f *FakeRunner) resolve(name string, args []string) (string, error) {
	f.mu.Lock()
	call := Call{Name: name, Args: append([]string(nil), args...)}
	f.calls = append(f.calls, call)
	handler := f.Handler
	res, ok := f.Results[call.Key()]
	if !ok {
		res = f.Default
	}
	f.mu.Unlock()

	if handler != nil {
		return handler(name, args)
	}
	return res.Stdout, res.Err
}

// Calls returns a copy of the recorded invocations in order.
func (f *FakeRunner) Calls() []Call {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]Call(nil), f.calls...)
}

// Keys returns the recorded invocations rendered as command strings, which is
// the most convenient form for asserting a command sequence.
func (f *FakeRunner) Keys() []string {
	calls := f.Calls()
	keys := make([]string, len(calls))
	for i, c := range calls {
		keys[i] = c.Key()
	}
	return keys
}

// Ran reports whether a command matching key was invoked.
func (f *FakeRunner) Ran(key string) bool {
	for _, k := range f.Keys() {
		if k == key {
			return true
		}
	}
	return false
}

// Reset clears the recorded calls, keeping the configured results.
func (f *FakeRunner) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = nil
}
