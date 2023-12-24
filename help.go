package main

import (
	"strings"
)

const spacing = "  "
const lineLength = 100

func Help(profile *VirtualProfile) {
	for _, commandName := range profile.CommandsOrdered {
		command := profile.Commands[commandName]
		HelpCommand(command)
	}
}

func HelpCommand(command *VirtualCommand) {
	output := strings.Builder{}
	commandName := command.Name

	output.WriteString(Yellow(commandName))

	description := ""
	if command.Help != nil {
		description = breakLines(command.Help.Text, lineLength, "")
	}

	if description != "" {
		output.WriteString("\n")
		output.WriteString(description)
	}

	if command.Help != nil && command.Help.Variables != nil && len(command.Help.Variables) > 0 {
		output.WriteString("\n\n")

		variablesHelp := strings.Builder{}

		for i, variable := range command.Help.Variables {
			if i > 0 {
				variablesHelp.WriteString("\n")
			}

			variablesHelp.WriteString(spacing)
			variablesHelp.WriteString(Cyan("--"))
			variablesHelp.WriteString(Cyan(variable.Name))

      if variable.Env != "" {
        variablesHelp.WriteString(Green(" $", variable.Env))
      }

			if variable.Default != "" {
				variablesHelp.WriteString(" [")
				variablesHelp.WriteString(Italic(variable.Default))
				variablesHelp.WriteString("]")
			}

			if variable.Text != "" {
				variablesHelp.WriteString("\n")
				variablesHelp.WriteString(spacing)
				variablesHelp.WriteString(variable.Text)
			}

			variablesHelp.WriteString("\n")
		}

		output.WriteString(breakLines(variablesHelp.String(), lineLength, spacing))
	}

	Out(StdOut).Println(output.String())
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
	if len(text) <= maxLineLength && strings.Index(text, "\n") == -1 {
		return strings.Trim(text, " ") 
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

		if len(remaining) < maxLength && strings.Index(remaining, "\n") == -1 {
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
