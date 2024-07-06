package param

import (
	"aux4/core"
	"aux4/cmd"
	"aux4/output"
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func ParameterLookups() []ParameterLookup {
	return []ParameterLookup{
		&ArgLookup{},
		&EnvironmentVariableLookup{},
		&ConfigLookup{},
		&EncryptedParameterLookup{},
		&DefaultLookup{},
		&PromptLookup{},
	}
}

type ConfigLookup struct {
}

func (l ConfigLookup) Get(parameters *Parameters, command core.Command, actions []string, name string) (any, error) {
	if !cmd.IsCommandAvailable("aux4-config") {
		return nil, nil
	}

	args := []string{}

	if parameters.Has("configFile") {
		configFile, err := parameters.Get(command, actions, "configFile")
		if err != nil {
			return nil, err
		}
		args = append(args, "--file "+configFile.(string))
	}

	if parameters.Has("config") {
		config, err := parameters.Get(command, actions, "config")
		if err != nil {
			return nil, err
		}
		args = append(args, fmt.Sprintf("--name %s/", config.(string)))
	} else {
		args = append(args, "--name ")
	}

	stdout, _, err := cmd.ExecuteCommandLine(fmt.Sprintf("aux4-config get %s%s", strings.Join(args, " "), name))
	if err != nil {
		return nil, nil
	}

	if stdout == "" {
		return nil, nil
	}

	return strings.TrimSpace(stdout), nil
}

type EncryptedParameterLookup struct {
}

func (l EncryptedParameterLookup) Get(parameters *Parameters, command core.Command, actions []string, name string) (any, error) {
	if strings.HasPrefix(name, "encrypted") {
		return nil, nil
	}

	title := cases.Title(language.English)
	encryptedParameterName := "encrypted" + title.String(name)

	encryptedParameter, err := parameters.Get(command, actions, encryptedParameterName)
	if err != nil {
		return nil, err
	}

	if encryptedParameter == nil {
		return nil, nil
	}

	if !cmd.IsCommandAvailable("aux4-encrypt") {
		return nil, nil
	}

	stdout, _, err := cmd.ExecuteCommandLine(fmt.Sprintf("aux4-encrypt decrypt %s", encryptedParameter.(string)))
	if err != nil {
		return nil, err
	}

	return strings.TrimSpace(stdout), nil
}

type EnvironmentVariableLookup struct {
}

func (l EnvironmentVariableLookup) Get(parameters *Parameters, command core.Command, actions []string, name string) (any, error) {
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

func (l DefaultLookup) Get(parameters *Parameters, command core.Command, actions []string, name string) (any, error) {
	variable, ok := command.Help.GetVariable(name)
	if !ok {
		return nil, nil
	}

	if variable.Default == nil {
		return nil, nil
	}

	return *variable.Default, nil
}

type ArgLookup struct {
}

func (l ArgLookup) Get(parameters *Parameters, command core.Command, actions []string, name string) (any, error) {
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

func (l PromptLookup) Get(parameters *Parameters, command core.Command, action []string, name string) (any, error) {
	variable, ok := command.Help.GetVariable(name)
	if !ok {
		return nil, nil
	}

	if variable.Default != nil {
		return nil, nil
	}

	var text string
	var err error

	if variable.Options != nil && len(variable.Options) > 0 {
		text, err = promptSelect(*variable)
	} else {
		text, err = promptText(*variable)
	}

	if err != nil {
		return nil, err
	}

	return strings.TrimSpace(text), nil
}

func promptText(variable core.CommandHelpVariable) (string, error) {
	var prompt promptui.Prompt

	var text = fmt.Sprintf("%s %s", variable.Name, output.Gray(variable.Text))

	if variable.Hide {
		prompt = promptui.Prompt{Label: text, Mask: '*'}
	} else {
		prompt = promptui.Prompt{Label: text}
	}

	text, err := prompt.Run()
	if err != nil {
		if err.Error() == "^C" {
			return "", core.UserAbortedError()
		}
		return "", core.InternalError("Error to enter the value of "+variable.Name, err)
	}

	if variable.Encrypt {
		if !cmd.IsCommandAvailable("aux4-encrypt") {
			return text, nil
		}

		stdout, _, err := cmd.ExecuteCommandLine(fmt.Sprintf("aux4-encrypt encrypt %s", text))
		if err != nil {
			return text, err
		}

		text = strings.TrimSpace(stdout)
	}

	return text, nil
}

func promptSelect(variable core.CommandHelpVariable) (string, error) {
	var text = fmt.Sprintf("%s %s", variable.Name, output.Gray(variable.Text))

	prompt := promptui.Select{Label: text, Items: variable.Options}

	_, text, err := prompt.Run()
	if err != nil {
		if err.Error() == "^C" {
			return "", core.UserAbortedError()
		}
		return "", core.InternalError("Error to select the value of "+variable.Name, err)
	}

	return text, nil
}
