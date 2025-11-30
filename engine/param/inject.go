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

func InjectParameters(command core.Command, instruction string, actions []string, params *Parameters) (string, error) {
	var err error

	instruction, err = resolveBareVariables(command, instruction, actions, params)
	if err != nil {
		return "", err
	}

	instruction, err = resolveBracedVariables(command, instruction, actions, params)
	if err != nil {
		return "", err
	}

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

	instruction, err = resolveConditional(command, instruction, actions, params)
	if err != nil {
		return "", err
	}

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

func parseParam(command core.Command, actions []string, params *Parameters, variablePath string, allowAlias bool) (string, error) {
	var expressionValue string

	var actualVariablePath, alias string
	if allowAlias && strings.Contains(variablePath, ",") && !strings.HasSuffix(variablePath, "**") {
		parts := strings.SplitN(variablePath, ",", 2)
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
