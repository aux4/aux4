package param

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"aux4.dev/aux4/core"
	"aux4.dev/aux4/io"
)

// placeholder for $$ escape — chosen to be unlikely in real instructions
const dollarEscapePlaceholder = "\x00DOLLAR\x00"

func InjectParameters(command core.Command, instruction string, actions []string, params *Parameters) (string, error) {
	var err error

	// Protect $$ escape sequences from variable resolution
	instruction = strings.ReplaceAll(instruction, "$$", dollarEscapePlaceholder)

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
	const variableRegex = "value\\(([^)]+)\\)"

	expr := regexp.MustCompile(variableRegex)
	matches := expr.FindAllSubmatch([]byte(instruction), -1)

	variables := map[string]any{}
	for _, match := range matches {
		variableExpression := string(match[0])

		variablePath := string(match[1])
		value, err := getVariableValueAsString(command, actions, params, variablePath, true)
		if err != nil {
			if !strings.Contains(err.Error(), "Variable not found") {
				return "", err
			}
		}

		variables[variableExpression] = fmt.Sprintf("'%s'", value)
	}

	return replaceVariables(expr, instruction, variables), nil
}

func resolveValuesVariables(command core.Command, instruction string, actions []string, params *Parameters) (string, error) {
	const variableRegex = "values\\(([^)]+)\\)"

	expr := regexp.MustCompile(variableRegex)
	matches := expr.FindAllSubmatch([]byte(instruction), -1)

	variables := map[string]any{}
	for _, match := range matches {
		variableExpression := string(match[0])
		expressionValue := ""

		variableList := strings.Split(string(match[1]), ",")
		for i := 0; i < len(variableList); i++ {
			variablePath := strings.TrimSpace(variableList[i])
			variableValue, err := getVariableValueAsString(command, actions, params, variablePath, true)
			if err != nil {
				if !strings.Contains(err.Error(), "Variable not found") {
					return "", err
				}
			}

			if i > 0 {
				expressionValue += " "
			}

			expressionValue += fmt.Sprintf("'%s'", variableValue)
		}

		variables[variableExpression] = expressionValue
	}

	return replaceVariables(expr, instruction, variables), nil
}

func resolveParamVariables(command core.Command, instruction string, actions []string, params *Parameters) (string, error) {
	const variableRegex = "param\\(([^)]+)\\)"

	expr := regexp.MustCompile(variableRegex)
	matches := expr.FindAllSubmatch([]byte(instruction), -1)

	variables := map[string]any{}
	for _, match := range matches {
		variableExpression := string(match[0])
		variablePath := string(match[1])

		expressionValue, err := parseParam(command, actions, params, variablePath, true)
		if err != nil {
			return "", err
		}

		variables[variableExpression] = expressionValue
	}

	return replaceVariables(expr, instruction, variables), nil
}

func resolveParamsVariables(command core.Command, instruction string, actions []string, params *Parameters) (string, error) {
	const variableRegex = "params\\(([^)]+)\\)"

	expr := regexp.MustCompile(variableRegex)
	matches := expr.FindAllSubmatch([]byte(instruction), -1)

	variables := map[string]any{}
	for _, match := range matches {
		variableExpression := string(match[0])
		var expressionValue string

		expressionValue = ""

		variableList := strings.Split(string(match[1]), ",")
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

		variables[variableExpression] = expressionValue
	}

	return replaceVariables(expr, instruction, variables), nil
}

func resolveObjectVariables(command core.Command, instruction string, actions []string, params *Parameters) (string, error) {
	const variableRegex = "object\\(([^)]+)\\)"

	expr := regexp.MustCompile(variableRegex)
	matches := expr.FindAllSubmatch([]byte(instruction), -1)

	variables := map[string]any{}
	for _, match := range matches {
		variableExpression := string(match[0])

		expressionValue, err := parseObject(command, actions, params, string(match[1]))
		if err != nil {
			return "", err
		}

		variables[variableExpression] = expressionValue
	}

	return replaceVariables(expr, instruction, variables), nil
}

