package param

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"aux4.dev/aux4/core"
	"aux4.dev/aux4/io"
)

// coerceType converts a string value into a typed value according to a
// variable's declared "type". Unknown types or unparseable values fall back to
// the original string, so a bad declaration never loses data.
//
//	number  -> json.Number (keeps the exact numeric text, e.g. "12.50")
//	boolean -> bool
//	json    -> parsed JSON value (object, array, number, ...)
//	(other) -> unchanged string
func coerceType(value string, typeName string) any {
	switch typeName {
	case "number":
		trimmed := strings.TrimSpace(value)
		if _, err := strconv.ParseFloat(trimmed, 64); err == nil {
			return json.Number(trimmed)
		}
		return value
	case "boolean", "bool":
		switch strings.TrimSpace(value) {
		case "true":
			return true
		case "false":
			return false
		}
		return value
	case "json":
		decoder := json.NewDecoder(strings.NewReader(value))
		decoder.UseNumber()
		var parsed any
		if err := decoder.Decode(&parsed); err == nil {
			return parsed
		}
		return value
	default:
		return value
	}
}

// placeholder for $$ escape — chosen to be unlikely in real instructions
const dollarEscapePlaceholder = "\x00DOLLAR\x00"

func InjectParameters(command core.Command, instruction string, actions []string, params *Parameters) (string, error) {
	var err error

	// Protect $$ escape sequences from variable resolution
	instruction = strings.ReplaceAll(instruction, "$$", dollarEscapePlaceholder)

	// Phase 0: Resolve self-contained generators (date/time/uuid) BEFORE variable
	// expansion. These take no variables, so resolving them first means they only
	// ever see the authored execute string — never values expanded from ${...} or
	// command output — which keeps common tokens like uuid()/date(col) in
	// interpolated data or SQL from being clobbered. Authored occurrences can be
	// kept literal with a backslash (\uuid(), \date(...)).
	instruction, err = resolveDateTimeVariables(instruction)
	if err != nil {
		return "", err
	}

	instruction, err = resolveUUIDVariables(instruction)
	if err != nil {
		return "", err
	}

	// Phase 1: Resolve variable references ($var and ${var})
	instruction, err = resolveBareVariables(command, instruction, actions, params)
	if err != nil {
		return "", err
	}

	instruction, err = resolveBracedVariables(command, instruction, actions, params)
	if err != nil {
		return "", err
	}

	// Phase 2: Resolve conditionals before function-style resolvers.
	// This prevents if(...) patterns inside value()/values()/param()/params() results
	// from being misinterpreted as aux4 conditionals (e.g., JavaScript code containing
	// if(...) would otherwise be matched by the conditional regex).
	instruction, err = resolveConditional(command, instruction, actions, params)
	if err != nil {
		return "", err
	}

	instruction, err = resolveExists(command, instruction, actions, params)
	if err != nil {
		return "", err
	}

	// Phase 3: Resolve function-style parameter references.
	// These produce quoted/escaped output that should not be further processed.
	instruction, err = resolveValueVariables(command, instruction, actions, params)
	if err != nil {
		return "", err
	}

	instruction, err = resolveValuesVariables(command, instruction, actions, params)
	if err != nil {
		return "", err
	}

	instruction, err = resolveParamVariables(command, instruction, actions, params)
	if err != nil {
		return "", err
	}

	instruction, err = resolveParamsVariables(command, instruction, actions, params)
	if err != nil {
		return "", err
	}

	instruction, err = resolveObjectVariables(command, instruction, actions, params)
	if err != nil {
		return "", err
	}

	instruction, err = resolveNvlVariables(command, instruction, actions, params)
	if err != nil {
		return "", err
	}

	instruction, err = resolvePathVariables(command, instruction, actions, params)
	if err != nil {
		return "", err
	}

	instruction, err = resolveArgVariable(instruction, actions)
	if err != nil {
		return "", err
	}

	instruction, err = resolveArgsVariable(instruction, actions)
	if err != nil {
		return "", err
	}

	// Restore $$ escape sequences to literal $
	instruction = strings.ReplaceAll(instruction, dollarEscapePlaceholder, "$")

	instruction = strings.TrimSpace(instruction)
	instruction = regexp.MustCompile(`\s+`).ReplaceAllString(instruction, " ")

	return instruction, nil
}

