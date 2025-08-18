package param

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"aux4.dev/aux4/cmd"
	"aux4.dev/aux4/core"
	"aux4.dev/aux4/io"
	"aux4.dev/aux4/output"

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

type ParameterLookup interface {
	Get(parameters *Parameters, command core.Command, actions []string, name string) (any, error)
}

type ConfigLookup struct {
	load       bool
	parameters *io.OrderedMap
}

func (l *ConfigLookup) Get(parameters *Parameters, command core.Command, actions []string, name string) (any, error) {
	if !parameters.Has("configFile") && !parameters.Has("config") {
		return nil, nil
	}

	if l.load {
		value, found := l.parameters.Get(name)
		if !found {
			return nil, nil
		}
		return value, nil
	}

	args := []string{}

	if parameters.Has("configFile") {
		configFile := parameters.JustGet("configFile")
		args = append(args, "--file "+configFile.(string))
	}

	if parameters.Has("config") {
		config := parameters.JustGet("config")
		configParam := config.(string)
		if configParam != "true" {
			args = append(args, fmt.Sprintf("--name %s", configParam))
		}
	}

	l.load = true

	stdout, _, err := cmd.ExecuteCommandLineNoOutput(fmt.Sprintf("aux4 config get %s", strings.Join(args, " ")))
	if err != nil {
		return nil, nil
	}

	if stdout == "" {
		return nil, nil
	}

	jsonString := strings.TrimSpace(stdout)

	params := io.NewOrderedMap()
	err = json.Unmarshal([]byte(jsonString), params)
	if err != nil {
		return nil, nil
	}

	l.parameters = params

	value, _ := params.Get(name)
	return value, nil
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
		variable, exists := command.Help.GetVariable(name)
		if !exists {
			return nil, nil
		}

		if variable.Encrypt {
			encryptedParameter = parameters.JustGet(name)
			if encryptedParameter == nil {
				return nil, nil
			}
		} else {
			return nil, nil
		}
	}

	iv, _ := parameters.Get(command, actions, "iv")
	secret, _ := parameters.Get(command, actions, "secret")

	if iv == nil || secret == nil {
		return nil, nil
	}

	stdout, _, err := cmd.ExecuteCommandLineNoOutput(fmt.Sprintf("aux4 encrypter decrypt --iv %s --secret %s %s", iv, secret, encryptedParameter.(string)))
	if err != nil {
		return nil, core.InternalError("Error decrypting the value of '"+name+"' (it may not be encrypted)", nil)
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

	value := os.Getenv(variable.Env)

	if value == "" {
		return nil, nil
	}

	return value, nil
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

	value := *variable.Default

	iv, _ := parameters.Get(command, actions, "iv")
	secret, _ := parameters.Get(command, actions, "secret")

	if variable.Encrypt && iv != nil && secret != nil {
		stdout, _, err := cmd.ExecuteCommandLineNoOutput(fmt.Sprintf("aux4 encrypter decrypt --iv %s --secret %s %s", iv, secret, value))
		if err != nil {
			return nil, core.InternalError("Error decrypting the value of '"+name+"' (it may not be encrypted)", nil)
		}
		value = strings.TrimSpace(stdout)
	}

	return value, nil
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

	if variable.Hide || variable.Encrypt {
		prompt = promptui.Prompt{Label: text, HideEntered: true, Mask: '*'}
	} else {
		prompt = promptui.Prompt{Label: text, HideEntered: true}
	}

	text, err := prompt.Run()
	if err != nil {
		if err.Error() == "^C" {
			return "", core.UserAbortedError()
		}
		return "", core.InternalError("Error to enter the value of "+variable.Name, err)
	}

	return text, nil
}

func promptSelect(variable core.CommandHelpVariable) (string, error) {
	var text = fmt.Sprintf("%s %s", variable.Name, output.Gray(variable.Text))

	prompt := promptui.Select{Label: text, HideSelected: true, Items: variable.Options}

	_, text, err := prompt.Run()
	if err != nil {
		if err.Error() == "^C" {
			return "", core.UserAbortedError()
		}
		return "", core.InternalError("Error to select the value of "+variable.Name, err)
	}

	return text, nil
}
