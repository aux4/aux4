package param

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"aux4.dev/aux4/core"
)

// CORE-044 also changed ConfigLookup.Get (this file) to single-quote+escape
// the --file/--name values it shells out to "aux4 config get" with, matching
// the same escapeValue() treatment that value()/param()/values()/params()
// already use. These tests cover that part specifically; it shipped in
// 5.2.14 with no direct test coverage.
//
// lookupParams builds a Parameters the same way the engine does from real
// CLI args: --configFile <file> --config <name>.
func lookupParams(configFile string, config string) *Parameters {
	args := []string{}
	if configFile != "" {
		args = append(args, "--configFile", configFile)
	}
	if config != "" {
		args = append(args, "--config", config)
	}
	_, _, params := ParseArgs(args)
	return &params
}

// withFakeAux4 puts a fake "aux4" shell script first on PATH for the
// duration of the test. The fake script appends each argument it was
// called with (one per line, exactly as received) to captureFile, then
// prints a minimal valid JSON object so ConfigLookup.Get's own
// json.Unmarshal step succeeds.
func withFakeAux4(t *testing.T) (captureFile string) {
	t.Helper()
	dir := t.TempDir()
	captureFile = filepath.Join(dir, "capture.txt")

	script := "#!/bin/sh\n" +
		"for a in \"$@\"; do printf '%s\\n' \"$a\" >> \"" + captureFile + "\"; done\n" +
		"echo '{\"captured\":\"ok\"}'\n"

	fakeAux4 := filepath.Join(dir, "aux4")
	if err := os.WriteFile(fakeAux4, []byte(script), 0755); err != nil {
		t.Fatalf("failed to write fake aux4: %v", err)
	}

	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return captureFile
}

func readCaptured(t *testing.T, captureFile string) []string {
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

// TestConfigLookupBuildsExpectedCommand confirms the built "aux4 config get"
// invocation receives --file/--name as two argv entries each, with the
// value passed through byte-identical to what the caller supplied — i.e.
// the single-quote wrapping + escaping round-trips normal values correctly
// (spaces, '=' , JSON-looking text) exactly like value()/param() do.
func TestConfigLookupBuildsExpectedCommand(t *testing.T) {
	tests := []struct {
		name       string
		configFile string
		config     string
	}{
		{"plain", "config.yaml", "dev"},
		{"path with spaces", "my config/config.yaml", "dev"},
		{"nested config name", "config.yaml", "a/b"},
		{"equals and json-ish", "config.yaml", `key=value{"a":1}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			captureFile := withFakeAux4(t)
			lookup := NewConfigLookup()
			params := lookupParams(test.configFile, test.config)

			_, err := lookup.Get(params, core.Command{}, []string{}, "whatever")
			if err != nil {
				t.Fatalf("Get returned error: %v", err)
			}

			args := readCaptured(t, captureFile)
			want := []string{"config", "get", "--file", test.configFile, "--name", test.config}
			if fmt.Sprint(args) != fmt.Sprint(want) {
				t.Errorf("got args %q, want %q", args, want)
			}
		})
	}
}

// TestConfigLookupConfigTrueOmitsName mirrors the existing "config == true"
// special case (bare --config with no value) — only --file should be
// passed, never a --name flag.
func TestConfigLookupConfigTrueOmitsName(t *testing.T) {
	captureFile := withFakeAux4(t)
	lookup := NewConfigLookup()
	params := lookupParams("config.yaml", "true")

	_, err := lookup.Get(params, core.Command{}, []string{}, "whatever")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	args := readCaptured(t, captureFile)
	want := []string{"config", "get", "--file", "config.yaml"}
	if fmt.Sprint(args) != fmt.Sprint(want) {
		t.Errorf("got args %q, want %q", args, want)
	}
}

// TestConfigLookupEscapesInjectionInConfigFile proves a single-quote +
// shell-metacharacter payload in --configFile cannot break out of the
// quoted --file '...' argument and run an injected command. Before
// CORE-044 this value was interpolated with zero escaping.
func TestConfigLookupEscapesInjectionInConfigFile(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "pwned")
	payload := "'; touch " + marker + "; '"

	lookup := NewConfigLookup()
	params := lookupParams(payload, "dev")

	_, _ = lookup.Get(params, core.Command{}, []string{}, "whatever")

	if _, err := exec.Command("test", "-e", marker).CombinedOutput(); err == nil {
		t.Fatalf("command injection via --configFile: marker file was created")
	}
}

// TestConfigLookupEscapesInjectionInConfigName is the same proof for
// --config (the --name flag).
func TestConfigLookupEscapesInjectionInConfigName(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "pwned")
	payload := "'; touch " + marker + "; '"

	lookup := NewConfigLookup()
	params := lookupParams("config.yaml", payload)

	_, _ = lookup.Get(params, core.Command{}, []string{}, "whatever")

	if _, err := exec.Command("test", "-e", marker).CombinedOutput(); err == nil {
		t.Fatalf("command injection via --config: marker file was created")
	}
}

// TestConfigLookupEscapesVariousPayloads sweeps the same payload classes as
// the param(name**) coverage (CORE-044) through both --configFile and
// --config, confirming none of them reach the shell unescaped and none of
// them crash Get.
func TestConfigLookupEscapesVariousPayloads(t *testing.T) {
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
			lookup := NewConfigLookup()
			params := lookupParams(payload, "dev")
			_, _ = lookup.Get(params, core.Command{}, []string{}, "whatever")

			lookup2 := NewConfigLookup()
			params2 := lookupParams("config.yaml", payload)
			_, _ = lookup2.Get(params2, core.Command{}, []string{}, "whatever")
		})
	}

	if _, err := exec.Command("test", "-e", marker).CombinedOutput(); err == nil {
		t.Fatalf("command injection via ConfigLookup.Get: marker file was created from one of the payloads")
	}
}
