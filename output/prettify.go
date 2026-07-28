package output

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"sync/atomic"
)

// PrettifyParameter is the reserved global parameter that turns the JSON output
// of any command into indented JSON. It is consumed by aux4 core and is never
// forwarded to the command itself.
const PrettifyParameter = "prettify"

const prettifyIndent = "  "

var prettifyEnabled atomic.Bool

// SetPrettify records the --prettify decision for this invocation. It is called
// from the entry points that parse the command line (the CLI, the shell and the
// daemon), so a nested in-process execution keeps the outer decision.
func SetPrettify(enabled bool) {
	prettifyEnabled.Store(enabled)
}

// PrettifyEnabled reports whether --prettify was given.
func PrettifyEnabled() bool {
	return prettifyEnabled.Load()
}

// Prettify indents the JSON in data.
//
// data is treated as a STREAM of JSON values, which covers a single document
// and NDJSON (or any concatenation of values) with the same code path: each
// value is decoded and emitted indented on its own.
//
// The indenting is done with json.Indent over the RAW bytes — never by
// unmarshalling into interface{} and marshalling back, which would turn objects
// into Go maps and alphabetize their keys.
//
// If data is not valid JSON, or stops being valid part way through, the
// ORIGINAL bytes are returned unchanged. --prettify never corrupts and never
// swallows output; at worst it does nothing.
func Prettify(data []byte) []byte {
	if len(bytes.TrimSpace(data)) == 0 {
		return data
	}

	decoder := json.NewDecoder(bytes.NewReader(data))

	var result bytes.Buffer
	for {
		var value json.RawMessage

		if err := decoder.Decode(&value); err != nil {
			if err == io.EOF {
				break
			}
			return data
		}

		var indented bytes.Buffer
		if err := json.Indent(&indented, value, "", prettifyIndent); err != nil {
			return data
		}

		result.Write(indented.Bytes())
		result.WriteByte('\n')
	}

	if result.Len() == 0 {
		return data
	}

	return result.Bytes()
}

// WriteOutput writes the output of a command, indenting it first when
// --prettify is on. Without the flag the bytes are written exactly as they were
// produced.
func WriteOutput(writer io.Writer, data []byte) {
	if len(data) == 0 {
		return
	}

	if PrettifyEnabled() {
		data = Prettify(data)
	}

	writer.Write(data)
}

// PrintResponse writes aux4's own ${response} output to stdout, indenting it
// when --prettify is on, and always terminating it with a single newline.
func PrintResponse(text string) {
	data := []byte(text)

	if PrettifyEnabled() {
		data = Prettify(data)
	}

	os.Stdout.Write(data)

	if len(data) == 0 || data[len(data)-1] != '\n' {
		os.Stdout.Write([]byte("\n"))
	}
}
