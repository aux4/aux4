package param

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"aux4.dev/aux4/core"

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
				parts := strings.Split(name, "=")
				name = parts[0]
				value = parts[1]
			} else if index+1 >= len(args) || strings.HasPrefix(args[index+1], "-") {
				value = "true"
			} else {
				value = args[index+1]
				index++
			}

			if params[name] == nil {
				params[name] = make([]any, 0)
			}
			params[name] = append(params[name], value)
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

type Aux4Parameters struct {
	params map[string]string
}

func (params *Aux4Parameters) Local() bool {
	value, ok := params.params["local"]
	return ok && value == "true"
}

func (params *Aux4Parameters) NoPackages() bool {
	value, ok := params.params["no-packages"]
	return ok && value == "true"
}

type Parameters struct {
	params  map[string][]any
	lookups []ParameterLookup
}

func (p *Parameters) Set(name string, value any) {
	if p.params[name] == nil {
		p.params[name] = make([]any, 0)
	}

	typeOfValue := reflect.TypeOf(value)
	if typeOfValue.Kind() == reflect.Slice || typeOfValue.Kind() == reflect.Array {
		p.params[name] = append(p.params[name], value.([]any)...)
	} else {
		p.params[name] = append(p.params[name], value)
	}
}

func (p *Parameters) Update(name string, value any) {
	p.params[name] = make([]any, 0)
	p.params[name] = append(p.params[name], value)
}

func (p *Parameters) Has(name string) bool {
	return p.params[name] != nil
}

func (p *Parameters) JustGet(name string) any {
	if p.params[name] != nil {
		return p.params[name][(len(p.params[name]) - 1)]
	}
	return nil
}

func (p *Parameters) Get(command core.Command, actions []string, name string) (any, error) {
	if p.params[name] != nil {
		variable, exists := command.Help.GetVariable(name)
		if !exists || !variable.Encrypt {
			return p.params[name][(len(p.params[name]) - 1)], nil
		}
	}

	value := any(nil)

	for _, lookup := range p.lookups {
		result, err := lookup.Get(p, command, actions, name)
		if err != nil {
			return nil, err
		}

		if result != nil {
			p.Set(name, result)
			value = result
			break
		}
	}

	return value, nil
}

func (p *Parameters) GetMultiple(command core.Command, actions []string, name string) ([]any, error) {
	if p.params[name] == nil {
		return make([]any, 0), nil
	}
	return p.params[name], nil
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
			name = name[:strings.Index(name, "[")]
			idx := originalName[strings.Index(originalName, "[")+1 : strings.Index(originalName, "]")]
			if parsedIdx, err := strconv.Atoi(idx); err == nil {
				index = parsedIdx
			} else {
				key = idx
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

	if index != -1 {
		typeOfValue := reflect.TypeOf(value)
		if typeOfValue.Kind() == reflect.Slice || typeOfValue.Kind() == reflect.Array {
			if len(value.([]any)) > index {
				value = value.([]any)[index]
			} else {
				return nil, core.InternalError("Index out of range: "+expression, nil)
			}
		} else {
			return nil, core.InternalError("Index out of range: "+expression, nil)
		}
	} else if key != "" {
		if valueMap, ok := value.(map[string]any); ok {
			if keyValue, exists := valueMap[key]; exists {
				value = keyValue
			} else {
				return nil, core.InternalError("Key not found: "+expression, nil)
			}
		} else {
			return nil, core.InternalError("Cannot apply key lookup: "+expression, nil)
		}
	}

	if jsonExpr != "" {
		jsonValue, err := jsonpath.Read(value, "$."+jsonExpr)
		if err != nil {
			return nil, core.InternalError("Error trying to access field '"+jsonExpr+"' from "+fmt.Sprintf("'%v'", value), err)
		}
		value = jsonValue
	}

	return value, nil
}

func (p *Parameters) String() string {
	var builder strings.Builder
	for name, values := range p.params {
		for _, value := range values {
			if builder.Len() > 0 {
				builder.WriteString(" ")
			}

			builder.WriteString("--")
			builder.WriteString(name)
			builder.WriteString(" '")
			builder.WriteString(value.(string))
			builder.WriteString("'")
		}
	}
	return builder.String()
}

func standardizeParameterName(name string) string {
	if !strings.Contains(name, ".") {
		return name
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

	return result.String()
}
