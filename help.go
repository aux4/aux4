package main

import (
	"encoding/json"
	"strings"
)

const spacing = "  "
const lineLength = 100

func Help(profile *VirtualProfile, json bool) {
  if json {
    helpJson(profile)
    return
  }

	for i, commandName := range profile.CommandsOrdered {
		if i > 0 {
			Out(StdOut).Println("")
		}
		command := profile.Commands[commandName]
		HelpCommand(command, json)
	}
}

func HelpCommand(command *VirtualCommand, json bool) {
	if json {
		helpCommandJson(command)
		return
	}

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

			if variable.Options != nil && len(variable.Options) > 0 {
				variablesHelp.WriteString("\n")
				variablesHelp.WriteString(spacing)
				variablesHelp.WriteString(Bold("Options:"))
				variablesHelp.WriteString("\n")

				for i, option := range variable.Options {
					if i > 0 {
						variablesHelp.WriteString("\n")
					}
					variablesHelp.WriteString(spacing)
					variablesHelp.WriteString("* ")
					variablesHelp.WriteString(Green(option))
				}
			}

			variablesHelp.WriteString("\n")
		}

		output.WriteString(breakLines(variablesHelp.String(), lineLength, spacing))
	}

	Out(StdOut).Println(output.String())
}

func helpJson(profile *VirtualProfile) {
  Out(StdOut).Print("[")
  for i, commandName := range profile.CommandsOrdered {
    if i > 0 {
      Out(StdOut).Print(",")
    }
    command := profile.Commands[commandName]
    helpCommandJson(command)
  }
  Out(StdOut).Print("]")
}

func helpCommandJson(command *VirtualCommand) {
	man := Man{
		Name:      command.Name,
		Parameters: make([]ManParameter, 0),
	}

	if command.Help != nil {
		man.Text = command.Help.Text

		if command.Help.Variables != nil && len(command.Help.Variables) > 0 {
			for _, v := range command.Help.Variables {
				variable := ManParameter{
					Name:    v.Name,
					Text:    v.Text,
					Default: v.Default,
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

	Out(StdOut).Print(string(value))
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

type Man struct {
	Name      string        `json:"name"`
	Text      string        `json:"text"`
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