func resolveValueVariables(command core.Command, instruction string, actions []string, params *Parameters) (string, error) {
	return resolveFunction(instruction, "value\\(([^)]+)\\)", func(groups []string) (string, error) {
		value, err := getVariableValueAsString(command, actions, params, groups[0], true)
		if err != nil && !strings.Contains(err.Error(), "Variable not found") {
			return "", err
		}
		return fmt.Sprintf("'%s'", value), nil
	})
}

func resolveValuesVariables(command core.Command, instruction string, actions []string, params *Parameters) (string, error) {
	return resolveFunction(instruction, "values\\(([^)]+)\\)", func(groups []string) (string, error) {
		expressionValue := ""
		variableList := strings.Split(groups[0], ",")
		for i := 0; i < len(variableList); i++ {
			variablePath := strings.TrimSpace(variableList[i])
			variableValue, err := getVariableValueAsString(command, actions, params, variablePath, true)
			if err != nil && !strings.Contains(err.Error(), "Variable not found") {
				return "", err
			}

			if i > 0 {
				expressionValue += " "
			}
			expressionValue += fmt.Sprintf("'%s'", variableValue)
		}
		return expressionValue, nil
	})
}

func resolveParamVariables(command core.Command, instruction string, actions []string, params *Parameters) (string, error) {
	return resolveFunction(instruction, "param\\(([^)]+)\\)", func(groups []string) (string, error) {
		return parseParam(command, actions, params, groups[0], true)
	})
}

func resolveParamsVariables(command core.Command, instruction string, actions []string, params *Parameters) (string, error) {
	return resolveFunction(instruction, "params\\(([^)]+)\\)", func(groups []string) (string, error) {
		expressionValue := ""
		variableList := strings.Split(groups[0], ",")
		for i := 0; i < len(variableList); i++ {
			variablePath := strings.TrimSpace(variableList[i])

			variableValue, err := parseParam(command, actions, params, variablePath, false)
			if err != nil {
				return "", err
			}

			if variableValue == "" {
				continue
			}

			if i > 0 {
				expressionValue += " "
			}
			expressionValue += variableValue
		}
		return expressionValue, nil
	})
}

func resolveObjectVariables(command core.Command, instruction string, actions []string, params *Parameters) (string, error) {
	return resolveFunction(instruction, "object\\(([^)]+)\\)", func(groups []string) (string, error) {
		return parseObject(command, actions, params, groups[0])
	})
}

func resolveNvlVariables(command core.Command, instruction string, actions []string, params *Parameters) (string, error) {
	return resolveFunction(instruction, "nvl\\(([^)]+)\\)", func(groups []string) (string, error) {
		candidates := strings.Split(groups[0], ",")
		result := ""
		for _, candidate := range candidates {
			candidate = strings.TrimSpace(candidate)
			if candidate == "" {
				continue
			}

			// Quoted string — strip quotes, use as literal
			if (strings.HasPrefix(candidate, "'") && strings.HasSuffix(candidate, "'")) ||
				(strings.HasPrefix(candidate, "\"") && strings.HasSuffix(candidate, "\"")) {
				result = candidate[1 : len(candidate)-1]
				break
			}

			// Unquoted numeric or boolean literal
			if isLiteral(candidate) {
				result = candidate
				break
			}

			// Variable lookup
			value, err := getVariableValueAsString(command, actions, params, candidate, false)
			if err == nil && value != "" {
				result = value
				break
			}
		}
		return result, nil
	})
}

// resolvePathVariables resolves path(seg1/seg2/...) into an absolute path.
// Each segment is joined with the OS separator and the result is made absolute
// (relative paths resolve against the current working directory). Segments are:
//   - a bare word         -> resolved as a variable value
//   - ".." or "."         -> kept as a literal path segment
//   - a quoted string     -> kept as a literal (quotes stripped)
// An empty result stays empty, matching the shell `case` idiom it replaces.
func resolvePathVariables(command core.Command, instruction string, actions []string, params *Parameters) (string, error) {
	return resolveFunction(instruction, "path\\(([^)]+)\\)", func(groups []string) (string, error) {
		parts := []string{}
		for _, segment := range splitPathSegments(groups[0]) {
			segment = strings.TrimSpace(segment)
			if segment == "" {
				continue
			}

			// "." and ".." are literal path segments, never variable names.
			if segment == "." || segment == ".." {
				parts = append(parts, segment)
				continue
			}

			// Quoted string — strip quotes, use as a literal segment.
			if (strings.HasPrefix(segment, "'") && strings.HasSuffix(segment, "'")) ||
				(strings.HasPrefix(segment, "\"") && strings.HasSuffix(segment, "\"")) {
				parts = append(parts, segment[1:len(segment)-1])
				continue
			}

			// Otherwise resolve it as a variable value.
			value, err := getVariableValueAsString(command, actions, params, segment, false)
			if err != nil {
				if strings.Contains(err.Error(), "Variable not found") {
					continue
				}
				return "", err
			}
			if value != "" {
				parts = append(parts, value)
			}
		}

		if len(parts) == 0 {
			return "", nil
		}

		absolute, err := filepath.Abs(filepath.Join(parts...))
		if err != nil {
			return "", err
		}
		return absolute, nil
	})
}

