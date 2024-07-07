package man

import (
	"aux4/core"
	"aux4/output"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const spacing = "  "
const lineLength = 100

func Help(profile core.Profile, json bool, long bool) {
	if json {
		helpJson(profile)
		return
	}

	for i, command := range profile.Commands {
		if i > 0 {
			output.Out(output.StdOut).Println("")
		}
		HelpCommand(command, json, long)
	}
}

func HelpCommand(command core.Command, json bool, long bool) {
	if json {
		helpCommandJson(command)
		return
	}

	outputHelp := strings.Builder{}
	commandName := command.Name

	outputHelp.WriteString(output.Yellow(output.Bold(commandName)))

	if long {
		description := ""
		if command.Help != nil {
			description = breakLines(command.Help.Text, lineLength, "")
		}

		if description != "" {
			outputHelp.WriteString("\n")
			outputHelp.WriteString(description)
		}
	} else {
		description := ""
		if command.Help != nil {
			description = command.Help.Text
			if len(description) > 100 {
				if strings.Index(description, ".") > -1 {
					description = description[:strings.Index(description, ".")+1]
				}
				if len(description) > 100 {
					description = description[:100] + "..."
				}
			}

			outputHelp.WriteString("\n")
			outputHelp.WriteString(description)
		}
	}

	if command.Help != nil && command.Help.Variables != nil && len(command.Help.Variables) > 0 {
		outputHelp.WriteString("\n")
		if long {
			outputHelp.WriteString("\n")
		}

		variablesHelp := strings.Builder{}

		for i, variable := range command.Help.Variables {
			if i > 0 {
				if long {
					variablesHelp.WriteString("\n")
				}
				variablesHelp.WriteString("\n")
			}

			variablesHelp.WriteString(spacing)
			variablesHelp.WriteString(output.Cyan("--"))
			variablesHelp.WriteString(output.Cyan(variable.Name))

			if long {
				if variable.Text != "" {
					variablesHelp.WriteString("\n")
					variablesHelp.WriteString(breakLines(variable.Text, lineLength, spacing+spacing))
				}

				if variable.Options != nil && len(variable.Options) > 0 {
					variablesHelp.WriteString("\n\n")
					variablesHelp.WriteString(spacing)
					variablesHelp.WriteString(spacing)
					variablesHelp.WriteString(output.Bold("Options:"))
					variablesHelp.WriteString("\n")

					for i, option := range variable.Options {
						if i > 0 {
							variablesHelp.WriteString("\n")
						}
						variablesHelp.WriteString(spacing)
						variablesHelp.WriteString(spacing)
						variablesHelp.WriteString("* ")
						variablesHelp.WriteString(output.Green(option))
					}
				}

				if variable.Default != nil {
					variablesHelp.WriteString("\n\n")
					variablesHelp.WriteString(spacing)
					variablesHelp.WriteString(spacing)
					variablesHelp.WriteString(output.Bold("Default: "))
					variablesHelp.WriteString(output.Italic(*variable.Default))
				}

				if variable.Env != "" {
					variablesHelp.WriteString("\n\n")
					variablesHelp.WriteString(spacing)
					variablesHelp.WriteString(spacing)
					variablesHelp.WriteString(output.Bold("Environment variable: "))
					variablesHelp.WriteString(output.Green(variable.Env))
				}
			}
		}

		outputHelp.WriteString(variablesHelp.String())
	}

	output.Out(output.StdOut).Println(outputHelp.String())
}

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

		if prefix == "log" {
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

  formatted = strings.ReplaceAll(formatted, "aux4", output.Bold("aux4"))

	return formatted
}

func helpJson(profile core.Profile) {
	output.Out(output.StdOut).Print("[")
	for i, command := range profile.Commands {
		if i > 0 {
			output.Out(output.StdOut).Print(",")
		}
		helpCommandJson(command)
	}
	output.Out(output.StdOut).Print("]")
}

func helpCommandJson(command core.Command) {
	man := Man{
		Name:       command.Name,
		Parameters: make([]ManParameter, 0),
	}

	if command.Help != nil {
		man.Text = command.Help.Text

		if command.Help.Variables != nil && len(command.Help.Variables) > 0 {
			for _, v := range command.Help.Variables {
				variable := ManParameter{
					Name:    v.Name,
					Text:    v.Text,
					Default: *v.Default,
					Env:     v.Env,
					Arg:     v.Arg,
					Options: v.Options,
				}

				man.Parameters = append(man.Parameters, variable)
			}
		}
	}

	value, err := json.Marshal(man)
	if err != nil {
		return
	}

	output.Out(output.StdOut).Print(string(value))
}

func maxCommandNameLength(commandNames []string) int {
	max := 0
	for _, commandName := range commandNames {
		if len(commandName) > max {
			max = len(commandName)
		}
	}
	return max
}

func breakLines(text string, maxLineLength int, spacing string) string {
	if len(text)+len(spacing) <= maxLineLength && strings.Index(text, "\n") == -1 {
		return spacing + strings.Trim(text, " ")
	}

	spacingLength := len(spacing)
	maxLength := maxLineLength - spacingLength

	remaining := text
	newText := strings.Builder{}

	for remaining != "" {
		if newText.Len() > 0 {
			newText.WriteString("\n")
		}

		newText.WriteString(spacing)

		if len(remaining)+len(spacing) < maxLength && strings.Index(remaining, "\n") == -1 {
			newText.WriteString(strings.Trim(remaining, " "))
			break
		}

		nextLine := ""

		nextBreak := strings.Index(remaining, "\n")
		if nextBreak > -1 {
			nextLine = remaining[:nextBreak]
			nextBreak += 1
		} else {
			end := maxLength + 1
			if len(remaining) < maxLength {
				end = len(remaining)
			}
			nextLine = remaining[:end]
			nextBreak = len(nextLine)
		}

		if len(nextLine) > maxLength {
			nextLine = nextLine[:maxLength]
			nextBreak = strings.LastIndex(nextLine, " ")

			if nextBreak == -1 {
				nextBreak = len(nextLine)
			}

			nextLine = nextLine[:nextBreak]
		}

		newText.WriteString(strings.Trim(nextLine, " "))
		remaining = strings.Trim(remaining[nextBreak:], " ")
	}

	return newText.String()
}

type Man struct {
	Name       string         `json:"name"`
	Text       string         `json:"text"`
	Parameters []ManParameter `json:"params"`
}

type ManParameter struct {
	Name    string   `json:"name"`
	Text    string   `json:"text"`
	Default string   `json:"default"`
	Env     string   `json:"env"`
	Arg     bool     `json:"arg"`
	Options []string `json:"options"`
}
