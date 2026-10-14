package executor

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"unicode"

	"aux4.dev/aux4/cmd"
	"aux4.dev/aux4/core"
	"aux4.dev/aux4/engine"
	"aux4.dev/aux4/engine/param"
)

// schemeFile tracks one declared file variable whose resolved value was a
// <scheme>:// URI, from the moment it is staged to a local temp file until it
// is put back (or discarded) once the command finishes.
type schemeFile struct {
	variable      string
	uri           string
	scheme        string
	tempFile      string
	mode          string // "read", "write" or "readwrite"
	existedBefore bool
	version       string
	beforeHash    string
}

// resolveSchemeFileVariables stages every declared file variable
// (help.variables[].file) whose resolved value is a <scheme>:// URI to a local
// temp file, substituting the temp file path into params so the command's own
// execute steps see a plain local path. Variables with no declared "file"
// mode, or whose value has no "://", are left completely untouched -- this is
// what keeps the feature strictly additive.
func resolveSchemeFileVariables(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) ([]*schemeFile, error) {
	if command.Help == nil || env == nil || env.Schemes == nil {
		return nil, nil
	}

	var staged []*schemeFile

	for _, variable := range command.Help.Variables {
		if variable == nil || variable.File == "" {
			continue
		}

		raw, err := params.Get(command, actions, variable.Name)
		if err != nil {
			cleanupSchemeFiles(staged)
			return nil, err
		}

		strVal, ok := raw.(string)
		if !ok || strVal == "" {
			continue
		}

		scheme, ok := param.ParseSchemeURI(strVal)
		if !ok {
			// Not a "<scheme>://..." value -- a plain local path, unchanged.
			continue
		}

		provider, found := env.Schemes.Get(scheme)
		if !found {
			cleanupSchemeFiles(staged)
			return nil, core.InternalError(fmt.Sprintf("no package provides %s://; install aux4/%s", scheme, scheme), nil)
		}

		tempFile, err := newSchemeTempFile(scheme)
		if err != nil {
			cleanupSchemeFiles(staged)
			return nil, err
		}

		exists, version, err := schemeGet(provider, strVal, tempFile)
		if err != nil {
			os.Remove(tempFile)
			cleanupSchemeFiles(staged)
			return nil, err
		}

		sf := &schemeFile{
			variable:      variable.Name,
			uri:           strVal,
			scheme:        scheme,
			tempFile:      tempFile,
			mode:          variable.File,
			existedBefore: exists,
			version:       version,
		}

		if exists {
			hash, hashErr := hashFile(tempFile)
			if hashErr != nil {
				cleanupSchemeFiles(staged)
				return nil, hashErr
			}
			sf.beforeHash = hash
		} else {
			// Nothing was fetched: present the variable with a path that does
			// not exist yet, exactly like a brand-new local file.
			os.Remove(tempFile)
		}

		staged = append(staged, sf)
		params.Update(variable.Name, tempFile)
	}

	return staged, nil
}

// finalizeSchemeFileVariables puts every "write"/"readwrite" staged file back
// to its remote URI if its content changed (or it was newly created), using
// the version read by resolveSchemeFileVariables as the optimistic-concurrency
// If-Match (or "none"/If-None-Match for a file that did not exist). It must
// only be called when the command itself succeeded -- a failed command never
// writes back. Temp files are always removed separately via cleanupSchemeFiles.
func finalizeSchemeFileVariables(env *engine.VirtualEnvironment, staged []*schemeFile) error {
	for _, sf := range staged {
		if sf.mode != "write" && sf.mode != "readwrite" {
			continue
		}

		if _, err := os.Stat(sf.tempFile); err != nil {
			if os.IsNotExist(err) {
				// The command never (re)created the file -- nothing to push.
				continue
			}
			return core.InternalError("Error reading "+sf.uri+" after command execution", err)
		}

		afterHash, err := hashFile(sf.tempFile)
		if err != nil {
			return err
		}

		if sf.existedBefore && afterHash == sf.beforeHash {
			continue
		}

		ifMatch := "none"
		if sf.existedBefore {
			ifMatch = sf.version
		}

		provider, found := env.Schemes.Get(sf.scheme)
		if !found {
			return core.InternalError(fmt.Sprintf("no package provides %s://; install aux4/%s", sf.scheme, sf.scheme), nil)
		}

		if err := schemePut(provider, sf.uri, sf.tempFile, ifMatch); err != nil {
			return err
		}
	}
	return nil
}

// cleanupSchemeFiles removes every staged temp file, regardless of outcome.
func cleanupSchemeFiles(staged []*schemeFile) {
	for _, sf := range staged {
		os.Remove(sf.tempFile)
	}
}

func newSchemeTempFile(scheme string) (string, error) {
	f, err := os.CreateTemp("", "aux4-"+scheme+"-*")
	if err != nil {
		return "", core.InternalError("Error creating temp file for "+scheme+"://", err)
	}
	path := f.Name()
	f.Close()
	return path, nil
}

func hashFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", core.InternalError("Error reading "+path, err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

type schemeGetResponse struct {
	Exists  bool   `json:"exists"`
	Version string `json:"version"`
}

