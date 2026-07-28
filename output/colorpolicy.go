package output

import (
	"os"
	"strings"
	"sync"
)

// Environment variables that carry the color decision.
const (
	EnvNoColor       = "NO_COLOR"
	EnvClicolorForce = "CLICOLOR_FORCE"
)

// The color decision is resolved once per process and cached. aux4 never
// rewrites the bytes a child process produces — it only tells the child what
// the user wants by propagating exactly one of NO_COLOR / CLICOLOR_FORCE.
var (
	colorMutex    sync.RWMutex
	colorResolved bool
	colorEnabled  bool
)

// ColorEnabled reports whether aux4 (and the commands it runs) should emit
// color. The decision is resolved the first time it is needed and cached for
// the rest of the process.
func ColorEnabled() bool {
	colorMutex.RLock()
	if colorResolved {
		defer colorMutex.RUnlock()
		return colorEnabled
	}
	colorMutex.RUnlock()

	colorMutex.Lock()
	defer colorMutex.Unlock()

	if !colorResolved {
		colorEnabled = resolveColor(os.LookupEnv, isTerminal(os.Stdout))
		colorResolved = true
	}

	return colorEnabled
}

// ResolveColor recomputes the color decision from the current environment and
// the current os.Stdout, replacing the cached value. The daemon uses it: each
// request runs with the client's environment and a fresh os.Stdout, so the
// decision has to be taken again per request.
func ResolveColor() bool {
	setColorEnabled(resolveColor(os.LookupEnv, isTerminal(os.Stdout)))
	return ColorEnabled()
}

// setColorEnabled overrides the cached decision, and marks it as resolved so
// ColorEnabled does not recompute it afterwards.
func setColorEnabled(enabled bool) {
	colorMutex.Lock()
	defer colorMutex.Unlock()

	colorEnabled = enabled
	colorResolved = true
}

// resolveColor applies the aux4 color policy, in precedence order:
//
//  1. NO_COLOR set          -> no color (explicit opt-out always wins)
//  2. CLICOLOR_FORCE set    -> color   (inherit the ancestor's decision)
//  3. TERM=dumb             -> no color
//  4. stdout is a terminal  -> color
//  5. otherwise             -> no color
//
// Rule 2 is what keeps composition working: an aux4 command that runs aux4
// always has a pipe for stdout, so without inheriting the ancestor's decision
// every nested invocation would resolve "not a terminal" and kill color for the
// whole chain.
func resolveColor(lookupEnv func(string) (string, bool), stdoutIsTerminal bool) bool {
	if noColorSet(lookupEnv) {
		return false
	}

	if clicolorForceSet(lookupEnv) {
		return true
	}

	if term, ok := lookupEnv("TERM"); ok && term == "dumb" {
		return false
	}

	return stdoutIsTerminal
}

// noColorSet follows the no-color.org convention: the variable counts as set
// when it is present with any value other than the empty string.
func noColorSet(lookupEnv func(string) (string, bool)) bool {
	value, ok := lookupEnv(EnvNoColor)
	return ok && value != ""
}

// clicolorForceSet follows the CLICOLOR_FORCE convention: present and not "0".
func clicolorForceSet(lookupEnv func(string) (string, bool)) bool {
	value, ok := lookupEnv(EnvClicolorForce)
	return ok && value != "" && value != "0"
}

// IsTerminal reports whether the given file is attached to a character device
// (a terminal). It is the single place aux4 asks that question.
func IsTerminal(file *os.File) bool {
	return isTerminal(file)
}

func isTerminal(file *os.File) bool {
	if file == nil {
		return false
	}

	info, err := file.Stat()
	if err != nil {
		return false
	}

	return info.Mode()&os.ModeCharDevice != 0
}

// ChildEnv returns the environment for a child process: the current
// environment with the resolved color decision propagated. NO_COLOR and
// CLICOLOR_FORCE are mutually exclusive in the result — whichever value was
// inherited is dropped and replaced by the decision aux4 took.
func ChildEnv() []string {
	return ColorEnv(os.Environ())
}

// ColorEnv applies the resolved color decision to the given environment.
func ColorEnv(env []string) []string {
	result := make([]string, 0, len(env)+1)

	for _, entry := range env {
		if strings.HasPrefix(entry, EnvNoColor+"=") || strings.HasPrefix(entry, EnvClicolorForce+"=") {
			continue
		}
		result = append(result, entry)
	}

	if ColorEnabled() {
		return append(result, EnvClicolorForce+"=1")
	}

	return append(result, EnvNoColor+"=1")
}

// ColorEnvMap applies the resolved color decision to an environment map, in
// place. Used by the daemon client, which ships its environment as a map.
func ColorEnvMap(env map[string]string) map[string]string {
	if env == nil {
		return env
	}

	if ColorEnabled() {
		delete(env, EnvNoColor)
		env[EnvClicolorForce] = "1"
	} else {
		delete(env, EnvClicolorForce)
		env[EnvNoColor] = "1"
	}

	return env
}
