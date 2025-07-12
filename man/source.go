package man

import (
  "aux4.dev/aux4/core"
  "aux4.dev/aux4/output"
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

  executorPrefixRegex := regexp.MustCompile(`^([a-zA-Z0-9]+):(.+)`)
	executorPrefixMatch := executorPrefixRegex.FindStringSubmatch(formatted)
	if len(executorPrefixMatch) > 0 {
		prefix := executorPrefixMatch[1]
		command := executorPrefixMatch[2]

		formatted = output.Magenta(prefix, ":")

		if prefix == "log" || prefix == "debug" || prefix == "confirm" {
			formatted += output.Green(command)
		} else if prefix == "profile" {
			formatted += output.Cyan(command)
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

  paramRegex := regexp.MustCompile("(param|params|value|values|if)\\(([^)]+)\\)")
	paramMatches := paramRegex.FindAllStringSubmatch(formatted, -1)
	for _, paramMatch := range paramMatches {
		match := paramMatch[0]
		formattedMatch := match
		formattedMatch = output.ColorText(formattedMatch, match, output.ColorMagenta)
		formattedMatch = output.ColorText(formattedMatch, paramMatch[2], output.ColorCyan)
		formatted = strings.ReplaceAll(formatted, match, formattedMatch)
  }

	textRegex := regexp.MustCompile(`"(.*?)"|'(.*?)'`)
	textMatches := textRegex.FindAllStringSubmatch(formatted, -1)
	for _, textMatch := range textMatches {
		formatted = output.ColorText(formatted, textMatch[0], output.ColorGreen)
	}

	filePathRegex := regexp.MustCompile(`\/?(\S*)\/(\S*)`)
	filePathMatches := filePathRegex.FindAllStringSubmatch(formatted, -1)
	for _, filePathMatch := range filePathMatches {
		formatted = output.ColorText(formatted, filePathMatch[0], output.ColorYellow)
	}

	variableRegex := regexp.MustCompile("\\$([a-zA-Z0-9]+)|\\$\\{([^}\\s]+)\\}")
	variableMatches := variableRegex.FindAllStringSubmatch(formatted, -1)
	for _, variableMatch := range variableMatches {
		formatted = output.ColorText(formatted, variableMatch[0], output.ColorCyan)
	}

  parameterRegex := regexp.MustCompile(`\s-{1,2}([a-zA-Z0-9-_]+)`)
  parameterMatches := parameterRegex.FindAllStringSubmatch(formatted, -1)
  for _, parameterMatch := range parameterMatches {
    formatted = output.ColorText(formatted, parameterMatch[0], output.ColorBlue)
  }

  environmentVariableRegex := regexp.MustCompile(`(^|\s)\b([^\s:;]+)\b=\b([^\s:;]+)\b`)
  environmentVariableMatches := environmentVariableRegex.FindAllStringSubmatch(formatted, -1)
  for _, environmentVariableMatch := range environmentVariableMatches {
    variable := environmentVariableMatch[0]
    variable = output.ColorText(variable, environmentVariableMatch[2], output.ColorCyan)
    variable = output.ColorText(variable, environmentVariableMatch[3], output.ColorMagenta)
    formatted = strings.ReplaceAll(formatted, environmentVariableMatch[0], variable)
  }

  formatted = output.ColorText(formatted, "aux4", output.FormatBold)

	return formatted
}
