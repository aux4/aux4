package output

import (
	"os"
	"testing"
)

// The daemon always reads stdin from a pipe, so it relies entirely on the
// AUX4_STDIN_TTY hint the client sets. These tests pin that contract: when the
// hint is present it is authoritative, "1" meaning interactive and anything else
// meaning not.
func TestStdinIsInteractiveHonorsTheEnvHint(t *testing.T) {
	cases := []struct {
		value string
		want  bool
	}{
		{"1", true},
		{"0", false},
		{"", false},
		{"true", false}, // only the exact "1" enables prompting
	}

	for _, c := range cases {
		os.Setenv(EnvStdinInteractive, c.value)
		if got := StdinIsInteractive(); got != c.want {
			t.Errorf("StdinIsInteractive() with %s=%q = %v, want %v", EnvStdinInteractive, c.value, got, c.want)
		}
	}

	os.Unsetenv(EnvStdinInteractive)
}

// StdinEnvMap must always record a decision so the daemon never falls back to
// probing its own (piped) stdin.
func TestStdinEnvMapAlwaysRecordsADecision(t *testing.T) {
	env := StdinEnvMap(map[string]string{})
	value, ok := env[EnvStdinInteractive]
	if !ok {
		t.Fatalf("StdinEnvMap did not set %s", EnvStdinInteractive)
	}
	if value != "0" && value != "1" {
		t.Errorf("StdinEnvMap set %s=%q, want \"0\" or \"1\"", EnvStdinInteractive, value)
	}

	if StdinEnvMap(nil) != nil {
		t.Errorf("StdinEnvMap(nil) should return nil")
	}
}
