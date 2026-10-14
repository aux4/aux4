package executor

import (
	"os"
	"testing"
)

func TestSchemeURIPrefixEnd(t *testing.T) {
	cases := []struct {
		expression string
		want       int
	}{
		{"cloud://my-package/a.txt:content", len("cloud://")},
		{"s3://bucket/key:content", len("s3://")},
		{"${dest}:content", -1},
		{"plain/path.txt:content", -1},
		{"content with http://example.com inside", -1}, // scheme not at start
	}

	for _, c := range cases {
		got := schemeURIPrefixEnd(c.expression)
		if got != c.want {
			t.Errorf("schemeURIPrefixEnd(%q) = %d, want %d", c.expression, got, c.want)
		}
	}
}

func TestSplitFileExpression_PlainPath(t *testing.T) {
	path, content, err := splitFileExpression("output.txt:hello world")
	if err != nil {
		t.Fatal(err)
	}
	if path != "output.txt" || content != "hello world" {
		t.Errorf("got path=%q content=%q", path, content)
	}
}

func TestSplitFileExpression_SchemeLiteralPath(t *testing.T) {
	path, content, err := splitFileExpression("cloud://my-package/a.txt:hello world")
	if err != nil {
		t.Fatal(err)
	}
	if path != "cloud://my-package/a.txt" || content != "hello world" {
		t.Errorf("got path=%q content=%q", path, content)
	}
}

func TestSplitFileExpression_SchemeLiteralPathWithColonInContent(t *testing.T) {
	path, content, err := splitFileExpression("cloud://my-package/a.txt:key: value")
	if err != nil {
		t.Fatal(err)
	}
	if path != "cloud://my-package/a.txt" || content != "key: value" {
		t.Errorf("got path=%q content=%q", path, content)
	}
}

func TestSplitFileExpression_VariablePath(t *testing.T) {
	// A ${var} path is never a literal scheme URI at parse time, so it still
	// splits on the first ":" exactly like before.
	path, content, err := splitFileExpression("${dest}:hello")
	if err != nil {
		t.Fatal(err)
	}
	if path != "${dest}" || content != "hello" {
		t.Errorf("got path=%q content=%q", path, content)
	}
}

func TestSplitFileExpression_Invalid(t *testing.T) {
	_, _, err := splitFileExpression("no-colon-here")
	if err == nil {
		t.Fatal("expected an error for a missing ':' separator")
	}
}

func TestHashFile_SameContentSameHash(t *testing.T) {
	f1, err := os.CreateTemp("", "scheme-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f1.Name())
	f1.WriteString("hello world")
	f1.Close()

	f2, err := os.CreateTemp("", "scheme-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f2.Name())
	f2.WriteString("hello world")
	f2.Close()

	h1, err := hashFile(f1.Name())
	if err != nil {
		t.Fatal(err)
	}
	h2, err := hashFile(f2.Name())
	if err != nil {
		t.Fatal(err)
	}
	if h1 != h2 {
		t.Errorf("expected identical content to hash the same: %q != %q", h1, h2)
	}
}

func TestHashFile_DifferentContentDifferentHash(t *testing.T) {
	f1, err := os.CreateTemp("", "scheme-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f1.Name())
	f1.WriteString("hello")
	f1.Close()

	f2, err := os.CreateTemp("", "scheme-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f2.Name())
	f2.WriteString("world")
	f2.Close()

	h1, _ := hashFile(f1.Name())
	h2, _ := hashFile(f2.Name())
	if h1 == h2 {
		t.Error("expected different content to hash differently")
	}
}
