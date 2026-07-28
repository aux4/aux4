package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrettifyKeepsKeyOrder(t *testing.T) {
	// The whole point of using json.Indent on the raw bytes: unmarshalling into
	// interface{} and marshalling back would sort the keys alphabetically.
	input := []byte(`{"zebra":1,"apple":2,"mango":{"yellow":true,"green":false}}`)

	got := string(Prettify(input))
	want := `{
  "zebra": 1,
  "apple": 2,
  "mango": {
    "yellow": true,
    "green": false
  }
}
`

	if got != want {
		t.Errorf("Prettify() = %q, want %q", got, want)
	}
}

func TestPrettifyIndentsASingleDocument(t *testing.T) {
	got := string(Prettify([]byte(`[{"id":1},{"id":2}]`)))
	want := `[
  {
    "id": 1
  },
  {
    "id": 2
  }
]
`

	if got != want {
		t.Errorf("Prettify() = %q, want %q", got, want)
	}
}

func TestPrettifyIndentsEveryValueOfAStream(t *testing.T) {
	// NDJSON and any concatenation of values go through the same path.
	got := string(Prettify([]byte("{\"b\":1}\n{\"a\":2}\n")))
	want := `{
  "b": 1
}
{
  "a": 2
}
`

	if got != want {
		t.Errorf("Prettify() = %q, want %q", got, want)
	}

	if got := string(Prettify([]byte(`{"b":1} {"a":2}`))); !strings.Contains(got, "\"b\": 1") || !strings.Contains(got, "\"a\": 2") {
		t.Errorf("Prettify() = %q, want both values indented", got)
	}
}

func TestPrettifyKeepsNonJsonUntouched(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{"plain text", "hello world\n"},
		{"multi line text", "line one\nline two\n"},
		{"log line that starts like json", "{ not json after all\n"},
		{"trailing garbage after a valid value", "{\"a\":1}\nnot json\n"},
		{"truncated document", `{"a":1,`},
		{"empty", ""},
		{"whitespace only", "  \n\t"},
		{"binary", "\x00\x01\x02\xff"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Prettify([]byte(c.input))
			if !bytes.Equal(got, []byte(c.input)) {
				t.Errorf("Prettify(%q) = %q, want the original bytes", c.input, got)
			}
		})
	}
}

func TestPrettifyKeepsScalars(t *testing.T) {
	if got := string(Prettify([]byte("42"))); got != "42\n" {
		t.Errorf("Prettify(42) = %q", got)
	}
	if got := string(Prettify([]byte(`"text"`))); got != "\"text\"\n" {
		t.Errorf("Prettify(\"text\") = %q", got)
	}
}

func TestPrettifyDoesNotLoseNumberPrecision(t *testing.T) {
	// A round trip through interface{} would turn this into 1.2345678901234568e+19.
	input := []byte(`{"id":12345678901234567890}`)

	if got := string(Prettify(input)); !strings.Contains(got, "12345678901234567890") {
		t.Errorf("Prettify() = %q, want the original number", got)
	}
}

func TestWriteOutputOnlyIndentsWhenEnabled(t *testing.T) {
	raw := []byte(`{"b":1}`)

	SetPrettify(false)
	var plain bytes.Buffer
	WriteOutput(&plain, raw)
	if !bytes.Equal(plain.Bytes(), raw) {
		t.Errorf("WriteOutput() = %q, want the original bytes when --prettify is off", plain.Bytes())
	}

	SetPrettify(true)
	defer SetPrettify(false)

	var pretty bytes.Buffer
	WriteOutput(&pretty, raw)
	if pretty.String() != "{\n  \"b\": 1\n}\n" {
		t.Errorf("WriteOutput() = %q, want indented json", pretty.String())
	}

	var empty bytes.Buffer
	WriteOutput(&empty, nil)
	if empty.Len() != 0 {
		t.Errorf("WriteOutput(nil) wrote %q, want nothing", empty.String())
	}
}
