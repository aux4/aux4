package param

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"aux4.dev/aux4/core"
	"aux4.dev/aux4/io"

	"github.com/yalp/jsonpath"
)

func ParseArgs(args []string) (Aux4Parameters, []string, Parameters) {
	actions := make([]string, 0)
	params := make(map[string][]any)
	aux4Params := make(map[string]string)

	for index := 0; index < len(args); index++ {
		arg := args[index]
		if strings.HasPrefix(arg, "--") {
			name := arg[2:]
			value := ""
			if strings.Contains(name, "=") {
				parts := strings.SplitN(name, "=", 2)
				name = parts[0]
				value = parts[1]
			} else if index+1 >= len(args) || strings.HasPrefix(args[index+1], "-") {
				value = "true"
			} else {
				value = args[index+1]
				index++
			}

			if strings.Contains(name, ".") {
				parts := strings.SplitN(name, ".", 2)
				baseName := parts[0]
				fieldPath := parts[1]

				var obj map[string]interface{}
				if params[baseName] != nil {
					last := params[baseName][len(params[baseName])-1]
					if m, ok := last.(map[string]interface{}); ok {
						obj = m
					}
				}
				if obj == nil {
					obj = make(map[string]interface{})
					if params[baseName] == nil {
						params[baseName] = make([]any, 0)
					}
					params[baseName] = append(params[baseName], obj)
				}
				setNestedField(obj, fieldPath, value)
			} else {
				if params[name] == nil {
					params[name] = make([]any, 0)
				}
				params[name] = append(params[name], value)
			}
		} else if strings.HasPrefix(arg, "-") {
			name := arg[1:]
			value := ""

			if strings.Contains(name, "=") {
				parts := strings.Split(name, "=")
				name = parts[0]
				value = parts[1]
			} else if index+1 >= len(args) || strings.HasPrefix(args[index+1], "-") {
				value = "true"
			} else {
				value = args[index+1]
				index++
			}

			aux4Params[name] = value
		} else {
			actions = append(actions, arg)
		}
	}

	return Aux4Parameters{params: aux4Params}, actions, Parameters{params: params, lookups: ParameterLookups()}
}

