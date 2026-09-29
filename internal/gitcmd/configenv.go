package gitcmd

import (
	"fmt"
	"os"
	"strconv"
)

// AddConfigEnv appends key=value to the GIT_CONFIG_COUNT / GIT_CONFIG_KEY_<n> /
// GIT_CONFIG_VALUE_<n> environment, which every child git reads as
// command-scope config (like `git -c`), so no config file is written.
// Existing entries are kept; an unset or empty count is 0, as in git.
func AddConfigEnv(key, value string) error {
	n := 0
	if raw := os.Getenv("GIT_CONFIG_COUNT"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			return fmt.Errorf("invalid GIT_CONFIG_COUNT %q", raw)
		}
		n = parsed
	}
	idx := strconv.Itoa(n)
	if err := os.Setenv("GIT_CONFIG_KEY_"+idx, key); err != nil {
		return err
	}
	if err := os.Setenv("GIT_CONFIG_VALUE_"+idx, value); err != nil {
		return err
	}
	return os.Setenv("GIT_CONFIG_COUNT", strconv.Itoa(n+1))
}