// splitPathSegments splits a path() argument on "/" while keeping quoted
// segments (which may themselves contain "/") intact.
func splitPathSegments(fieldList string) []string {
	segments := []string{}
	current := strings.Builder{}
	var quote rune

	for _, ch := range fieldList {
		if quote != 0 {
			current.WriteRune(ch)
			if ch == quote {
				quote = 0
			}
			continue
		}
		switch ch {
		case '\'', '"':
			quote = ch
			current.WriteRune(ch)
		case '/':
			segments = append(segments, current.String())
			current.Reset()
		default:
			current.WriteRune(ch)
		}
	}
	segments = append(segments, current.String())
	return segments
}

func isLiteral(candidate string) bool {
	if candidate == "true" || candidate == "false" || candidate == "null" {
		return true
	}
	if len(candidate) == 0 {
		return false
	}
	hasDigit := false
	hasDot := false
	for i, c := range candidate {
		if c == '-' && i == 0 {
			continue
		}
		if c == '.' && !hasDot {
			hasDot = true
			continue
		}
		if c >= '0' && c <= '9' {
			hasDigit = true
			continue
		}
		return false
	}
	return hasDigit
}

func parseObject(command core.Command, actions []string, params *Parameters, fieldList string) (string, error) {
	result := make(map[string]any)

	fields := strings.Split(fieldList, ",")
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}

		variablePath := field
		alias := ""
		if strings.Contains(field, ":") {
			parts := strings.SplitN(field, ":", 2)
			variablePath = strings.TrimSpace(parts[0])
			alias = strings.TrimSpace(parts[1])
		}

		// "*" collects every current parameter. Bare "*" spreads them into the
		// object itself; "*:key" nests them under that key as an object.
		if variablePath == "*" {
			all := make(map[string]any)
			for name, values := range params.params {
				if len(values) == 0 {
					continue
				}
				if len(values) == 1 {
					all[name] = values[0]
				} else {
					all[name] = values
				}
			}
			if alias == "" {
				for key, value := range all {
					result[key] = value
				}
			} else {
				result[alias] = all
			}
			continue
		}

		value, err := getVariableValue(command, actions, params, variablePath)
		if err != nil {
			if strings.Contains(err.Error(), "Variable not found") {
				continue
			}
			return "", err
		}

		// Skip empty string values (an unset variable), but keep any typed
		// value — a number, boolean or parsed JSON — so it stays typed.
		if strValue, ok := value.(string); ok && strValue == "" {
			continue
		}

		jsonKey := alias
		if jsonKey == "" {
			jsonKey = strings.ReplaceAll(variablePath, ".", "_")
		}
		result[jsonKey] = value
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func parseParam(command core.Command, actions []string, params *Parameters, variablePath string, allowAlias bool) (string, error) {
	var expressionValue string

	var actualVariablePath, alias string
	if allowAlias && strings.Contains(variablePath, ":") && !strings.HasSuffix(variablePath, "**") {
		parts := strings.SplitN(variablePath, ":", 2)
		actualVariablePath = strings.TrimSpace(parts[0])
		alias = strings.TrimSpace(parts[1])
	} else {
		actualVariablePath = variablePath
	}

	if strings.HasSuffix(actualVariablePath, "**") {
		result, err := params.Expr(command, actions, strings.TrimSuffix(actualVariablePath, "*"))
		if err != nil {
			return "", err
		}

		basePath := strings.TrimSuffix(actualVariablePath, "**")
		paramName := basePath
		if alias != "" {
			paramName = alias
		}

		if result != nil {
			typeOfResult := reflect.TypeOf(result)

			if typeOfResult.Kind() == reflect.Slice {
				for _, item := range result.([]any) {
					if expressionValue != "" {
						expressionValue += " "
					}
					expressionValue += fmt.Sprintf("--%s '%s'", paramName, item)
				}
			}
		}
	}

	if expressionValue == "" {
		result, err := getVariableValueAsString(command, actions, params, actualVariablePath, true)
		if err != nil {
			if strings.Contains(err.Error(), "Variable not found") {
				return "", nil
			}
			return "", err
		}

		if result != "" {
			var paramName string
			if alias != "" {
				paramName = alias
			} else {
				paramName = standardizeParameterName(actualVariablePath)
			}
			expressionValue = fmt.Sprintf("--%s '%s'", paramName, result)
		}
	}

	return expressionValue, nil
}

