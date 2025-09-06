package executor

import (
	"fmt"
	"strings"

	"aux4.dev/aux4/core"
	"aux4.dev/aux4/engine"
	"aux4.dev/aux4/engine/param"
	"aux4.dev/aux4/output"
)

type Aux4AutocompleteExecutor struct {
}

func (executor *Aux4AutocompleteExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	cmdInput := params.JustGet("cmd")
	if cmdInput == nil || cmdInput == "" {
		return nil
	}

	cmdStr := fmt.Sprintf("%v", cmdInput)
	
	args := param.ExtractArgs(cmdStr)
	if len(args) == 0 {
		return nil
	}

	if len(args) > 0 && args[0] == "aux4" {
		args = args[1:]
	}

	if len(args) == 0 {
		suggestions := getProfileCommands(env, "main")
		outputSuggestions(suggestions)
		return nil
	}

	endsWithOption := len(args) > 0 && strings.HasPrefix(args[len(args)-1], "--")
	endsWithPartialFlag := len(args) > 0 && strings.HasPrefix(args[len(args)-1], "-") && !strings.HasPrefix(args[len(args)-1], "--")
	
	currentProfile := "main"
	currentActions := make([]string, 0)
	
	for i, arg := range args {
		if strings.HasPrefix(arg, "-") {
			if i == len(args)-1 && (endsWithOption || endsWithPartialFlag) {
				cmd := findCommandInProfile(env, currentProfile, currentActions)
				if cmd != nil {
					suggestions := getCommandVariables(*cmd, arg)
					outputSuggestions(suggestions)
					return nil
				}
			}
			continue
		}

		profile := env.GetProfile(currentProfile)
		if profile == nil {
			return nil
		}

		command, exists := profile.Commands[arg]
		if !exists {
			if i == len(args)-1 {
				suggestions := getProfileCommands(env, currentProfile)
				filtered := make([]string, 0)
				for _, s := range suggestions {
					if strings.HasPrefix(s, arg) {
						filtered = append(filtered, s)
					}
				}
				outputSuggestions(filtered)
			}
			return nil
		}

		currentActions = append(currentActions, arg)

		if i == len(args)-1 {
			newProfile := getRedirectProfile(command)
			if newProfile != "" {
				suggestions := getProfileCommands(env, newProfile)
				outputSuggestions(suggestions)
				return nil
			} else {
				suggestions := getCommandVariables(command, "")
				outputSuggestions(suggestions)
				return nil
			}
		}

		newProfile := getRedirectProfile(command)
		if newProfile != "" {
			currentProfile = newProfile
			currentActions = make([]string, 0)
		}
	}

	return nil
}

func getRedirectProfile(command core.Command) string {
	if command.Execute != nil {
		for _, execute := range command.Execute {
			if profileName, found := strings.CutPrefix(execute, "profile:"); found {
				return profileName
			}
		}
	}
	return ""
}

func getProfileCommands(env *engine.VirtualEnvironment, profileName string) []string {
	profile := env.GetProfile(profileName)
	if profile == nil {
		return []string{}
	}

	commands := make([]string, 0)
	for _, cmdName := range profile.CommandsOrdered {
		command := profile.Commands[cmdName]
		if !command.Private {
			commands = append(commands, cmdName)
		}
	}

	return commands
}

func findCommandInProfile(env *engine.VirtualEnvironment, profileName string, actions []string) *core.Command {
	if len(actions) == 0 {
		return nil
	}

	currentProfile := profileName
	
	for i, action := range actions {
		profile := env.GetProfile(currentProfile)
		if profile == nil {
			return nil
		}

		command, exists := profile.Commands[action]
		if !exists {
			return nil
		}

		if i == len(actions)-1 {
			return &command
		}

		newProfile := getRedirectProfile(command)
		if newProfile != "" {
			currentProfile = newProfile
		}
	}

	return nil
}

func getCommandVariables(command core.Command, partial string) []string {
	if command.Help == nil || command.Help.Variables == nil {
		return []string{}
	}

	suggestions := make([]string, 0)
	
	for _, variable := range command.Help.Variables {
		if variable.Hide {
			continue
		}

		flag := "--" + variable.Name
		if partial == "" || strings.HasPrefix(flag, partial) {
			suggestions = append(suggestions, flag)
		}

		if valuePartial, found := strings.CutPrefix(partial, flag+"="); found {
			for _, option := range variable.Options {
				if strings.HasPrefix(option, valuePartial) {
					suggestions = append(suggestions, flag+"="+option)
				}
			}
		} else if partial == flag && len(variable.Options) > 0 {
			for _, option := range variable.Options {
				suggestions = append(suggestions, option)
			}
		}
	}

	return suggestions
}

func outputSuggestions(suggestions []string) {
	for _, suggestion := range suggestions {
		output.Out(output.StdOut).Println(suggestion)
	}
}