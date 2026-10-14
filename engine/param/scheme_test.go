package param

import "testing"

func TestParseSchemeURI(t *testing.T) {
	cases := []struct {
		value      string
		wantScheme string
		wantOk     bool
	}{
		{"cloud://my-package/path/to/file.txt", "cloud", true},
		{"secret://1password/Work/Jira/password", "secret", true},
		{"s3://bucket/key", "s3", true},
		{"/local/path/file.txt", "", false},
		{"file.txt", "", false},
		{"", "", false},
		{"://missing-scheme", "", false},
		{"relative/path:with:colons", "", false},
	}

	for _, c := range cases {
		scheme, ok := ParseSchemeURI(c.value)
		if ok != c.wantOk {
			t.Errorf("ParseSchemeURI(%q) ok = %v, want %v", c.value, ok, c.wantOk)
			continue
		}
		if ok && scheme != c.wantScheme {
			t.Errorf("ParseSchemeURI(%q) scheme = %q, want %q", c.value, scheme, c.wantScheme)
		}
	}
}
