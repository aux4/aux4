package output

import (
	"os"
	"strings"
	"testing"
)

func envLookup(env map[string]string) func(string) (string, bool) {
	return func(name string) (string, bool) {
		value, ok := env[name]
		return value, ok
	}
}

func TestResolveColorPrecedence(t *testing.T) {
	cases := []struct {
		name             string
		env              map[string]string
		stdoutIsTerminal bool
		want             bool
	}{
		// 1. NO_COLOR always wins.
		{"no color on a terminal", map[string]string{"NO_COLOR": "1"}, true, false},
		{"no color piped", map[string]string{"NO_COLOR": "1"}, false, false},
		{"no color beats clicolor force", map[string]string{"NO_COLOR": "1", "CLICOLOR_FORCE": "1"}, false, false},
		{"empty no color is not set", map[string]string{"NO_COLOR": ""}, true, true},

		// 2. CLICOLOR_FORCE inherits the ancestor's decision (nested aux4).
		{"clicolor force piped", map[string]string{"CLICOLOR_FORCE": "1"}, false, true},
		{"clicolor force beats dumb terminal", map[string]string{"CLICOLOR_FORCE": "1", "TERM": "dumb"}, false, true},
		{"clicolor force zero is not set", map[string]string{"CLICOLOR_FORCE": "0"}, false, false},
		{"empty clicolor force is not set", map[string]string{"CLICOLOR_FORCE": ""}, false, false},

		// 3. TERM=dumb means the terminal cannot render color.
		{"dumb terminal", map[string]string{"TERM": "dumb"}, true, false},

		// 4. stdout decides.
		{"terminal", map[string]string{"TERM": "xterm-256color"}, true, true},
		{"pipe", map[string]string{"TERM": "xterm-256color"}, false, false},
		{"empty environment piped", map[string]string{}, false, false},
		{"empty environment on a terminal", map[string]string{}, true, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := resolveColor(envLookup(c.env), c.stdoutIsTerminal); got != c.want {
				t.Errorf("resolveColor(%v, %v) = %v, want %v", c.env, c.stdoutIsTerminal, got, c.want)
			}
		})
	}
}

func TestColorEnvIsMutuallyExclusive(t *testing.T) {
	// The parent environment carries a stale decision that must be overridden,
	// never merged: a child must never see both variables.
	parent := []string{"PATH=/usr/bin", "NO_COLOR=1", "CLICOLOR_FORCE=1", "TERM=xterm"}

	t.Run("color enabled", func(t *testing.T) {
		setColorEnabled(true)
		env := ColorEnv(parent)
		assertEnv(t, env, "CLICOLOR_FORCE=1")
		assertNoEnv(t, env, "NO_COLOR")
		assertEnv(t, env, "PATH=/usr/bin")
		assertEnv(t, env, "TERM=xterm")
	})

	t.Run("color disabled", func(t *testing.T) {
		setColorEnabled(false)
		env := ColorEnv(parent)
		assertEnv(t, env, "NO_COLOR=1")
		assertNoEnv(t, env, "CLICOLOR_FORCE")
		assertEnv(t, env, "PATH=/usr/bin")
	})
}

func TestColorEnvMapIsMutuallyExclusive(t *testing.T) {
	setColorEnabled(true)
	env := ColorEnvMap(map[string]string{"NO_COLOR": "1", "PATH": "/usr/bin"})
	if env["CLICOLOR_FORCE"] != "1" {
		t.Errorf("expected CLICOLOR_FORCE=1, got %q", env["CLICOLOR_FORCE"])
	}
	if _, ok := env["NO_COLOR"]; ok {
		t.Error("expected NO_COLOR to be removed")
	}

	setColorEnabled(false)
	env = ColorEnvMap(map[string]string{"CLICOLOR_FORCE": "1", "PATH": "/usr/bin"})
	if env["NO_COLOR"] != "1" {
		t.Errorf("expected NO_COLOR=1, got %q", env["NO_COLOR"])
	}
	if _, ok := env["CLICOLOR_FORCE"]; ok {
		t.Error("expected CLICOLOR_FORCE to be removed")
	}
	if env["PATH"] != "/usr/bin" {
		t.Error("expected the rest of the environment to be preserved")
	}
}

func TestResolveColorUsesTheEnvironment(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	if ResolveColor() {
		t.Error("expected color to be disabled when NO_COLOR is set")
	}

	t.Setenv("NO_COLOR", "")
	t.Setenv("CLICOLOR_FORCE", "1")
	if !ResolveColor() {
		t.Error("expected color to be enabled when CLICOLOR_FORCE is set")
	}
}

func TestColorFunctionsRespectTheDecision(t *testing.T) {
	setColorEnabled(false)
	if got := Red("boom"); got != "boom" {
		t.Errorf("Red(%q) = %q, want plain text", "boom", got)
	}
	if got := Bold("title"); got != "title" {
		t.Errorf("Bold(%q) = %q, want plain text", "title", got)
	}
	if got := FormatReset(); got != "" {
		t.Errorf("FormatReset() = %q, want empty", got)
	}
	if got := ColorText("hello world", "world", ColorCyan); got != "hello world" {
		t.Errorf("ColorText() = %q, want plain text", got)
	}

	setColorEnabled(true)
	if got := Red("boom"); !strings.Contains(got, ansiRed) {
		t.Errorf("Red(%q) = %q, want an ANSI code", "boom", got)
	}
	if got := ColorText("hello world", "world", ColorCyan); !strings.Contains(got, ansiCyan) {
		t.Errorf("ColorText() = %q, want an ANSI code", got)
	}
}

func TestIsTerminal(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "colorpolicy")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	if isTerminal(file) {
		t.Error("expected a regular file not to be a terminal")
	}
	if isTerminal(nil) {
		t.Error("expected nil not to be a terminal")
	}
}

func assertEnv(t *testing.T, env []string, entry string) {
	t.Helper()
	for _, e := range env {
		if e == entry {
			return
		}
	}
	t.Errorf("expected %q in %v", entry, env)
}

func assertNoEnv(t *testing.T, env []string, name string) {
	t.Helper()
	for _, e := range env {
		if strings.HasPrefix(e, name+"=") {
			t.Errorf("expected no %s in %v", name, env)
		}
	}
}
