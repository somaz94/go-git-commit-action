package gitcmd

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// isolateGitConfigEnv starts the test with no GIT_CONFIG_COUNT/KEY_<n>/VALUE_<n>
// set; t.Setenv restores the originals and the cleanup drops entries the test added.
func isolateGitConfigEnv(t *testing.T) {
	t.Helper()
	for _, name := range append(gitConfigEntryNames(), "GIT_CONFIG_COUNT") {
		t.Setenv(name, "")
		os.Unsetenv(name)
	}
	t.Cleanup(func() {
		for _, name := range gitConfigEntryNames() {
			os.Unsetenv(name)
		}
	})
}

func gitConfigEntryNames() []string {
	var names []string
	for _, kv := range os.Environ() {
		name, _, _ := strings.Cut(kv, "=")
		if strings.HasPrefix(name, "GIT_CONFIG_KEY_") || strings.HasPrefix(name, "GIT_CONFIG_VALUE_") {
			names = append(names, name)
		}
	}
	return names
}

// gitConfigEntries returns the injected entries as key=value, in index order.
func gitConfigEntries(t *testing.T) []string {
	t.Helper()
	n, err := strconv.Atoi(os.Getenv("GIT_CONFIG_COUNT"))
	if err != nil {
		t.Fatalf("GIT_CONFIG_COUNT: %v", err)
	}
	entries := make([]string, 0, n)
	for i := range n {
		entries = append(entries, os.Getenv(fmt.Sprintf("GIT_CONFIG_KEY_%d", i))+"="+os.Getenv(fmt.Sprintf("GIT_CONFIG_VALUE_%d", i)))
	}
	return entries
}

func TestAddConfigEnv_FirstEntry(t *testing.T) {
	isolateGitConfigEnv(t)

	if err := AddConfigEnv(ConfigSafeDirectory, PathGitHubWorkspace); err != nil {
		t.Fatalf("AddConfigEnv() error = %v, want nil", err)
	}

	want := []string{"safe.directory=/github/workspace"}
	if got := gitConfigEntries(t); !reflect.DeepEqual(got, want) {
		t.Errorf("entries = %v, want %v", got, want)
	}
}

func TestAddConfigEnv_EmptyCount(t *testing.T) {
	isolateGitConfigEnv(t)
	t.Setenv("GIT_CONFIG_COUNT", "")

	if err := AddConfigEnv(ConfigSafeDirectory, PathGitHubWorkspace); err != nil {
		t.Fatalf("AddConfigEnv() error = %v, want nil", err)
	}

	want := []string{"safe.directory=/github/workspace"}
	if got := gitConfigEntries(t); !reflect.DeepEqual(got, want) {
		t.Errorf("entries = %v, want %v", got, want)
	}
}

func TestAddConfigEnv_AppendsToExisting(t *testing.T) {
	isolateGitConfigEnv(t)
	t.Setenv("GIT_CONFIG_COUNT", "2")
	t.Setenv("GIT_CONFIG_KEY_0", "safe.directory")
	t.Setenv("GIT_CONFIG_VALUE_0", "/runner/work")
	t.Setenv("GIT_CONFIG_KEY_1", "core.autocrlf")
	t.Setenv("GIT_CONFIG_VALUE_1", "false")

	if err := AddConfigEnv(ConfigSafeDirectory, PathGitHubWorkspace); err != nil {
		t.Fatalf("AddConfigEnv() error = %v, want nil", err)
	}

	want := []string{
		"safe.directory=/runner/work",
		"core.autocrlf=false",
		"safe.directory=/github/workspace",
	}
	if got := gitConfigEntries(t); !reflect.DeepEqual(got, want) {
		t.Errorf("entries = %v, want %v", got, want)
	}
}

func TestAddConfigEnv_InvalidCount(t *testing.T) {
	for _, count := range []string{"abc", "-1", "1.5"} {
		t.Run(count, func(t *testing.T) {
			isolateGitConfigEnv(t)
			t.Setenv("GIT_CONFIG_COUNT", count)

			if err := AddConfigEnv(ConfigSafeDirectory, PathGitHubWorkspace); err == nil {
				t.Fatal("AddConfigEnv() error = nil, want an invalid GIT_CONFIG_COUNT error")
			}
			if got := os.Getenv("GIT_CONFIG_COUNT"); got != count {
				t.Errorf("GIT_CONFIG_COUNT = %q, want it unchanged at %q", got, count)
			}
			if names := gitConfigEntryNames(); len(names) != 0 {
				t.Errorf("entries written = %v, want none", names)
			}
		})
	}
}

func TestAddConfigEnv_RejectsNUL(t *testing.T) {
	for name, kv := range map[string][2]string{
		"key":   {"safe.\x00directory", PathGitHubWorkspace},
		"value": {ConfigSafeDirectory, "/github/\x00workspace"},
	} {
		t.Run(name, func(t *testing.T) {
			isolateGitConfigEnv(t)

			if err := AddConfigEnv(kv[0], kv[1]); err == nil {
				t.Fatal("AddConfigEnv() error = nil, want a NUL byte error")
			}
			if _, ok := os.LookupEnv("GIT_CONFIG_COUNT"); ok {
				t.Error("GIT_CONFIG_COUNT is set, want it to stay unset")
			}
		})
	}
}
