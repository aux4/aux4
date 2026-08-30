package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"aux4.dev/aux4/cmd"
	"aux4.dev/aux4/core"
	"aux4.dev/aux4/engine/param"
	"aux4.dev/aux4/engine/security"

	"github.com/manifoldco/promptui"
)

// resolveSecurityPolicy builds the command-exposure policy for this invocation.
//
// Precedence is env -> config -> param, first non-nil wins. This deliberately
// inverts aux4's normal "param overrides config": a policy, once imposed, must
// never be loosened by a flag, or the boundary is worthless.
//
//   - AUX4_SECURITY (env) is authoritative — inherited by any subprocess
//     shell-out and unreachable from a param, so it is the real boundary used
//     in the cloud.
//   - A config file is read DIRECTLY (via `aux4 config get --file`), not through
//     the normal parameter lookup, so param>config does not apply. It protects
//     the top-level call but does not propagate to subprocess shell-outs, so it
//     is for local convenience rather than a hard boundary.
//   - --security / --security.* is only consulted when neither of the above
//     defines a policy.
func resolveSecurityPolicy(params *param.Parameters) (*security.Policy, error) {
	if env := strings.TrimSpace(os.Getenv("AUX4_SECURITY")); env != "" {
		return security.Parse(env)
	}

	if raw := configSecurityValue(params); raw != nil {
		if policy, err := security.Parse(raw); err == nil && policy != nil {
			return policy, nil
		}
	}

	if raw := params.JustGet("security"); raw != nil {
		return security.Parse(raw)
	}

	return nil, nil
}

// configSecurityValue reads the `security` key straight out of the configured
// config file. It shells `aux4 config get --file` (config's own flag) rather
// than --configFile, so the child does not see a config selector and therefore
// resolves no policy of its own — no recursion, and the internal read is never
// itself governed.
func configSecurityValue(params *param.Parameters) any {
	configFile := configSelector(params, "configFile", "AUX4_CONFIG_FILE")
	config := configSelector(params, "config", "AUX4_CONFIG")
	if configFile == "" && config == "" {
		return nil
	}

	// Fast path: most config files carry no security policy at all, and the
	// `aux4 config get` subprocess below is comparatively expensive. When we have
	// a readable config file that does not even mention "security", skip it — no
	// policy can be there. (A missing/unreadable file, or the --config-only case,
	// falls through to the authoritative read.)
	if configFile != "" {
		if content, err := os.ReadFile(configFile); err == nil && !strings.Contains(string(content), "security") {
			return nil
		}
	}

	args := []string{}
	if configFile != "" {
		args = append(args, "--file "+configFile)
	}
	if config != "" && config != "true" {
		args = append(args, "--name "+config)
	}

	stdout, _, err := cmd.ExecuteCommandLineNoOutput(fmt.Sprintf("aux4 config get %s", strings.Join(args, " ")))
	if err != nil {
		return nil
	}

	trimmed := strings.TrimSpace(stdout)
	if trimmed == "" {
		return nil
	}

	var m map[string]interface{}
	if err := json.Unmarshal([]byte(trimmed), &m); err != nil {
		return nil
	}
	return m["security"]
}

func configSelector(params *param.Parameters, name, envVar string) string {
	if value := params.JustGet(name); value != nil {
		if s, ok := value.(string); ok {
			return s
		}
	}
	if value := os.Getenv(envVar); value != "" {
		return value
	}
	return ""
}

// enforceSecurity blocks a top-level invocation the policy does not allow.
//
// It is called once, at the process entry point, against the command path as
// typed. Profile routing and nested in-process `aux4` calls do not re-enter here,
// so a denied command stays callable internally by an allowed one. A denied
// command is reported as not found, matching the fact that it is also hidden.
// An ask command prompts; a declined or non-interactive prompt is a deny.
func enforceSecurity(policy *security.Policy, actions []string) error {
	if policy == nil || !policy.Active() || len(actions) == 0 {
		return nil
	}

	path := strings.Join(actions, " ")
	switch policy.Evaluate(path) {
	case security.Deny:
		return core.CommandNotFoundError(actions[0])
	case security.Ask:
		prompt := promptui.Prompt{
			Label:     fmt.Sprintf("Run restricted command 'aux4 %s'", path),
			IsConfirm: true,
		}
		if _, err := prompt.Run(); err != nil {
			return core.CommandNotFoundError(actions[0])
		}
	}
	return nil
}