func resolveConditional(command core.Command, instruction string, actions []string, params *Parameters) (string, error) {
	return resolveFunction(instruction, "if\\(([^)]+)\\)", func(groups []string) (string, error) {
		variablePath := groups[0]
		conditionalSymbol := ""
		conditionalNot := ""
		expectedValue := ""

		if strings.Contains(variablePath, "==") {
			parts := strings.SplitN(variablePath, "==", 2)
			variablePath = strings.TrimSpace(parts[0])
			expectedValue = strings.TrimSpace(parts[1])
			conditionalSymbol = "="
		} else if strings.Contains(variablePath, "!=") {
			parts := strings.SplitN(variablePath, "!=", 2)
			variablePath = strings.TrimSpace(parts[0])
			expectedValue = strings.TrimSpace(parts[1])
			conditionalSymbol = "!="
		}

		if strings.HasPrefix(variablePath, "!") {
			conditionalNot = " ! "
			variablePath = strings.TrimSpace(variablePath[1:])
		}

		value, err := getVariableValueAsString(command, actions, params, variablePath, true)
		if err != nil && !strings.Contains(err.Error(), "Variable not found") {
			return "", err
		}

		if conditionalSymbol == "" {
			return fmt.Sprintf("[ %s'%s' ]", conditionalNot, value), nil
		}
		return fmt.Sprintf("[ %s'%s' %s '%s' ]", conditionalNot, value, conditionalSymbol, expectedValue), nil
	})
}

func resolveExists(command core.Command, instruction string, actions []string, params *Parameters) (string, error) {
	return resolveFunction(instruction, "exists\\(([^)]+)\\)", func(groups []string) (string, error) {
		variablePath := strings.TrimSpace(groups[0])
		value, err := getVariableValueAsString(command, actions, params, variablePath, true)
		if err != nil && !strings.Contains(err.Error(), "Variable not found") {
			return "", err
		}
		return fmt.Sprintf("[ -f '%s' ]", value), nil
	})
}

// resolveFunction resolves function-style calls of the form name(args) in the
// instruction. `inner` is the function-specific regex; its capture groups are
// handed to resolve() as `groups`. A leading backslash escapes a call — it is
// emitted literally (without the backslash) and resolve() is not invoked — which
// lets authored text keep tokens like \uuid() or \date(col) untouched (e.g. a
// database's own functions in an inline query). This is the single place the
// escape prefix and skip live, so individual resolvers only describe their own
// pattern and how to compute the value.
func resolveFunction(instruction, inner string, resolve func(groups []string) (string, error)) (string, error) {
	expr := regexp.MustCompile(`(\\)?` + inner)

	var resolveErr error
	result := expr.ReplaceAllStringFunc(instruction, func(match string) string {
		if resolveErr != nil {
			return match
		}

		groups := expr.FindStringSubmatch(match)
		if groups[1] != "" {
			return match[1:] // escaped — emit the literal without the backslash
		}

		value, err := resolve(groups[2:])
		if err != nil {
			resolveErr = err
			return match
		}
		return value
	})

	if resolveErr != nil {
		return "", resolveErr
	}
	return result, nil
}

func resolveBareVariables(command core.Command, instruction string, actions []string, params *Parameters) (string, error) {
	return resolveExpression(`\$[A-Za-z0-9]+`, command, instruction, actions, params)
}

func resolveBracedVariables(command core.Command, instruction string, actions []string, params *Parameters) (string, error) {
	return resolveExpression(`\$\{([^{}]+)\}`, command, instruction, actions, params)
}

func resolveExpression(expr string, command core.Command, instruction string, actions []string, params *Parameters) (string, error) {
	var innerErr error

	currentInstruction := instruction

	regex := regexp.MustCompile(expr)
	for {
		changed := false
		newInstruction := regex.ReplaceAllStringFunc(currentInstruction, func(fullMatch string) string {
			if innerErr != nil {
				return fullMatch
			}
			val, err := getVariableValueAsString(command, actions, params, fullMatch, false)
			if err != nil {
				if strings.Contains(err.Error(), "Variable not found") {
					return fullMatch
				}
				innerErr = err
				return fullMatch
			}
			changed = true
			return val
		})
		if innerErr != nil {
			return "", innerErr
		}
		if !changed {
			break
		}
		currentInstruction = newInstruction
	}

	return currentInstruction, nil
}

