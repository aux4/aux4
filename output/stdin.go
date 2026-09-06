package output

import (
	"os"

	"golang.org/x/term"
)

// EnvStdinInteractive carries the client's stdin-interactivity decision across
// the hop to the daemon, mirroring how NO_COLOR / CLICOLOR_FORCE carry the color
// decision. The daemon serves every request with a pipe for stdin (never a
// terminal), so on its own it can never tell whether a human with a terminal is
// waiting on the other end. The client owns the real stdin, so it records the
// answer here and the daemon reads it back.
const EnvStdinInteractive = "AUX4_STDIN_TTY"

// stdinIsTerminal reports whether os.Stdin is a real interactive terminal. It
// uses a proper TTY probe rather than the ModeCharDevice heuristic used for the
// color decision, because that heuristic treats /dev/null (a character device
// that is not a terminal) as interactive — exactly the case that must not
// prompt.
func stdinIsTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// StdinIsInteractive reports whether aux4 may prompt the user for input right
// now. It is the single place aux4 asks that question.
//
// In direct mode AUX4_STDIN_TTY is absent, so the answer is simply whether
// os.Stdin is a terminal. When the invocation was forwarded to the daemon the
// client set AUX4_STDIN_TTY (the daemon's own os.Stdin is always a pipe), and
// that decision is authoritative.
func StdinIsInteractive() bool {
	if value, ok := os.LookupEnv(EnvStdinInteractive); ok {
		return value == "1"
	}
	return stdinIsTerminal()
}

// StdinEnvMap records the client's stdin-interactivity decision into an
// environment map, in place, so it survives the hop to the daemon. Used by the
// daemon client, which ships its environment as a map.
func StdinEnvMap(env map[string]string) map[string]string {
	if env == nil {
		return env
	}

	if stdinIsTerminal() {
		env[EnvStdinInteractive] = "1"
	} else {
		env[EnvStdinInteractive] = "0"
	}

	return env
}
