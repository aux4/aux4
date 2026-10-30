package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"aux4.dev/aux4/engine/param"
)

// CORE-046: configSecurityValue (resolveSecurityPolicy's internal
// "aux4 config get" shell-out) built "--file <configFile>"/"--name <config>"
// via plain string concatenation with ZERO quoting/escaping -- not even
// unquoted-but-safe. This runs on every invocation that sets --configFile/
// --config (or AUX4_CONFIG_FILE/AUX4_CONFIG), independent of whether the
// command declares any config-bound variable. These tests cover the fix,
// mirroring the ConfigLookup.Get coverage added for CORE-044/045 in
// engine/param/lookup_test.go.

// securityParams builds a *param.Parameters the same way the engine does
// from real CLI args: --configFile <file> --config <name>.
func securityParams(configFile string, config string) *param.Parameters {
	args := []string{}
	if configFile != "" {
		args = append(args, "--configFile", configFile)
	}
	if config != "" {
		args = append(args, "--config", config)
	}
	_, _, params := param.ParseArgs(args)
	return &params
}

// withFakeAux4ForSecurity puts a fake "aux4" shell script first on PATH for
// the duration of the test. It appends each argument it receives (one per
// line, exactly as received) to captureFile, then prints a minimal valid
// JSON object so configSecurityValue's own json.Unmarshal step succeeds.
func withFakeAux4ForSecurity(t *testing.T) (captureFile string) {
	t.Helper()
	dir := t.TempDir()
	captureFile = filepath.Join(dir, "capture.txt")

	script := "#!/bin/sh\n" +
		"for a in \"$@\"; do printf '%s\\n' \"$a\" >> \"" + captureFile + "\"; done\n" +
		"echo '{}'\n"

	fakeAux4 := filepath.Join(dir, "aux4")
	if err := os.WriteFile(fakeAux4, []byte(script), 0755); err != nil {
		t.Fatalf("failed to write fake aux4: %v", err)
	}

	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return captureFile
}

func readCapturedSecurity(t *testing.T, captureFile string) []string {
	t.Helper()
	data, err := os.ReadFile(captureFile)
	if err != nil {
		t.Fatalf("fake aux4 was never invoked (no capture file): %v", err)
	}
	lines := []string{}
	current := ""
	for _, r := range string(data) {
		if r == '\n' {
			lines = append(lines, current)
			current = ""
			continue
		}
		current += string(r)
	}
	return lines
}

// TestConfigSecurityValueBuildsExpectedCommand confirms the built
// "aux4 config get" invocation receives --file/--name as two argv entries
// each, with the value passed through byte-identical to what the caller
// supplied (spaces, '=', JSON-looking text) -- same contract as
// ConfigLookup.Get and value()/param().
func TestConfigSecurityValueBuildsExpectedCommand(t *testing.T) {
	tests := []struct {
		name       string
		configFile string
		config     string
	}{
		{"plain", "nonexistent-config.yaml", "dev"},
		{"path with spaces", "my config/nonexistent.yaml", "dev"},
		{"nested config name", "nonexistent-config.yaml", "a/b"},
		{"equals and json-ish", "nonexistent-config.yaml", `key=value{"a":1}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			captureFile := withFakeAux4ForSecurity(t)
			params := securityParams(test.configFile, test.config)

			_ = configSecurityValue(params)

			args := readCapturedSecurity(t, captureFile)
			want := []string{"config", "get", "--file", test.configFile, "--name", test.config}
			if fmt.Sprint(args) != fmt.Sprint(want) {
				t.Errorf("got args %q, want %q", args, want)
			}
		})
	}
}

// TestConfigSecurityValueConfigTrueOmitsName mirrors the "--config" (no
// value) special case: only --file should be passed, never a --name flag.
func TestConfigSecurityValueConfigTrueOmitsName(t *testing.T) {
	captureFile := withFakeAux4ForSecurity(t)
	params := securityParams("nonexistent-config.yaml", "true")

	_ = configSecurityValue(params)

	args := readCapturedSecurity(t, captureFile)
	want := []string{"config", "get", "--file", "nonexistent-config.yaml"}
	if fmt.Sprint(args) != fmt.Sprint(want) {
		t.Errorf("got args %q, want %q", args, want)
	}
}

// TestConfigSecurityValueEscapesInjectionInConfigFile is the exact repro
// from the CORE-046 report: a malicious --configFile must not be able to
// run an injected command via the internal "aux4 config get" shell-out.
// Before the fix, this created the marker file.
func TestConfigSecurityValueEscapesInjectionInConfigFile(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "pwned")
	payload := "$(touch " + marker + ")"

	params := securityParams(payload, "dev")
	_ = configSecurityValue(params)

	if _, err := exec.Command("test", "-e", marker).CombinedOutput(); err == nil {
		t.Fatalf("command injection via --configFile: marker file was created")
	}
}

// TestConfigSecurityValueEscapesInjectionInConfigName is the same proof for
// --config (the --name flag).
func TestConfigSecurityValueEscapesInjectionInConfigName(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "pwned")
	payload := "$(touch " + marker + ")"

	params := securityParams("nonexistent-config.yaml", payload)
	_ = configSecurityValue(params)

	if _, err := exec.Command("test", "-e", marker).CombinedOutput(); err == nil {
		t.Fatalf("command injection via --config: marker file was created")
	}
}

// TestConfigSecurityValueEscapesVariousPayloads sweeps the same payload
// classes as the CORE-044/045 coverage through both --configFile and
// --config, confirming none of them reach the shell unescaped and none of
// them crash configSecurityValue.
func TestConfigSecurityValueEscapesVariousPayloads(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "pwned")

	payloads := []string{
		"'",
		`"`,
		"$(touch " + marker + ")",
		"`touch " + marker + "`",
		"; touch " + marker + " ;",
		"*.txt",
		"héllo wörld 日本語 🎉",
	}

	for _, payload := range payloads {
		t.Run(payload, func(t *testing.T) {
			params := securityParams(payload, "dev")
			_ = configSecurityValue(params)

			params2 := securityParams("nonexistent-config.yaml", payload)
			_ = configSecurityValue(params2)
		})
	}

	if _, err := exec.Command("test", "-e", marker).CombinedOutput(); err == nil {
		t.Fatalf("command injection via configSecurityValue: marker file was created from one of the payloads")
	}
}

// TestConfigSecurityValueFastPathUnaffected proves the fast path (a real,
// readable configFile whose content does not mention "security") still
// skips the shell-out entirely -- the fix must not force every invocation
// through the subprocess.
func TestConfigSecurityValueFastPathUnaffected(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configFile, []byte("config:\n  dev:\n    host: localhost\n"), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	captureFile := filepath.Join(dir, "capture.txt")
	script := "#!/bin/sh\n" +
		"for a in \"$@\"; do printf '%s\\n' \"$a\" >> \"" + captureFile + "\"; done\n" +
		"echo '{}'\n"
	fakeAux4 := filepath.Join(dir, "aux4")
	if err := os.WriteFile(fakeAux4, []byte(script), 0755); err != nil {
		t.Fatalf("failed to write fake aux4: %v", err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	params := securityParams(configFile, "dev")
	result := configSecurityValue(params)

	if result != nil {
		t.Errorf("expected nil (no security policy), got %v", result)
	}
	if _, err := os.ReadFile(captureFile); err == nil {
		t.Errorf("fast path should have skipped the 'aux4 config get' shell-out entirely, but it ran")
	}
}