func getVariableValueAsString(command core.Command, actions []string, params *Parameters, variableName string, escape bool) (string, error) {
	if variableName == "*" {
		allVars := make(map[string]any)
		for name, values := range params.params {
			if len(values) > 0 {
				if len(values) == 1 {
					allVars[name] = values[0]
				} else {
					allVars[name] = values
				}
			}
		}
		return valueToString(allVars, escape), nil
	}

	value, err := getVariableValue(command, actions, params, variableName)
	if err != nil {
		return "", err
	}

	return valueToString(value, escape), nil
}

// getVariableValue resolves a variable to its raw (typed) value, applying secret
// resolution but without stringifying. Callers that build JSON (e.g. object())
// use this so a variable declared `type: number` stays a number.
func getVariableValue(command core.Command, actions []string, params *Parameters, variableName string) (any, error) {
	value, err := params.Expr(command, actions, variableName)
	if err != nil {
		return nil, err
	}

	if value == nil {
		return nil, core.VariableNotFoundError(variableName)
	}

	if strValue, ok := value.(string); ok && strings.HasPrefix(strValue, secretPrefix) {
		resolved, err := ResolveSingleSecret(strValue)
		if err != nil {
			return nil, err
		}
		value = resolved
	}

	return value, nil
}

func valueToString(value any, escape bool) string {
	if value == nil {
		return ""
	}

	if orderedMap, ok := value.(*io.OrderedMap); ok {
		jsonValue, err := orderedMap.MarshalJSON()
		if err != nil {
			value = fmt.Sprintf("%v", value)
		} else {
			value = string(jsonValue)
		}
	} else {
		typeOfValue := reflect.TypeOf(value)
		if typeOfValue.Kind() != reflect.String {
			jsonValue, err := marshalWithOrderPreservation(value)
			if err != nil {
				value = fmt.Sprintf("%v", value)
			} else {
				value = string(jsonValue)
			}
		}
	}

	if escape {
		value = escapeValue(value)
	}
	return fmt.Sprintf("%v", value)
}

func marshalWithOrderPreservation(value interface{}) ([]byte, error) {
	if orderedMap, ok := value.(*io.OrderedMap); ok {
		return orderedMap.MarshalJSON()
	}

	if mapValue, ok := value.(map[string]interface{}); ok {
		for key, val := range mapValue {
			if orderedMapVal, ok := val.(*io.OrderedMap); ok {
				mapValue[key] = orderedMapVal
			}
		}
	}

	if sliceValue, ok := value.([]interface{}); ok {
		for i, val := range sliceValue {
			if orderedMapVal, ok := val.(*io.OrderedMap); ok {
				sliceValue[i] = orderedMapVal
			}
		}
	}

	return json.Marshal(value)
}

func escapeValue(value any) string {
	if value == nil {
		return ""
	}
	return strings.ReplaceAll(value.(string), "'", "'\\''")
}

func resolveArgVariable(instruction string, actions []string) (string, error) {
	return resolveFunction(instruction, `arg\((\d+)\)`, func(groups []string) (string, error) {
		index := 0
		for _, c := range groups[0] {
			index = index*10 + int(c-'0')
		}
		if index < len(actions) {
			return actions[index], nil
		}
		return "", nil
	})
}

func resolveArgsVariable(instruction string, actions []string) (string, error) {
	return resolveFunction(instruction, `args\(([^)]+)\)`, func(groups []string) (string, error) {
		content := strings.TrimSpace(groups[0])

		if content == "*" {
			jsonBytes, err := json.Marshal(actions)
			if err != nil {
				return "", err
			}
			return string(jsonBytes), nil
		}

		indices := strings.Split(content, ",")
		result := []string{}
		for _, indexStr := range indices {
			indexStr = strings.TrimSpace(indexStr)
			index := 0
			for _, c := range indexStr {
				index = index*10 + int(c-'0')
			}
			if index < len(actions) {
				result = append(result, actions[index])
			} else {
				result = append(result, "")
			}
		}
		jsonBytes, err := json.Marshal(result)
		if err != nil {
			return "", err
		}
		return string(jsonBytes), nil
	})
}
