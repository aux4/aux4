package param

import (
	"strings"
	"testing"
)

// value()/param()/values()/params() recursively resolve ${...} references that
// appear inside a resolved value — so a builtin-based default like
// ${aux4HomeDir}/.oauth/google.json resolves the same way ${var} does — while
// still shell-escaping the result. But a reference that resolves to a secret is
// left literal, so untrusted data flowing through these functions can never name
// a secret variable and exfiltrate it.

func TestValueResolvesNestedReference(t *testing.T) {
	got := inject(t, "run value(path)", map[string]string{
		"path": "${home}/config.json",
		"home": "/Users/x/.aux4.config",
	})
	want := "run '/Users/x/.aux4.config/config.json'"
	if got != want {
		t.Errorf("value() did not resolve the nested ${home} in its value\n  got:  %q\n  want: %q", got, want)
	}
}

func TestParamResolvesNestedReference(t *testing.T) {
	got := inject(t, "curl param(tokenFile)", map[string]string{
		"tokenFile": "${home}/.oauth/google.json",
		"home":      "/Users/x/.aux4.config",
	})
	want := "curl --tokenFile '/Users/x/.oauth/google.json'"
	// param() uses the standardized flag name; assert the resolved path is present.
	if !strings.Contains(got, "/Users/x/.aux4.config/.oauth/google.json") {
		t.Errorf("param() did not resolve the nested ${home}\n  got: %q\n  (want it to contain the resolved path)", got)
	}
	_ = want
}

func TestValueDoesNotExpandSecretReference(t *testing.T) {
	got := inject(t, "run value(title)", map[string]string{
		"title":  "${apiKey}",
		"apiKey": "secret://fake/vault/item/field",
	})
	// The secret-backed reference must be left literal and shell-escaped — never
	// resolved (which would exfiltrate the secret).
	want := "run '${apiKey}'"
	if got != want {
		t.Errorf("value() expanded a secret-backed reference\n  got:  %q\n  want: %q", got, want)
	}
	if strings.Contains(got, "secret://") {
		t.Errorf("value() leaked the secret:// reference: %q", got)
	}
}

func TestValueStillEscapesShellInjection(t *testing.T) {
	got := inject(t, "run value(title)", map[string]string{
		"title": "x'; touch /tmp/pwned; echo '",
	})
	// The single quotes in the value must be escaped so `; touch` can never break
	// out and become its own shell command.
	if strings.Contains(got, "x'; touch") {
		t.Errorf("value() left an unescaped single-quote breakout: %q", got)
	}
	if !strings.Contains(got, `'\''`) {
		t.Errorf("value() did not shell-escape the single quotes: %q", got)
	}
}

func TestValueLeavesUnknownReferenceLiteral(t *testing.T) {
	got := inject(t, "run value(title)", map[string]string{
		"title": "${notADeclaredVar}/x",
	})
	// An unknown reference isn't expanded — it stays literal (and escaped).
	want := "run '${notADeclaredVar}/x'"
	if got != want {
		t.Errorf("value() mishandled an unknown reference\n  got:  %q\n  want: %q", got, want)
	}
}