// schemeGet runs the provider's get command: aux4 <get> --uri <uri> --to <localFile>.
// A missing remote file is exists:false with exit code 0 -- not an error.
func schemeGet(provider core.SchemeProvider, uri string, localFile string) (exists bool, version string, err error) {
	if strings.TrimSpace(provider.Get) == "" {
		return false, "", core.InternalError("Scheme provider for "+uri+" does not declare a 'get' command", nil)
	}

	instruction := fmt.Sprintf("aux4 %s --uri %s --to %s", provider.Get, shellQuote(uri), shellQuote(localFile))

	stdout, stderr, runErr := cmd.ExecuteCommandLineNoOutput(instruction)
	if runErr != nil {
		return false, "", schemeProviderError(uri, stderr, runErr)
	}

	var resp schemeGetResponse
	if parseErr := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &resp); parseErr != nil {
		return false, "", core.InternalError("Invalid response from scheme provider for "+uri, parseErr)
	}

	return resp.Exists, resp.Version, nil
}

// schemePut runs the provider's put command:
// aux4 <put> --uri <uri> --file <localFile> --ifMatch <etag|none>.
// Exit code 3 is a conflict (the remote changed since it was read).
func schemePut(provider core.SchemeProvider, uri string, localFile string, ifMatch string) error {
	if strings.TrimSpace(provider.Put) == "" {
		return core.InternalError("Scheme provider for "+uri+" does not declare a 'put' command", nil)
	}

	instruction := fmt.Sprintf("aux4 %s --uri %s --file %s --ifMatch %s", provider.Put, shellQuote(uri), shellQuote(localFile), shellQuote(ifMatch))

	_, stderr, runErr := cmd.ExecuteCommandLineNoOutput(instruction)
	if runErr == nil {
		return nil
	}

	if aux4Err, ok := runErr.(core.Aux4Error); ok && aux4Err.ExitCode == 3 {
		return core.SchemeConflictError(uri)
	}

	return schemeProviderError(uri, stderr, runErr)
}

func schemeProviderError(uri string, stderr string, err error) error {
	if aux4Err, ok := err.(core.Aux4Error); ok && aux4Err.ExitCode == 127 {
		return core.InternalError("Scheme provider command for "+uri+" was not found", nil)
	}
	if strings.TrimSpace(stderr) != "" {
		return core.InternalError(strings.TrimSpace(stderr), err)
	}
	return core.InternalError("Error resolving "+uri, err)
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

// splitFileExpression splits a file: executor expression ("<path>:<content>")
// into its path and content parts. A plain path is split at the first ":", as
// before. A path that is itself a literal "<scheme>://..." URI (not a ${var}
// that merely resolves to one -- that case has no literal colon to confuse)
// would otherwise be mis-split at the colon inside "scheme://", so in that
// case the split point is the first ":" found *after* the "://".
func splitFileExpression(expression string) (pathExpr string, contentExpr string, err error) {
	if prefixEnd := schemeURIPrefixEnd(expression); prefixEnd != -1 {
		if colon := strings.Index(expression[prefixEnd:], ":"); colon != -1 {
			splitAt := prefixEnd + colon
			return expression[:splitAt], expression[splitAt+1:], nil
		}
	}

	parts := strings.SplitN(expression, ":", 2)
	if len(parts) != 2 {
		return "", "", core.InternalError("file: requires format file:<path>:<content>", nil)
	}
	return parts[0], parts[1], nil
}

// schemeURIPrefixEnd returns the index right after "://" if expression starts
// with a "<scheme>://" URI (scheme made only of letters/digits/+/-/.), or -1
// otherwise. Only the very start of expression is checked -- the path is
// always first in a file: expression, so this never mistakes "://" appearing
// inside the content for a scheme.
func schemeURIPrefixEnd(expression string) int {
	schemeEnd := strings.Index(expression, "://")
	if schemeEnd <= 0 {
		return -1
	}
	for _, r := range expression[:schemeEnd] {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '+' || r == '-' || r == '.') {
			return -1
		}
	}
	return schemeEnd + len("://")
}

// writeSchemeFile implements the file: executor for a "<scheme>://" target:
// read the remote file's current ETag (if any), write/append the content to a
// local temp file, then CAS-put it back -- If-Match the read ETag, or
// If-None-Match ("none") when the remote file did not exist yet.
func writeSchemeFile(env *engine.VirtualEnvironment, scheme string, uri string, content string, isAppend bool) error {
	if env == nil || env.Schemes == nil {
		return core.InternalError(fmt.Sprintf("no package provides %s://; install aux4/%s", scheme, scheme), nil)
	}

	provider, found := env.Schemes.Get(scheme)
	if !found {
		return core.InternalError(fmt.Sprintf("no package provides %s://; install aux4/%s", scheme, scheme), nil)
	}

	tempFile, err := newSchemeTempFile(scheme)
	if err != nil {
		return err
	}
	defer os.Remove(tempFile)

	exists, version, err := schemeGet(provider, uri, tempFile)
	if err != nil {
		return err
	}

	if isAppend && exists {
		f, openErr := os.OpenFile(tempFile, os.O_APPEND|os.O_WRONLY, 0644)
		if openErr != nil {
			return core.InternalError("file: error opening file "+uri, openErr)
		}
		_, writeErr := f.WriteString(content + "\n")
		f.Close()
		if writeErr != nil {
			return core.InternalError("file: error appending to file "+uri, writeErr)
		}
	} else {
		if writeErr := os.WriteFile(tempFile, []byte(content+"\n"), 0644); writeErr != nil {
			return core.InternalError("file: error writing file "+uri, writeErr)
		}
	}

	ifMatch := "none"
	if exists {
		ifMatch = version
	}

	return schemePut(provider, uri, tempFile, ifMatch)
}
