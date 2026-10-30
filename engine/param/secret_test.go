package param

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// CORE-046 sweep: fetchSecrets (used by ResolveSecrets/ResolveSingleSecret
// for every "secret://provider/ref/.../field" default or set: value) built
// "aux4 secret %s get --ref '%s'%s" via a bare fmt.Sprintf. provider was
// spliced in completely UNQUOTED, and ref/fields were wrapped in single
// quotes but never escaped -- all three come straight out of splitting a
// secret:// URI, which can itself be an attacker/config-authored value
// (e.g. a package or config.yaml default).

// withFakeSecretProvider puts a fake "aux4" on PATH that understands
// "secret <provider> get --ref <ref> --fields <fields>" and answers with a
// JSON object mapping every requested field to a literal string, so
// ResolveSingleSecret completes its round trip.
func withFakeSecretProvider(t *testing.T) (captureFile string) {
	t.Helper()
	dir := t.TempDir()
	captureFile = filepath.Join(dir, "capture.txt")

	script := "#!/bin/sh\n" +
		"for a in \"$@\"; do printf '%s\\n' \"$a\" >> \"" + captureFile + "\"; done\n" +
		"echo '{\"username\":\"resolved\",\"password\":\"resolved\",\"credential\":\"resolved\"}'\n"

	fakeAux4 := filepath.Join(dir, "aux4")
	if err := os.WriteFile(fakeAux4, []byte(script), 0755); err != nil {
		t.Fatalf("failed to write fake aux4: %v", err)
	}

	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return captureFile
}

// TestFetchSecretsEscapesInjectionInProvider proves a secret:// URI whose
// provider segment is itself a command substitution (no slash required,
// e.g. "secret://$(cmd)/vault/item/field") cannot run that command. Before
// the fix, provider was interpolated completely unquoted.
func TestFetchSecretsEscapesInjectionInProvider(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "pwned")
	withFakeSecretProvider(t)

	uri := "secret://$(touch " + marker + ")/vault/item/credential"
	_, _ = ResolveSingleSecret(uri)

	if _, err := exec.Command("test", "-e", marker).CombinedOutput(); err == nil {
		t.Fatalf("command injection via secret:// provider segment: marker file was created")
	}
}

// TestFetchSecretsEscapesInjectionInRefAndField proves a single-quote /
// shell-metacharacter payload inside the ref or field segment cannot break
// out of the manually single-quoted --ref/--fields arguments.
func TestFetchSecretsEscapesInjectionInRefAndField(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "pwned")
	withFakeSecretProvider(t)

	payload := "'; touch " + marker + "; '"

	// payload in the ref segment
	uriRef := "secret://1password/" + payload + "/credential"
	_, _ = ResolveSingleSecret(uriRef)

	// payload in the field (final) segment
	uriField := "secret://1password/vault/item/" + payload
	_, _ = ResolveSingleSecret(uriField)

	if _, err := exec.Command("test", "-e", marker).CombinedOutput(); err == nil {
		t.Fatalf("command injection via secret:// ref/field segment: marker file was created")
	}
}

// TestFetchSecretsNormalURIStillWorks proves an ordinary secret:// URI still
// resolves to the expected field, byte-identical to before the fix, and
// that provider/ref/fields reach the subprocess as the exact same argv
// values that were given.
func TestFetchSecretsNormalURIStillWorks(t *testing.T) {
	captureFile := withFakeSecretProvider(t)

	value, err := ResolveSingleSecret("secret://1password/Work/GitHub/credential")
	if err != nil {
		t.Fatalf("ResolveSingleSecret returned error: %v", err)
	}
	if value != "resolved" {
		t.Errorf("got %q, want %q", value, "resolved")
	}

	data, err := os.ReadFile(captureFile)
	if err != nil {
		t.Fatalf("fake aux4 was never invoked: %v", err)
	}
	got := string(data)
	want := "secret\n1password\nget\n--ref\nWork/GitHub\n--fields\ncredential\n"
	if got != want {
		t.Errorf("got args %q, want %q", got, want)
	}
}
