package executor

import (
	"aux4/cmd"
	"aux4/core"
	"aux4/engine"
	"aux4/engine/param"
	"aux4/man"
	"aux4/output"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strings"

	"github.com/manifoldco/promptui"
)

func Execute(env *engine.VirtualEnvironment, actions []string, params *param.Parameters) error {
	virtualProfile := env.GetProfile(env.CurrentProfile)
	if virtualProfile == nil {
		return core.InternalError(fmt.Sprintf("Profile not found: %s", env.CurrentProfile), nil)
	}

	if len(actions) == 0 {
		json := params.JustGet("json")
		isJson := json == true || json == "true"

		help := params.JustGet("help")
		isHelp := help == true || help == "true"

		var profile = virtualProfile.GetProfile()
		man.Help(profile, isJson, isHelp)

		return nil
	}

	commandName := actions[0]
	command, exists := virtualProfile.Commands[commandName]
	if !exists {
		return core.CommandNotFoundError(commandName)
	}

	if params.Has("help") && len(actions) == 1 {
		json := params.JustGet("json")
		isJson := json == true || json == "true"

		help := params.JustGet("help")
		isHelp := help == true || help == "true"

		man.HelpCommand(command, isJson, isHelp)
		return nil
	} 

  if params.Has("show-source") && len(actions) == 1 {
    man.ShowCommandSource(command)
    return nil
  }

	for _, commandLine := range command.Execute {
		executor := commandExecutorFactory(commandLine)
		err := executor.Execute(env, command, actions, params)
		if err != nil {
			return err
		}
	}

	if len(command.Execute) == 0 {
		key := fmt.Sprintf("%s.%s", virtualProfile.Name, command.Name)
		executor, exists := env.Registry.GetExecutor(key)
		if exists {
			err := executor.Execute(env, command, actions, params)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func commandExecutorFactory(command string) engine.VirtualCommandExecutor {
	if strings.HasPrefix(command, "profile:") {
		return &ProfileCommandExecutor{Command: command}
	} else if strings.HasPrefix(command, "set:") {
		return &SetCommandExecutor{Command: command}
	} else if strings.HasPrefix(command, "each:") {
		return &EachCommandExecutor{Command: command}
	} else if strings.HasPrefix(command, "confirm:") {
		return &ConfirmCommandExecutor{Command: command}
	} else if strings.HasPrefix(command, "log:") {
		return &LogCommandExecutor{Command: command}
	} else if strings.HasPrefix(command, "debug:") {
		return &DebugCommandExecutor{Command: command}
	} else if strings.HasPrefix(command, "alias:") {
		return &AliasCommandExecutor{Command: command}
	} else if strings.HasPrefix(command, "json:") {
		return &JsonCommandExecutor{Command: command}
	} else if strings.HasPrefix(command, "nout:") {
		return &NoutCommandExecutor{Command: command}
	} else if strings.HasPrefix(command, "stdin:") {
		return &StdinCommandExecutor{Command: command}
	}
	return &CommandLineExecutor{Command: command}
}

type ProfileCommandExecutor struct {
	Command string
}

func (executor *ProfileCommandExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	profileNameExpression := strings.TrimPrefix(executor.Command, "profile:")
	profileName, err := param.InjectParameters(command, profileNameExpression, actions, params)
	if err != nil {
		return err
	}
	env.SetProfile(profileName)
	return Execute(env, actions[1:], params)
}

type DebugCommandExecutor struct {
	Command string
}

func (executor *DebugCommandExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	expression := strings.TrimPrefix(executor.Command, "debug:")

	instruction, err := param.InjectParameters(command, expression, actions, params)
	if err != nil {
		return err
	}
	output.Out(output.Debug).Println(instruction)
	return nil
}

type JsonCommandExecutor struct {
	Command string
}

func (executor *JsonCommandExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	expression := strings.TrimPrefix(executor.Command, "json:")

	instruction, err := param.InjectParameters(command, expression, actions, params)
	if err != nil {
		return err
	}

	stdout, _, err := cmd.ExecuteCommandLine(instruction)
	if err != nil {
		return err
	}

	var data interface{}
	err = json.Unmarshal([]byte(stdout), &data)
	if err != nil {
		return err
	}

	params.Update("response", data)

	return nil
}

type EachCommandExecutor struct {
	Command string
}

func (executor *EachCommandExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	expression := strings.TrimPrefix(executor.Command, "each:")

	response := params.JustGet("response")

	var list []any

	typeOfResponse := reflect.TypeOf(response)
	if typeOfResponse.Kind() == reflect.Slice || typeOfResponse.Kind() == reflect.Array {
		list = response.([]any)
	} else if typeOfResponse.Kind() == reflect.String {
		lines := strings.Split(response.(string), "\n")
		list = make([]any, len(lines))
		for index, line := range lines {
			list[index] = line
		}
	} else {
		return core.InternalError("response is not array", nil)
	}

	for index, item := range list {
		if item == "" {
			continue
		}

		params.Update("index", index)
		params.Update("item", item)

		instruction, err := param.InjectParameters(command, expression, actions, params)
		if err != nil {
			return err
		}

		stdout, _, err := cmd.ExecuteCommandLine(instruction)
		if err != nil {
			return err
		}

		output.Out(output.StdOut).Print(stdout)
	}

	return nil
}

type ConfirmCommandExecutor struct {
	Command string
}

func (executor *ConfirmCommandExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	text := strings.TrimPrefix(executor.Command, "confirm:")

	instruction, err := param.InjectParameters(command, text, actions, params)
	if err != nil {
		return err
	}

	prompt := promptui.Prompt{
		Label:     instruction,
		IsConfirm: true,
	}

	result, _ := prompt.Run()
	if result != "y" {
		return core.UserAbortedError()
	}

	return nil
}

type LogCommandExecutor struct {
	Command string
}

func (executor *LogCommandExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	text := strings.TrimPrefix(executor.Command, "log:")
	instruction, err := param.InjectParameters(command, text, actions, params)
	if err != nil {
		return err
	}
	output.Out(output.StdOut).Println(instruction)
	return nil
}

type SetCommandExecutor struct {
	Command string
}

func (executor *SetCommandExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	expression := strings.TrimPrefix(executor.Command, "set:")
	multiple := strings.Split(expression, ";")
	for _, pair := range multiple {
		parts := strings.Split(pair, "=")
		name := parts[0]
		valueExpression := parts[1]

		if strings.HasPrefix(valueExpression, "!") {
			valueExpression = strings.TrimPrefix(valueExpression, "!")

			instruction, err := param.InjectParameters(command, valueExpression, actions, params)
			if err != nil {
				return err
			}

			stdout, _, err := cmd.ExecuteCommandLine(instruction)
			if err != nil {
				return err
			}

			params.Update(name, strings.TrimSpace(stdout))
		} else if strings.HasPrefix(valueExpression, "$") {
			value, err := params.Expr(command, actions, valueExpression)
			if err != nil {
				return err
			}
			params.Update(name, value)
		} else {
			value, err := param.InjectParameters(command, valueExpression, actions, params)
			if err != nil {
				return err
			}
			params.Update(name, value)
		}
	}
	return nil
}

type AliasCommandExecutor struct {
	Command string
}

func (executor *AliasCommandExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	expression := strings.TrimPrefix(executor.Command, "alias:")

	if len(actions) > 1 {
		expression = expression + " " + strings.Join(actions[1:], " ")
	}

	stringParams := params.String()
	if stringParams != "" {
		expression = expression + " " + stringParams
	}

	instruction, err := param.InjectParameters(command, expression, actions, params)
	if err != nil {
		return err
	}

	cmd := exec.Command("bash", "-c", instruction)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return core.InternalError("Error executing command", err)
	}

	return nil
}

type NoutCommandExecutor struct {
	Command string
}

func (executor *NoutCommandExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	expression := strings.TrimPrefix(executor.Command, "nout:")

	instruction, err := param.InjectParameters(command, expression, actions, params)
	if err != nil {
		return err
	}

	stdout, stderr, err := cmd.ExecuteCommandLine(instruction)
	if err != nil {
		output.Out(output.StdErr).Print(stderr)
		output.Out(output.StdOut).Print(stdout)
		return err
	}

	params.Update("response", strings.TrimSpace(stdout))

	return nil
}

type StdinCommandExecutor struct {
	Command string
}

func (executor *StdinCommandExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	expression := strings.TrimPrefix(executor.Command, "stdin:")

	instruction, err := param.InjectParameters(command, expression, actions, params)
	if err != nil {
		return err
	}

	stdout, stderr, err := cmd.ExecuteCommandLineWithStdIn(instruction)
	if err != nil {
		output.Out(output.StdErr).Print(stderr)
		output.Out(output.StdOut).Print(stdout)
		return err
	}

	output.Out(output.StdErr).Print(stderr)

	params.Update("response", strings.TrimSpace(stdout))
	output.Out(output.StdOut).Print(stdout)

	return nil
}

type CommandLineExecutor struct {
	Command string
}

func (executor *CommandLineExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	instruction, err := param.InjectParameters(command, executor.Command, actions, params)
	if err != nil {
		return err
	}

	stdout, stderr, err := cmd.ExecuteCommandLine(instruction)
	if err != nil {
		output.Out(output.StdErr).Print(stderr)
		output.Out(output.StdOut).Print(stdout)
		return err
	}

	output.Out(output.StdErr).Print(stderr)

	params.Update("response", strings.TrimSpace(stdout))
	output.Out(output.StdOut).Print(stdout)

	return nil
}
