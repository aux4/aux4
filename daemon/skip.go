package daemon

import (
	"os"
	"strings"
)

// NoDaemonFlag is the reserved aux4-level flag that makes an invocation run
// directly instead of being forwarded into a running daemon. It is consumed by
// aux4 core and is never forwarded to the command itself. It follows aux4's
// camelCase flag convention (like --noBuild, --configFile).
const NoDaemonFlag = "--noDaemon"

// NoDaemonEnv is the environment variable that, when set to "1", makes every
// invocation bypass the daemon. Unlike the flag it is inherited by child
// processes, so it is an explicit escape hatch rather than the primary
// mechanism.
const NoDaemonEnv = "AUX4_NO_DAEMON"

// ExtractNoDaemonFlag scans args for the bare boolean --noDaemon flag, removes
// every occurrence, and reports whether the flag is enabled. It returns the
// cleaned args so the flag never reaches command parsing.
//
// The flag is a boolean: `--noDaemon` (or `--noDaemon=true`) turns it on;
// `--noDaemon=false` (or =0/no/off) turns it off. Crucially it never consumes
// the following argument, so `aux4 --noDaemon mcp` keeps `mcp` as the command
// rather than treating it as the flag's value.
//
// When the flag is absent the args are returned unchanged, so default behavior
// is untouched.
func ExtractNoDaemonFlag(args []string) (bool, []string) {
	enabled := false
	found := false
	cleaned := make([]string, 0, len(args))

	for _, arg := range args {
		if arg == NoDaemonFlag {
			enabled = true
			found = true
			continue
		}
		if strings.HasPrefix(arg, NoDaemonFlag+"=") {
			value := strings.TrimPrefix(arg, NoDaemonFlag+"=")
			enabled = isEnabledValue(value)
			found = true
			continue
		}
		cleaned = append(cleaned, arg)
	}

	if !found {
		return false, args
	}

	return enabled, cleaned
}

// isEnabledValue mirrors Parameters.IsEnabled: an empty value or one of
// false/0/no/off means off, anything else means on.
func isEnabledValue(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "false", "0", "no", "off":
		return false
	}
	return true
}

// SkipForwarding reports whether the current invocation should bypass the
// daemon and run the command directly. It is true when the --noDaemon flag was
// given (flagSet) OR when AUX4_NO_DAEMON=1 is present in the environment.
//
// The flag applies only to the invocation it is on: it lives in the command
// line and is not propagated to any subprocess. A long-running server started
// with `aux4 --noDaemon mcp` therefore runs directly, while the plain
// `aux4 <toolcmd>` subprocesses it spawns carry no flag and still forward to the
// daemon, keeping daemon speed for tool calls.
func SkipForwarding(flagSet bool) bool {
	if flagSet {
		return true
	}
	return os.Getenv(NoDaemonEnv) == "1"
}
