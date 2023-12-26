package main

import (
	"github.com/manifoldco/promptui"
	"os"
	"strings"
)

func ParameterLookups() []ParameterLookup {
	return []ParameterLookup{
		&ArgLookup{},
		&EnvironmentVariableLookup{},
		&DefaultLookup{},
		&PromptLookup{},
	}
}

type EnvironmentVariableLookup struct {
}

func (l EnvironmentVariableLookup) Get(parameters *Parameters, command *VirtualCommand, actions []string, name string) (any, error) {
	variable, ok := command.Help.GetVariable(name)
	if !ok {
		return nil, nil
	}

	if variable.Env == "" {
		return nil, nil
	}

	return os.Getenv(variable.Env), nil
}

type DefaultLookup struct {
}

func (l DefaultLookup) Get(parameters *Parameters, command *VirtualCommand, actions []string, name string) (any, error) {
	variable, ok := command.Help.GetVariable(name)
	if !ok {
		return nil, nil
	}

	if variable.Default == "" {
		return nil, nil
	}

	return variable.Default, nil
}

type ArgLookup struct {
}

func (l ArgLookup) Get(parameters *Parameters, command *VirtualCommand, actions []string, name string) (any, error) {
	variable, ok := command.Help.GetVariable(name)
	if !ok {
		return nil, nil
	}

	if !variable.Arg {
		return nil, nil
	}

	if len(actions) != 2 {
		return nil, nil
	}

	return actions[1], nil
}

type PromptLookup struct {
}

func (l PromptLookup) Get(parameters *Parameters, command *VirtualCommand, action []string, name string) (any, error) {
	variable, ok := command.Help.GetVariable(name)
	if !ok {
		return nil, nil
	}

	if variable.Default != "" {
		return nil, nil
	}
 
  var text string 
  var err error
  
	if variable.Options != nil && len(variable.Options) > 0 {
    text, err = promptSelect(variable)	
	} else {
    text, err = promptText(variable)
  }

  if err != nil {
    return nil, err
  }

	return strings.TrimSpace(text), nil
}

func promptText(variable *CommandHelpVariable) (string, error) {
	prompt := promptui.Prompt{Label: variable.Text}

	text, err := prompt.Run()
	if err != nil {
    if err.Error() == "^C" { 
      return "", UserAbortedError()
    }
    return "", InternalError("Error to enter the value of " + variable.Name, err) 
	}

  return text, nil
}

func promptSelect(variable *CommandHelpVariable) (string, error) {
  prompt := promptui.Select{Label: variable.Text, Items: variable.Options}

  _, text, err := prompt.Run()
  if err != nil {
    if err.Error() == "^C" { 
      return "", UserAbortedError()
    }
    return "", InternalError("Error to select the value of " + variable.Name, err) 
  }

  return text, nil
}