func ExtractArgs(instruction string) []string {
	var args []string
	var current strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	escapeNext := false
	for _, r := range instruction {
		if escapeNext {
			current.WriteRune(r)
			escapeNext = false
		} else if r == '\\' {
			escapeNext = true
		} else if r == '\'' && !inDoubleQuote {
			inSingleQuote = !inSingleQuote
		} else if r == '"' && !inSingleQuote {
			inDoubleQuote = !inDoubleQuote
		} else if unicode.IsSpace(r) && !inSingleQuote && !inDoubleQuote {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		} else {
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args
}

type Aux4Parameters struct {
	params map[string]string
}

func (params *Aux4Parameters) Local() bool {
	value, ok := params.params["local"]
	return ok && value == "true"
}

func (params *Aux4Parameters) NoPackages() bool {
	value, ok := params.params["noPackages"]
	return ok && value == "true"
}

func (params *Aux4Parameters) Has(name string) bool {
	_, ok := params.params[name]
	return ok
}

type Parameters struct {
	params  map[string][]any
	lookups []ParameterLookup
	// Names that must never resolve, whatever the source. Removing a value from params is
	// not enough on its own: a lookup would fetch it again from config or the environment.
	blocked map[string]bool
}

// ForHook returns a copy with every variable the command declares as hide or encrypt
// removed, and blocked from resolving again through config, env or any other lookup.
//
// A command's own execute steps still receive these values — they are what the command
// was given them for. A hook is an observer: it runs around someone else's command, so
// handing it credentials is a leak, and `secret://` cannot prevent it because secrets are
// resolved into params before hooks run. Values are dropped rather than masked, so nothing
// downstream can mistake a placeholder for a real credential.
//
// The copy is shallow by design: hook steps read the command's parameters, they do not
// write back into the command being hooked.
func (p *Parameters) ForHook(command core.Command) *Parameters {
	blocked := make(map[string]bool)
	if command.Help != nil {
		for _, variable := range command.Help.Variables {
			if variable != nil && (variable.Hide || variable.Encrypt) {
				blocked[variable.Name] = true
			}
		}
	}

	if len(blocked) == 0 {
		return p
	}

	params := make(map[string][]any, len(p.params))
	for name, values := range p.params {
		if blocked[name] {
			continue
		}
		params[name] = values
	}

	return &Parameters{params: params, lookups: p.lookups, blocked: blocked}
}

// Clone returns a snapshot with its own map, so later mutations of the original are not
// reflected in it. Values are shared; only the mapping is copied.
func (p *Parameters) Clone() *Parameters {
	params := make(map[string][]any, len(p.params))
	for name, values := range p.params {
		copied := make([]any, len(values))
		copy(copied, values)
		params[name] = copied
	}

	blocked := make(map[string]bool, len(p.blocked))
	for name := range p.blocked {
		blocked[name] = true
	}

	return &Parameters{params: params, lookups: p.lookups, blocked: blocked}
}

func (p *Parameters) Set(name string, value any) {
	if p.params[name] == nil {
		p.params[name] = make([]any, 0)
	}

	typeOfValue := reflect.TypeOf(value)
	if typeOfValue != nil && (typeOfValue.Kind() == reflect.Slice || typeOfValue.Kind() == reflect.Array) {
		p.params[name] = append(p.params[name], value.([]any)...)
	} else {
		p.params[name] = append(p.params[name], value)
	}
}

func (p *Parameters) Update(name string, value any) {
	p.params[name] = make([]any, 0)
	p.params[name] = append(p.params[name], value)
}

// UpdateField stores value under name, treating a dotted name as a field path
// into an object (e.g. "obj.name" sets the "name" field on the "obj" object).
// An existing object is preserved and merged into; otherwise a new one is
// created. Non-dotted names behave exactly like Update.
func (p *Parameters) UpdateField(name string, value any) {
	if !strings.Contains(name, ".") {
		p.Update(name, value)
		return
	}

	parts := strings.SplitN(name, ".", 2)
	baseName := parts[0]
	fieldPath := parts[1]

	obj, ok := p.JustGet(baseName).(map[string]interface{})
	if !ok {
		obj = make(map[string]interface{})
	}

	setNestedField(obj, fieldPath, value)
	p.Update(baseName, obj)
}

func (p *Parameters) Has(name string) bool {
	if p.blocked[name] {
		return false
	}
	return p.params[name] != nil
}

// IsEnabled reports whether a flag-style parameter is on. `--flag` alone is
// parsed as "true"; an explicit "false", "0", "no", "off" or an empty value
// turns it off.
func (p *Parameters) IsEnabled(name string) bool {
	value := p.JustGet(name)
	if value == nil {
		return false
	}

	switch strings.ToLower(strings.TrimSpace(fmt.Sprint(value))) {
	case "", "false", "0", "no", "off":
		return false
	}

	return true
}

func (p *Parameters) JustGet(name string) any {
	if p.blocked[name] {
		return nil
	}
	if p.params[name] != nil {
		return p.params[name][(len(p.params[name]) - 1)]
	}
	return nil
}

func (p *Parameters) Get(command core.Command, actions []string, name string) (any, error) {
	if p.blocked[name] {
		return nil, nil
	}
	if p.params[name] != nil {
		variable, exists := command.Help.GetVariable(name)
		if !exists || !variable.Encrypt {
			last := len(p.params[name]) - 1
			value := p.params[name][last]
			// Coerce a not-yet-typed value to the variable's declared type and
			// cache it back, so every later reader sees the typed value.
			if exists && variable.Type != "" {
				if strValue, ok := value.(string); ok {
					value = coerceType(strValue, variable.Type)
					p.params[name][last] = value
				}
			}
			return value, nil
		}
	}

	value := any(nil)

	for _, lookup := range p.lookups {
		result, err := lookup.Get(p, command, actions, name)
		if err != nil {
			return nil, err
		}

		if result != nil {
			if variable, exists := command.Help.GetVariable(name); exists && variable.Type != "" {
				if strValue, ok := result.(string); ok {
					result = coerceType(strValue, variable.Type)
				}
			}
			p.Set(name, result)
			value = result
			break
		}
	}

	return value, nil
}

func (p *Parameters) GetMultiple(command core.Command, actions []string, name string) ([]any, error) {
	if p.blocked[name] {
		return make([]any, 0), nil
	}
	if p.params[name] != nil {
		return p.params[name], nil
	}

	for _, lookup := range p.lookups {
		result, err := lookup.Get(p, command, actions, name)
		if err != nil {
			return nil, err
		}

		if result != nil {
			p.Set(name, result)
			return p.params[name], nil
		}
	}

	return make([]any, 0), nil
}

func (p *Parameters) Expr(command core.Command, actions []string, originalExpression string) (any, error) {
	var name string
	var value any
	index := -1
	var key string
	jsonExpr := ""

	var expression = strings.TrimSpace(originalExpression)

	if strings.HasPrefix(expression, "$") {
		expression = strings.TrimPrefix(expression, "$")
		expression = strings.TrimPrefix(expression, "{")
		expression = strings.TrimSuffix(expression, "}")
	}

	if !strings.Contains(expression, ".") && !strings.Contains(expression, "[") {
		name = expression
	} else {
		parts := strings.Split(expression, ".")
		name = parts[0]
		jsonExpr = strings.Join(parts[1:], ".")

		if strings.Contains(name, "[") {
			originalName := name
			openBracket := strings.Index(originalName, "[")
			closeBracket := strings.Index(originalName, "]")
			name = originalName[:openBracket]

			// A malformed reference (no closing "]", e.g. "${items[3}") falls
			// back to resolving the base name unindexed instead of panicking
			// on an invalid slice range.
			if closeBracket > openBracket {
				idx := originalName[openBracket+1 : closeBracket]
				if parsedIdx, err := strconv.Atoi(idx); err == nil {
					index = parsedIdx
				} else {
					key = idx
				}
			}
		}
	}

	multiple := false

	if strings.HasSuffix(name, "*") {
		name = strings.TrimSuffix(name, "*")
		multiple = true
	}

	if multiple {
		multiValue, err := p.GetMultiple(command, actions, name)
		if err != nil {
			return nil, err
		}

		value = multiValue
	} else {
		result, err := p.Get(command, actions, name)
		if err != nil {
			return nil, err
		}

		value = result
	}

	if value == nil {
		return nil, nil
	}

	if index != -1 {
		typeOfValue := reflect.TypeOf(value)
		indexable := typeOfValue != nil && (typeOfValue.Kind() == reflect.Slice || typeOfValue.Kind() == reflect.Array)

		if !indexable {
			multiValue, err := p.GetMultiple(command, actions, name)
			if err != nil {
				return nil, err
			}
			value = multiValue
			typeOfValue = reflect.TypeOf(value)
			indexable = typeOfValue != nil && (typeOfValue.Kind() == reflect.Slice || typeOfValue.Kind() == reflect.Array)
		}

		if indexable {
			if len(value.([]any)) > index {
				value = value.([]any)[index]
			} else {
				return nil, nil
			}
		} else {
			return nil, nil
		}
	} else if key != "" {
		if orderedMap, ok := value.(*io.OrderedMap); ok {
			if keyValue, exists := orderedMap.Get(key); exists {
				value = keyValue
			} else {
				return nil, nil
			}
		} else if valueMap, ok := value.(map[string]any); ok {
			if keyValue, exists := valueMap[key]; exists {
				value = keyValue
			} else {
				return nil, nil
			}
		} else {
			return nil, nil
		}
	}

	if jsonExpr != "" {
		if value == nil {
			return nil, nil
		}

		if strValue, ok := value.(string); ok {
			strValue = strings.TrimSpace(strValue)
			if (strings.HasPrefix(strValue, "{") && strings.HasSuffix(strValue, "}")) ||
				(strings.HasPrefix(strValue, "[") && strings.HasSuffix(strValue, "]")) {
				var parsed interface{}
				if err := json.Unmarshal([]byte(strValue), &parsed); err == nil {
					value = parsed
				}
			}
		}

		var jsonValue interface{}
		var err error

		if orderedMap, ok := value.(*io.OrderedMap); ok {
			jsonValue, err = navigateOrderedMap(orderedMap, jsonExpr)
		} else {
			jsonValue, err = jsonpath.Read(value, "$."+jsonExpr)
		}

		if err != nil {
			return nil, nil
		}

		value = jsonValue
	}

	return value, nil
}

func navigateOrderedMap(om *io.OrderedMap, path string) (interface{}, error) {
	parts := strings.Split(path, ".")
	current := interface{}(om)

	for _, part := range parts {
		if orderedMap, ok := current.(*io.OrderedMap); ok {
			value, exists := orderedMap.Get(part)
			if !exists {
				return nil, fmt.Errorf("key '%s' not found", part)
			}
			current = value
		} else if regularMap, ok := current.(map[string]interface{}); ok {
			value, exists := regularMap[part]
			if !exists {
				return nil, fmt.Errorf("key '%s' not found", part)
			}
			current = value
		} else {
			return nil, fmt.Errorf("cannot navigate into non-object type %T", current)
		}
	}

	return current, nil
}

func setNestedField(obj map[string]interface{}, path string, value interface{}) {
	parts := strings.SplitN(path, ".", 2)
	if len(parts) == 1 {
		obj[parts[0]] = value
		return
	}
	child, ok := obj[parts[0]].(map[string]interface{})
	if !ok {
		child = make(map[string]interface{})
		obj[parts[0]] = child
	}
	setNestedField(child, parts[1], value)
}

// reservedParameters are consumed by aux4 core itself and must never reach the
// command being executed — not as an argument, and not through the collectors
// that spread every parameter (`value(*)`, `object(*)`, alias forwarding).
var reservedParameters = map[string]bool{
	"prettify": true,
	// security is the command-exposure policy, consumed by core at the entry
	// point. It must never reach a command, or forward through value(*)/object(*),
	// so a package can neither read nor re-broadcast the policy it runs under.
	"security": true,
}

// injectedParameters are set by core onto the parameter set before a command runs
// (see MainExecute). They are not user input, so the unknown-parameter check must
// never flag them.
var injectedParameters = map[string]bool{
	"aux4HomeDir": true,
	"packageDir":  true,
	"configDir":   true,
}

// normalizeParameterName lowercases a name and drops every non-alphanumeric
// character, so customerId, customer-id, customer_id and CUSTOMER_ID all collapse
// to the same key. This is the comparison used to tell a real typo (a misspelling
// of a declared variable) apart from a genuinely different, undeclared parameter.
func normalizeParameterName(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// ValidateParameterNames errors when a passed parameter is the same as one the
// command declares once case and separators are ignored — the sign of a mistyped
// flag (--customer-id where the command declares customerId). Such a parameter was
// almost certainly meant to be the declared one, and silently accepting it as an
// undeclared extra means the declared variable never gets the value.
//
// A parameter that matches nothing declared is left alone: passing undeclared
// variables is a supported feature, so this only ever fires on the near-miss
// pattern, never on a genuinely new name. Core-injected and reserved parameters
// are skipped because they are not user input.
func ValidateParameterNames(command core.Command, params *Parameters) error {
	if command.Help == nil || len(command.Help.Variables) == 0 {
		return nil
	}

	declared := make(map[string]bool)
	normalized := make(map[string]string)
	for _, variable := range command.Help.Variables {
		if variable == nil || variable.Name == "" {
			continue
		}
		declared[variable.Name] = true
		normalized[normalizeParameterName(variable.Name)] = variable.Name
	}

	for name := range params.params {
		// Nested object fields (--obj.field) are stored under the base name already,
		// but guard against a dotted key defensively.
		base := name
		if i := strings.Index(base, "."); i >= 0 {
			base = base[:i]
		}

		if declared[base] || reservedParameters[base] || injectedParameters[base] || strings.HasPrefix(base, "__") {
			continue
		}

		if canonical, ok := normalized[normalizeParameterName(base)]; ok && canonical != base {
			return core.UnknownParameterError(base, canonical)
		}
	}

	return nil
}

// IsReservedParameter reports whether a parameter belongs to aux4 core.
func IsReservedParameter(name string) bool {
	return reservedParameters[name]
}

// UnexpectedParameters returns the names of parameters passed on the command line
// that this command does not declare, skipping core-reserved and core-injected names.
// It is the raw material for suggesting "you passed --x, did you mean --y?" when a
// required parameter turns out missing -- the flag wasn't silently dropped, it just
// didn't match anything declared, so it's worth naming back to the caller.
//
// Order is alphabetical so callers (and their error messages) are deterministic
// regardless of map iteration order or argv order.
func UnexpectedParameters(command core.Command, params *Parameters) []string {
	declared := make(map[string]bool)
	if command.Help != nil {
		for _, variable := range command.Help.Variables {
			if variable != nil && variable.Name != "" {
				declared[variable.Name] = true
			}
		}
	}

	seen := make(map[string]bool)
	unexpected := make([]string, 0)
	for name := range params.params {
		base := name
		if i := strings.Index(base, "."); i >= 0 {
			base = base[:i]
		}

		if declared[base] || reservedParameters[base] || injectedParameters[base] || strings.HasPrefix(base, "__") {
			continue
		}

		if !seen[base] {
			seen[base] = true
			unexpected = append(unexpected, base)
		}
	}

	sort.Strings(unexpected)
	return unexpected
}

// suggestUnexpectedParameter picks, among the flags the caller passed that aren't
// declared, the one most likely meant to be the missing parameter -- so the error can
// say "did you mean --messageId?" even for a name that doesn't look like a typo of it
// (e.g. --id vs --messageId), which ValidateParameterNames' near-miss check does not
// catch.
//
// Preference order:
//  1. A name that contains the other, case-insensitive (id ⊂ messageId) -- the
//     strongest signal that one is a shorthand for the other.
//  2. Otherwise, the candidate with the smallest edit distance to the missing name.
//  3. If there is exactly one unexpected flag, it is always the suggestion: a lone
//     extra flag next to a lone missing parameter is this call's only guess, even
//     when the names share nothing in common.
//
// Returns "" when there is nothing to suggest.
func suggestUnexpectedParameter(missingName string, unexpected []string) string {
	if len(unexpected) == 0 {
		return ""
	}

	if len(unexpected) == 1 {
		return unexpected[0]
	}

	lowerMissing := strings.ToLower(missingName)
	for _, candidate := range unexpected {
		lowerCandidate := strings.ToLower(candidate)
		if strings.Contains(lowerMissing, lowerCandidate) || strings.Contains(lowerCandidate, lowerMissing) {
			return candidate
		}
	}

	best := unexpected[0]
	bestDistance := levenshteinDistance(lowerMissing, strings.ToLower(best))
	for _, candidate := range unexpected[1:] {
		distance := levenshteinDistance(lowerMissing, strings.ToLower(candidate))
		if distance < bestDistance {
			bestDistance = distance
			best = candidate
		}
	}

	return best
}

// levenshteinDistance returns the classic edit distance between two strings: the
// minimum number of single-character insertions, deletions or substitutions needed
// to turn a into b. Used to rank unexpected-flag suggestions when no substring
// relationship exists between the candidate and the missing parameter name.
func levenshteinDistance(a, b string) int {
	ar := []rune(a)
	br := []rune(b)
	rows := len(ar) + 1
	cols := len(br) + 1

	dist := make([][]int, rows)
	for i := 0; i < rows; i++ {
		dist[i] = make([]int, cols)
		dist[i][0] = i
	}
	for j := 0; j < cols; j++ {
		dist[0][j] = j
	}

	for i := 1; i < rows; i++ {
		for j := 1; j < cols; j++ {
			cost := 1
			if ar[i-1] == br[j-1] {
				cost = 0
			}

			deletion := dist[i-1][j] + 1
			insertion := dist[i][j-1] + 1
			substitution := dist[i-1][j-1] + cost

			min := deletion
			if insertion < min {
				min = insertion
			}
			if substitution < min {
				min = substitution
			}
			dist[i][j] = min
		}
	}

	return dist[rows-1][cols-1]
}

func (p *Parameters) String() string {
	var builder strings.Builder
	for name, values := range p.params {
		if IsReservedParameter(name) {
			continue
		}

		for _, value := range values {
			if builder.Len() > 0 {
				builder.WriteString(" ")
			}

			builder.WriteString("--")
			builder.WriteString(name)
			builder.WriteString(" '")
			builder.WriteString(valueToString(value, false))
			builder.WriteString("'")
		}
	}
	return builder.String()
}

func standardizeParameterName(name string) string {
	if !strings.Contains(name, ".") {
		return removeStarSuffix(name)
	}

	var result strings.Builder
	upperNext := false

	for _, char := range name {
		if char == '.' {
			upperNext = true
			continue
		}

		if upperNext {
			result.WriteRune(unicode.ToUpper(char))
			upperNext = false
		} else {
			result.WriteRune(char)
		}
	}

	return removeStarSuffix(result.String())
}

func removeStarSuffix(name string) string {
	for strings.HasSuffix(name, "*") {
		name = strings.TrimSuffix(name, "*")
	}
	return name
}
