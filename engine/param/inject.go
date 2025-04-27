package param

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"aux4.dev/aux4/core"
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
		var expressionValue string

		variablePath := string(match[1])
		value, err := getVariableValueAsString(command, actions, params, variablePath, true)
		if err != nil {
			if strings.Contains(err.Error(), "Variable not found") {
				continue
			}
			return "", err
		}

		if value != "" {
			paramName := standardizeParameterName(variablePath)
			expressionValue = fmt.Sprintf("--%s '%s'", paramName, value)
			variables[variableExpression] = expressionValue
		}
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
			variableValue, err := getVariableValueAsString(command, actions, params, variablePath, true)
			if err != nil {
				if strings.Contains(err.Error(), "Variable not found") {
					continue
				}
				return "", err
			}

			if variableValue == "" {
				continue
			}

			if i > 0 {
				expressionValue += " "
			}

			paramName := standardizeParameterName(variablePath)
			expressionValue += fmt.Sprintf("--%s '%s'", paramName, variableValue)
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

	typeOfValue := reflect.TypeOf(value)
	if typeOfValue.Kind() != reflect.String {
		jsonValue, err := json.Marshal(value)
		if err != nil {
			value = fmt.Sprintf("%v", value)
		} else {
			value = string(jsonValue)
		}
	}

	if escape {
		value = escapeValue(value)
	}
	return fmt.Sprintf("%v", value)
}

func escapeValue(value any) string {
	if value == nil {
		return ""
	}
	return strings.ReplaceAll(value.(string), "'", "'\\''")
}
