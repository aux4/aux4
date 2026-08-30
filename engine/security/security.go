// Package security implements aux4 core's command-exposure policy: a runtime
// decision, imposed by whoever runs the CLI, about which commands may be invoked
// directly and which are listed. It is intentionally self-contained (no aux4
// engine imports) so it can be stored on the VirtualEnvironment and consulted
// from the entry points, the executor and the help/autocomplete renderers
// without creating an import cycle.
//
// The model has two axes:
//
//   - Callability — allow / ask / deny, matched against the command path as
//     typed after `aux4`, space-joined and without the `aux4` prefix
//     (e.g. "db query", "ai agent config get").
//   - Visibility — a denied command is not listed. This follows from
//     callability; there is no separate hide list.
//
// deny blocks DIRECT invocation only. A command that runs another command
// internally (its execute steps call `aux4 other`) is never evaluated here, so a
// denied command remains a usable building block for an exposed one.
package security

import (
	"encoding/json"
	"regexp"
	"strings"
	"sync"
)

// Decision is the outcome of evaluating a command path against a policy.
type Decision int

const (
	// Allow — the command may be invoked directly.
	Allow Decision = iota
	// Ask — the command may be invoked after interactive confirmation. In a
	// non-interactive context the caller must treat this as Deny (fail closed).
	Ask
	// Deny — the command is not directly callable and is not listed.
	Deny
)

// Policy is a parsed, immutable set of allow/ask/deny glob patterns. The zero
// value (and nil) is an inactive policy that allows everything, so a CLI with no
// policy configured behaves exactly as before.
type Policy struct {
	allow []string
	ask   []string
	deny  []string

	mu    sync.Mutex
	cache map[string]*regexp.Regexp
}

type policyJSON struct {
	Allow []string `json:"allow"`
	Ask   []string `json:"ask"`
	Deny  []string `json:"deny"`
}

// Parse builds a Policy from a value that is either a JSON string
// ({"allow":[...],"deny":[...],"ask":[...]}) or an already-decoded
// map[string]interface{} / map with string-slice fields (as produced by aux4's
// dotted-parameter and config parsing). A nil or empty value yields a nil
// policy (inactive). Unknown shapes are ignored rather than erroring so a
// malformed policy never silently opens the CLI — an empty policy stays inactive.
func Parse(value any) (*Policy, error) {
	switch v := value.(type) {
	case nil:
		return nil, nil
	case string:
		s := strings.TrimSpace(v)
		if s == "" || s == "true" {
			return nil, nil
		}
		return ParseJSON(s)
	case *Policy:
		return v, nil
	case map[string]interface{}:
		return fromMap(v), nil
	default:
		// Round-trip anything else through JSON — covers ordered maps and other
		// structured decodes without depending on their concrete type.
		b, err := json.Marshal(v)
		if err != nil {
			return nil, nil
		}
		return ParseJSON(string(b))
	}
}

// ParseJSON builds a Policy from a JSON object string.
func ParseJSON(s string) (*Policy, error) {
	var pj policyJSON
	if err := json.Unmarshal([]byte(s), &pj); err != nil {
		return nil, err
	}
	p := &Policy{allow: clean(pj.Allow), ask: clean(pj.Ask), deny: clean(pj.Deny)}
	if !p.Active() {
		return nil, nil
	}
	return p, nil
}

func fromMap(m map[string]interface{}) *Policy {
	p := &Policy{
		allow: toStrings(m["allow"]),
		ask:   toStrings(m["ask"]),
		deny:  toStrings(m["deny"]),
	}
	if !p.Active() {
		return nil
	}
	return p
}

func toStrings(v any) []string {
	switch t := v.(type) {
	case []string:
		return clean(t)
	case []interface{}:
		out := make([]string, 0, len(t))
		for _, e := range t {
			if s, ok := e.(string); ok {
				out = append(out, s)
			}
		}
		return clean(out)
	case string:
		return clean([]string{t})
	}
	return nil
}