func resolveNvlVariables(command core.Command, instruction string, actions []string, params *Parameters) (string, error) {
	const variableRegex = "nvl\\(([^)]+)\\)"

	expr := regexp.MustCompile(variableRegex)
	matches := expr.FindAllSubmatch([]byte(instruction), -1)

	variables := map[string]any{}
	for _, match := range matches {
		variableExpression := string(match[0])
		candidates := strings.Split(string(match[1]), ",")

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

		variables[variableExpression] = result
	}

	return replaceVariables(expr, instruction, variables), nil
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
	result := make(map[string]string)

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

		value, err := getVariableValueAsString(command, actions, params, variablePath, false)
		if err != nil {
			if strings.Contains(err.Error(), "Variable not found") {
				continue
			}
			return "", err
		}

		if value != "" {
			jsonKey := alias
			if jsonKey == "" {
				jsonKey = strings.ReplaceAll(variablePath, ".", "_")
			}
			result[jsonKey] = value
		}
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
	const variableRegex = "if\\(([^)]+)\\)"

	expr := regexp.MustCompile(variableRegex)
	matches := expr.FindAllSubmatch([]byte(instruction), -1)

	variables := map[string]any{}
	for _, match := range matches {
		variableExpression := string(match[0])

		variablePath := string(match[1])
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

		var expressionValue string

		value, err := getVariableValueAsString(command, actions, params, variablePath, true)
		if err != nil {
			if !strings.Contains(err.Error(), "Variable not found") {
				return "", err
			}
		}

		if conditionalSymbol == "" {
			expressionValue = fmt.Sprintf("[ %s'%s' ]", conditionalNot, value)
		} else {
			expressionValue = fmt.Sprintf("[ %s'%s' %s '%s' ]", conditionalNot, value, conditionalSymbol, expectedValue)
		}

		variables[variableExpression] = expressionValue
	}

	return replaceVariables(expr, instruction, variables), nil
}

func resolveExists(command core.Command, instruction string, actions []string, params *Parameters) (string, error) {
	const variableRegex = "exists\\(([^)]+)\\)"

	expr := regexp.MustCompile(variableRegex)
	matches := expr.FindAllSubmatch([]byte(instruction), -1)

	variables := map[string]any{}
	for _, match := range matches {
		variableExpression := string(match[0])
		variablePath := strings.TrimSpace(string(match[1]))

		value, err := getVariableValueAsString(command, actions, params, variablePath, true)
		if err != nil {
			if !strings.Contains(err.Error(), "Variable not found") {
				return "", err
			}
		}

		variables[variableExpression] = fmt.Sprintf("[ -f '%s' ]", value)
	}

	return replaceVariables(expr, instruction, variables), nil
}

func replaceVariables(expr *regexp.Regexp, instruction string, variables map[string]any) string {
	return expr.ReplaceAllStringFunc(instruction, func(match string) string {
		value, exists := variables[match]
		if !exists {
			return match
		}

		return fmt.Sprintf("%v", value)
	})
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

	value, err := params.Expr(command, actions, variableName)
	if err != nil {
		return "", err
	}

	if value == nil {
		return "", core.VariableNotFoundError(variableName)
	}

	if strValue, ok := value.(string); ok && strings.HasPrefix(strValue, secretPrefix) {
		resolved, err := ResolveSingleSecret(strValue)
		if err != nil {
			return "", err
		}
		value = resolved
	}

	return valueToString(value, escape), nil
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
	const variableRegex = `arg\((\d+)\)`

	expr := regexp.MustCompile(variableRegex)
	matches := expr.FindAllSubmatch([]byte(instruction), -1)

	variables := map[string]any{}
	for _, match := range matches {
		variableExpression := string(match[0])
		indexStr := string(match[1])

		index := 0
		for _, c := range indexStr {
			index = index*10 + int(c-'0')
		}

		if index < len(actions) {
			variables[variableExpression] = actions[index]
		} else {
			variables[variableExpression] = ""
		}
	}

	return replaceVariables(expr, instruction, variables), nil
}

func resolveArgsVariable(instruction string, actions []string) (string, error) {
	const variableRegex = `args\(([^)]+)\)`

	expr := regexp.MustCompile(variableRegex)
	matches := expr.FindAllSubmatch([]byte(instruction), -1)

	variables := map[string]any{}
	for _, match := range matches {
		variableExpression := string(match[0])
		content := strings.TrimSpace(string(match[1]))

		if content == "*" {
			jsonBytes, err := json.Marshal(actions)
			if err != nil {
				return "", err
			}
			variables[variableExpression] = string(jsonBytes)
		} else {
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
			variables[variableExpression] = string(jsonBytes)
		}
	}

	return replaceVariables(expr, instruction, variables), nil
}
