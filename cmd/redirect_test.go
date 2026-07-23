package cmd

import "testing"

func TestExpandRedirects(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"ignore all", "echo hi >ignore", "echo hi >/dev/null 2>&1"},
		{"ignore all at end", "cmd >ignore", "cmd >/dev/null 2>&1"},
		{"ignore error", "cmd >ignoreError", "cmd 2>/dev/null"},
		{"ignore output", "cmd >ignoreOutput", "cmd >/dev/null"},
		{"before semicolon", "a >ignore; b", "a >/dev/null 2>&1; b"},
		{"followed by text", "cmd >ignoreError next", "cmd 2>/dev/null next"},
		{"before pipe", "cmd >ignore | wc -l", "cmd >/dev/null 2>&1 | wc -l"},
		{"start of string", ">ignore x", ">/dev/null 2>&1 x"},

		// Must NOT touch real redirects or filenames.
		{"real fd redirect", "cmd 2>ignore", "cmd 2>ignore"},
		{"file named ignore.log", "cmd >ignore.log", "cmd >ignore.log"},
		{"unknown suffix", "cmd >ignoreFoo", "cmd >ignoreFoo"},
		{"error with trailing letter", "cmd >ignoreErrorX", "cmd >ignoreErrorX"},
		{"no tokens", "echo just a plain command", "echo just a plain command"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := expandRedirects(c.in); got != c.want {
				t.Errorf("expandRedirects(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}