func clean(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// Active reports whether the policy constrains anything. An inactive policy
// (nil, or with no patterns) allows and shows everything.
func (p *Policy) Active() bool {
	if p == nil {
		return false
	}
	return len(p.allow) > 0 || len(p.ask) > 0 || len(p.deny) > 0
}

// Evaluate returns the decision for invoking the exact command at path (the
// space-joined command path, e.g. "db query").
//
// Rules:
//   - An inactive policy allows everything.
//   - An empty allow list means "allow by default" (allow = ["*"]); a non-empty
//     allow list is an explicit allow-list and anything unmatched is denied.
//   - The most specific matching pattern wins; on a specificity tie deny beats
//     ask beats allow (fail closed).
//   - A path that matches nothing but is an ancestor of an allow pattern (e.g.
//     "db" when "db *" is allowed) is allowed so routers stay navigable —
//     unless a deny pattern explicitly matched it.
func (p *Policy) Evaluate(path string) Decision {
	if !p.Active() {
		return Allow
	}

	path = strings.TrimSpace(path)

	allow := p.allow
	if len(allow) == 0 {
		allow = []string{"*"}
	}

	bestScore := -1
	best := Deny

	consider := func(patterns []string, d Decision) {
		for _, pat := range patterns {
			if !match(p, pat, path) {
				continue
			}
			s := specificity(pat)
			if s > bestScore || (s == bestScore && priority(d) > priority(best)) {
				bestScore = s
				best = d
			}
		}
	}

	consider(allow, Allow)
	consider(p.ask, Ask)
	consider(p.deny, Deny)

	if bestScore >= 0 && best != Deny {
		return best
	}

	// The direct decision is Deny (or nothing matched). Keep a router navigable
	// toward an allowed descendant, unless a deny pattern covering this path is at
	// least as specific as the allow that would open the subtree.
	if p.navigable(path, allow) {
		return Allow
	}
	return Deny
}

// Visible reports whether the command at path should appear in help and
// autocomplete. Everything not denied is visible (an ask command is still shown).
func (p *Policy) Visible(path string) bool {
	return p.Evaluate(path) != Deny
}

// navigable reports whether path should stay reachable/visible as a router
// because some allow pattern at or under it opens an allowed descendant, and no
// deny pattern covering path is specific enough to shut that subtree.
//
// It compares the most specific allow pattern that sits at or under path against
// the most specific deny pattern that matches path itself (a deny on the
// ancestor covers the whole subtree). A broad default-deny like "*" therefore
// never hides a router that leads to an explicitly allowed command, while an
// explicit deny of the router (or a deny more specific than the allow) does.
func (p *Policy) navigable(path string, allow []string) bool {
	denyScore := -1
	for _, pat := range p.deny {
		if match(p, pat, path) {
			if s := specificity(pat); s > denyScore {
				denyScore = s
			}
		}
	}

	prefix := path + " "
	for _, pat := range allow {
		atOrUnder := pat == path || (path != "" && strings.HasPrefix(pat, prefix)) || (path == "" && pat != "*")
		if !atOrUnder {
			continue
		}
		if specificity(pat) > denyScore {
			return true
		}
	}
	return false
}

// priority orders decisions for tie-breaking: deny > ask > allow.
func priority(d Decision) int {
	switch d {
	case Deny:
		return 2
	case Ask:
		return 1
	default:
		return 0
	}
}

// specificity scores how tightly a pattern is anchored. An exact pattern (no
// wildcard) always beats a wildcard one; among wildcard patterns, more literal
// characters win. "*" is the least specific.
func specificity(pattern string) int {
	if pattern == "*" {
		return 0
	}
	score := len(strings.ReplaceAll(pattern, "*", ""))
	if !strings.Contains(pattern, "*") {
		score += 1000
	}
	return score
}

// match reports whether a glob pattern (where * matches any run of characters,
// spaces included) matches the whole path.
func match(p *Policy, pattern, path string) bool {
	re := p.compile(pattern)
	return re.MatchString(path)
}

func (p *Policy) compile(pattern string) *regexp.Regexp {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.cache == nil {
		p.cache = make(map[string]*regexp.Regexp)
	}
	if re, ok := p.cache[pattern]; ok {
		return re
	}
	var b strings.Builder
	b.WriteString("^")
	for _, r := range pattern {
		if r == '*' {
			b.WriteString(".*")
		} else {
			b.WriteString(regexp.QuoteMeta(string(r)))
		}
	}
	b.WriteString("$")
	re := regexp.MustCompile(b.String())
	p.cache[pattern] = re
	return re
}
