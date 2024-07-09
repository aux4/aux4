package man

import (
  "aux4/core"
  "aux4/output"
  "fmt"
  "regexp"
  "strconv"
  "strings"
)

func ShowCommandSource(command core.Command) {
	lineNumbers := len(command.Execute)
	digits := len(strconv.Itoa(lineNumbers))

	for line, commandLine := range command.Execute {
		lineNumber := fmt.Sprintf("%*d", digits, line+1)
		output.Out(output.StdOut).Println(output.Gray(lineNumber), formatCommandLine(commandLine))
	}
}

func formatCommandLine(commandLine string) string {
	formatted := commandLine

	executorPrefixRegex := regexp.MustCompile(`^(\S*):(.+)`)
	executorPrefixMatch := executorPrefixRegex.FindStringSubmatch(formatted)
	if len(executorPrefixMatch) > 0 {
		prefix := executorPrefixMatch[1]
		command := executorPrefixMatch[2]

		formatted = output.Magenta(prefix, ":")

		if prefix == "log" || prefix == "debug" || prefix == "confirm" {
			formatted += output.Green(command)
		} else if prefix == "set" {
			declarations := strings.Split(command, ";")
      for i, declaration := range declarations {
        if i > 0 {
          formatted += output.Gray(";")
        }

        parts := strings.Split(declaration, "=")
        variable := parts[0]
        value := parts[1]

        formatted += output.Cyan(variable) + "="

        if strings.HasPrefix(value, "!") {
          formatted += value
        } else {
          formatted += output.Green(value)
        }
      }
		} else {
			formatted += command
		}
	}

	textRegex := regexp.MustCompile(`"(.*?)"|'(.*?)'`)
	textMatches := textRegex.FindAllStringSubmatch(formatted, -1)
	for _, textMatch := range textMatches {
		formatted = strings.ReplaceAll(formatted, textMatch[0], output.Green(textMatch[0]))
	}

	filePathRegex := regexp.MustCompile(`\/?(\S*)\/(\S*)`)
	filePathMatches := filePathRegex.FindAllStringSubmatch(formatted, -1)
	for _, filePathMatch := range filePathMatches {
		formatted = strings.ReplaceAll(formatted, filePathMatch[0], output.Yellow(filePathMatch[0]))
	}

	variableRegex := regexp.MustCompile("\\$([a-zA-Z0-9]+)|\\$\\{([^}\\s]+)\\}")
	variableMatches := variableRegex.FindAllStringSubmatch(formatted, -1)
	for _, variableMatch := range variableMatches {
		formatted = strings.ReplaceAll(formatted, variableMatch[0], output.Cyan(variableMatch[0]))
	}

  parameterRegex := regexp.MustCompile(`\s-{1,2}([a-zA-Z0-9-_]+)`)
  parameterMatches := parameterRegex.FindAllStringSubmatch(formatted, -1)
  for _, parameterMatch := range parameterMatches {
    formatted = strings.ReplaceAll(formatted, parameterMatch[0], output.Blue(parameterMatch[0]))
  }

  environmentVariableRegex := regexp.MustCompile(`(^|\s)\b([^\s:;]+)\b=\b([^\s:;]+)\b`)
  environmentVariableMatches := environmentVariableRegex.FindAllStringSubmatch(formatted, -1)
  for _, environmentVariableMatch := range environmentVariableMatches {
    variable := environmentVariableMatch[0]
    variable = strings.ReplaceAll(variable, environmentVariableMatch[2], output.Cyan(environmentVariableMatch[2]))
    variable = strings.ReplaceAll(variable, environmentVariableMatch[3], output.Magenta(environmentVariableMatch[3]))
    formatted = strings.ReplaceAll(formatted, environmentVariableMatch[0], variable)
  }

  formatted = strings.ReplaceAll(formatted, "aux4", output.Bold("aux4"))

	return formatted
}
