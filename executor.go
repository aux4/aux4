package main

import (
	"encoding/json"
	"github.com/manifoldco/promptui"
	"os"
	"os/exec"
	"reflect"
	"strings"
)

func VirtualCommandExecutorFactory(command string) VirtualCommandExecutor {
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

func (executor *ProfileCommandExecutor) Execute(env *VirtualEnvironment, command *VirtualCommand, actions []string, params *Parameters) error {
	profileNameExpression := strings.TrimPrefix(executor.Command, "profile:")
	profileName, err := InjectParameters(command, profileNameExpression, actions, params)
	if err != nil {
		return err
	}
	env.SetProfile(profileName)
	return env.Execute(actions[1:], params)
}

type DebugCommandExecutor struct {
	Command string
}

func (executor *DebugCommandExecutor) Execute(env *VirtualEnvironment, command *VirtualCommand, actions []string, params *Parameters) error {
	expression := strings.TrimPrefix(executor.Command, "debug:")

	instruction, err := InjectParameters(command, expression, actions, params)
	if err != nil {
		return err
	}
	Out(Debug).Println(instruction)
	return nil
}

type JsonCommandExecutor struct {
	Command string
}

func (executor *JsonCommandExecutor) Execute(env *VirtualEnvironment, command *VirtualCommand, actions []string, params *Parameters) error {
	expression := strings.TrimPrefix(executor.Command, "json:")

	instruction, err := InjectParameters(command, expression, actions, params)
	if err != nil {
		return err
	}

	stdout, _, err := ExecuteCommandLine(instruction)
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

func (executor *EachCommandExecutor) Execute(env *VirtualEnvironment, command *VirtualCommand, actions []string, params *Parameters) error {
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
    return InternalError("response is not array", nil)
  }

	for index, item := range list {
		if item == "" {
			continue
		}

		params.Update("index", index)
		params.Update("item", item)

		instruction, err := InjectParameters(command, expression, actions, params)
		if err != nil {
			return err
		}

		stdout, _, err := ExecuteCommandLine(instruction)
		if err != nil {
			return err
		}

		Out(StdOut).Print(stdout)
	}

	return nil
}

type ConfirmCommandExecutor struct {
	Command string
}

func (executor *ConfirmCommandExecutor) Execute(env *VirtualEnvironment, command *VirtualCommand, actions []string, params *Parameters) error {
	text := strings.TrimPrefix(executor.Command, "confirm:")

	instruction, err := InjectParameters(command, text, actions, params)
	if err != nil {
		return err
	}

	prompt := promptui.Prompt{
		Label:     instruction,
		IsConfirm: true,
	}

	result, _ := prompt.Run()
	if result != "y" {
		return UserAbortedError()
	}

	return nil
}

type LogCommandExecutor struct {
	Command string
}

func (executor *LogCommandExecutor) Execute(env *VirtualEnvironment, command *VirtualCommand, actions []string, params *Parameters) error {
	text := strings.TrimPrefix(executor.Command, "log:")
	instruction, err := InjectParameters(command, text, actions, params)
	if err != nil {
		return err
	}
	Out(StdOut).Println(instruction)
	return nil
}

type SetCommandExecutor struct {
	Command string
}

func (executor *SetCommandExecutor) Execute(env *VirtualEnvironment, command *VirtualCommand, actions []string, params *Parameters) error {
	expression := strings.TrimPrefix(executor.Command, "set:")
	multiple := strings.Split(expression, ";")
	for _, pair := range multiple {
		parts := strings.Split(pair, "=")
		name := parts[0]
		valueExpression := parts[1]

		if strings.HasPrefix(valueExpression, "!") {
			valueExpression = strings.TrimPrefix(valueExpression, "!")

			instruction, err := InjectParameters(command, valueExpression, actions, params)
			if err != nil {
				return err
			}

			stdout, _, err := ExecuteCommandLine(instruction)
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
			value, err := InjectParameters(command, valueExpression, actions, params)
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

func (executor *AliasCommandExecutor) Execute(env *VirtualEnvironment, command *VirtualCommand, actions []string, params *Parameters) error {
	expression := strings.TrimPrefix(executor.Command, "alias:")

	if len(actions) > 1 {
		expression = expression + " " + strings.Join(actions[1:], " ")
	}

	stringParams := params.String()
	if stringParams != "" {
		expression = expression + " " + stringParams
	}

  instruction, err := InjectParameters(command, expression, actions, params)
  if err != nil {
    return err
  }

	cmd := exec.Command("bash", "-c", instruction)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return InternalError("Error executing command", err)
	}

	return nil
}

type NoutCommandExecutor struct {
  Command string
}

func (executor *NoutCommandExecutor) Execute(env *VirtualEnvironment, command *VirtualCommand, actions []string, params *Parameters) error {
  expression := strings.TrimPrefix(executor.Command, "nout:")

  instruction, err := InjectParameters(command, expression, actions, params)
  if err != nil {
    return err
  }

  stdout, stderr, err := ExecuteCommandLine(instruction)
	if err != nil {
		Out(StdErr).Print(stderr)
		Out(StdOut).Print(stdout)
		return err
	}

	params.Update("response", strings.TrimSpace(stdout))

	return nil
}

type StdinCommandExecutor struct {
  Command string
}

func (executor *StdinCommandExecutor) Execute(env *VirtualEnvironment, command *VirtualCommand, actions []string, params *Parameters) error {
  expression := strings.TrimPrefix(executor.Command, "stdin:")

  instruction, err := InjectParameters(command, expression, actions, params)
  if err != nil {
    return err
  }

  stdout, stderr, err := ExecuteCommandLineWithStdIn(instruction)
  if err != nil {
    Out(StdErr).Print(stderr)
    Out(StdOut).Print(stdout)
    return err
  }

	Out(StdErr).Print(stderr)

	params.Update("response", strings.TrimSpace(stdout))
	Out(StdOut).Print(stdout)

  return nil
}

type CommandLineExecutor struct {
	Command string
}

func (executor *CommandLineExecutor) Execute(env *VirtualEnvironment, command *VirtualCommand, actions []string, params *Parameters) error {
	instruction, err := InjectParameters(command, executor.Command, actions, params)
	if err != nil {
		return err
	}

	stdout, stderr, err := ExecuteCommandLine(instruction)
	if err != nil {
		Out(StdErr).Print(stderr)
		Out(StdOut).Print(stdout)
		return err
	}

	Out(StdErr).Print(stderr)

	params.Update("response", strings.TrimSpace(stdout))
	Out(StdOut).Print(stdout)

	return nil
}
