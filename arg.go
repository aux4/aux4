package main

import (
	"regexp"
	"strings"
)

func ParseArgs(args []string) ([]string, Parameters) {
	actions := make([]string, 0)
	params := make(map[string][]any)

	for index, arg := range args {
		if strings.HasPrefix(arg, "--") {
			name := arg[2:]
			value := ""
			if strings.Contains(name, "=") {
				parts := strings.Split(name, "=")
				name = parts[0]
				value = parts[1]
			} else if index+1 >= len(args) || strings.HasPrefix(args[index+1], "--") {
				value = "true"
			} else {
				value = args[index+1]
			}

			if params[name] == nil {
				params[name] = make([]any, 0)
			}
			params[name] = append(params[name], value)
		} else {
			actions = append(actions, arg)
		}
	}

	return actions, Parameters{params: params, lookups: ParameterLookups()}
}

type ParameterLookup interface {
  Get(parameters *Parameters, command *VirtualCommand, actions []string, name string) (any, error)
}

type Parameters struct {
	params map[string][]any
  lookups []ParameterLookup
}

func (p *Parameters) Set(name string, value any) {
  if p.params[name] == nil {
    p.params[name] = make([]any, 0)
  }
  p.params[name] = append(p.params[name], value)
}

func (p *Parameters) Update(name string, value any) {
  p.params[name] = make([]any, 0)
  p.params[name] = append(p.params[name], value)
}

func (p *Parameters) Has(name string) bool {
  return p.params[name] != nil
}

func (p *Parameters) Get(command *VirtualCommand, actions []string, name string) (any, error) {
	if p.params[name] != nil {
    return p.params[name][(len(p.params[name]) - 1)], nil
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

func (p *Parameters) GetMultiple(command *VirtualCommand, name string) []any {
	if p.params[name] == nil {
		return make([]any, 0)
	}
	return p.params[name]
}

func (p *Parameters) String() string {
  var builder strings.Builder
  for name, values := range p.params {
    for i, value := range values {
      if i > 0 {
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

func InjectParameters(command *VirtualCommand, instruction string, actions []string, params *Parameters) (string, error) {
	const variableRegex = "\\$\\{([a-zA-Z0-9_]+)\\}"
	expr := regexp.MustCompile(variableRegex)
  matches := expr.FindAllSubmatch([]byte(instruction), -1)

  variables := map[string]any{}
  for _, match := range matches {
    name := string(match[1])
    value, err := params.Get(command, actions, name)
    if err != nil {
      return "", err
    }

    variables[name] = value
  }

	return expr.ReplaceAllStringFunc(instruction, func(match string) string {
		name := match[2 : len(match)-1]
	  value := variables[name] 
    if value == nil {
      return match
    }
    return value.(string)
	}), nil
}
