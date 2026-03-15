package param

import (
	"encoding/json"
	"fmt"
	"strings"

	"aux4.dev/aux4/cmd"
	"aux4.dev/aux4/core"
)

const secretPrefix = "secret://"

// parseSecretURI parses a secret URI like "secret://1password/Work/Jira/password"
// into provider ("1password"), ref ("Work/Jira"), and field ("password").
// Returns ok=false for non-secret values or malformed URIs (< 3 path segments).
func parseSecretURI(value string) (provider, ref, field string, ok bool) {
	if !strings.HasPrefix(value, secretPrefix) {
		return "", "", "", false
	}

	path := strings.TrimPrefix(value, secretPrefix)
	parts := strings.Split(path, "/")

	// Need at least 3 parts: provider, ref (1+), field
	if len(parts) < 3 {
		return "", "", "", false
	}

	provider = parts[0]
	field = parts[len(parts)-1]
	ref = strings.Join(parts[1:len(parts)-1], "/")

	return provider, ref, field, true
}

// secretTarget tracks where a resolved secret value should be placed back.
type secretTarget struct {
	paramName  string
	arrayIndex int
	field      string
}

// secretGroup collects fields and targets for a single provider+ref combination.
type secretGroup struct {
	provider string
	ref      string
	fields   []string
	targets  []secretTarget
}

// ResolveSecrets scans all parameter values for secret:// URIs, batches them
// by provider+ref, fetches each group with a single call to the secret provider,
// and replaces the values in-place.
func ResolveSecrets(params *Parameters) error {
	groups := map[string]*secretGroup{}

	for name, values := range params.params {
		for i, val := range values {
			strVal, ok := val.(string)
			if !ok {
				continue
			}

			provider, ref, field, ok := parseSecretURI(strVal)
			if !ok {
				continue
			}

			key := provider + ":" + ref
			g, exists := groups[key]
			if !exists {
				g = &secretGroup{provider: provider, ref: ref}
				groups[key] = g
			}

			// Only add field if not already present
			fieldExists := false
			for _, f := range g.fields {
				if f == field {
					fieldExists = true
					break
				}
			}
			if !fieldExists {
				g.fields = append(g.fields, field)
			}

			g.targets = append(g.targets, secretTarget{paramName: name, arrayIndex: i, field: field})
		}
	}

	if len(groups) == 0 {
		return nil
	}

	for _, g := range groups {
		resolved, err := fetchSecrets(g.provider, g.ref, g.fields)
		if err != nil {
			return err
		}

		for _, t := range g.targets {
			val, ok := resolved[t.field]
			if !ok {
				return core.InternalError(fmt.Sprintf("Secret field '%s' not found in response", t.field), nil)
			}
			params.params[t.paramName][t.arrayIndex] = val
		}
	}

	return nil
}

// ResolveSingleSecret resolves a single secret:// URI value.
// Returns the original value unchanged if it's not a secret URI.
func ResolveSingleSecret(value string) (string, error) {
	provider, ref, field, ok := parseSecretURI(value)
	if !ok {
		return value, nil
	}

	resolved, err := fetchSecrets(provider, ref, []string{field})
	if err != nil {
		return "", err
	}

	val, ok := resolved[field]
	if !ok {
		return "", core.InternalError(fmt.Sprintf("Secret field '%s' not found in response", field), nil)
	}

	return val, nil
}

// fetchSecrets calls the secret provider to resolve fields for a given ref.
// Executes: aux4 secret <provider> get --ref "<ref>" --fields "<f1>,<f2>"
func fetchSecrets(provider, ref string, fields []string) (map[string]string, error) {
	instruction := fmt.Sprintf("aux4 secret %s get --ref '%s' --fields '%s'",
		provider, ref, strings.Join(fields, ","))

	stdout, stderr, err := cmd.ExecuteCommandLineNoOutput(instruction)
	if err != nil {
		if aux4Err, ok := err.(core.Aux4Error); ok && aux4Err.ExitCode == 127 {
			return nil, core.InternalError(
				fmt.Sprintf("Secret provider 'aux4/secret-%s' is not installed. Install it with: aux4 aux4 pkger install aux4/secret-%s", provider, provider),
				nil,
			)
		}
		if strings.TrimSpace(stderr) != "" {
			return nil, core.InternalError(strings.TrimSpace(stderr), err)
		}
		return nil, err
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &result); err != nil {
		return nil, core.InternalError(
			fmt.Sprintf("Invalid JSON response from secret provider '%s'", provider),
			err,
		)
	}

	return result, nil
}
