package main

import (
	"io"
  "os"
	"os/exec"
	"strings"
  "github.com/manifoldco/promptui"
)

func VirtualCommandExecutorFactory(command string) VirtualCommandExecutor {
	if strings.HasPrefix(command, "profile:") {
		return &ProfileCommandExecutor{Command: command}
	} else if strings.HasPrefix(command, "set:") {
		return &SetCommandExecutor{Command: command}
	} else if strings.HasPrefix(command, "confirm:") {
    return &ConfirmCommandExecutor{Command: command}
	} else if strings.HasPrefix(command, "log:") {
		return &LogCommandExecutor{Command: command}
	} else if strings.HasPrefix(command, "debug:") {
		return &DebugCommandExecutor{Command: command}
	} else if strings.HasPrefix(command, "alias:") {
    return &AliasCommandExecutor{Command: command}
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
	instruction, err := InjectParameters(command, executor.Command, actions, params)
  if err != nil {
    return err
  }
	Out(Debug).Println(instruction)
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
		value, err := InjectParameters(command, valueExpression, actions, params)
    if err != nil {
      return err
    }
		params.Set(name, value)
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

  cmd := exec.Command("bash", "-c", expression)
  cmd.Stdin = os.Stdin
  cmd.Stdout = os.Stdout
  cmd.Stderr = os.Stderr

  if err := cmd.Run(); err != nil {
    return InternalError("Error executing command", err)
  }

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

	cmd := exec.Command("bash", "-c", instruction)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return InternalError("Error getting stdout from command", err)
	}

  stderr, err := cmd.StderrPipe()
  if err != nil {
    return InternalError("Error getting stderr from command", err)
  }

  if err := cmd.Start(); err != nil {
    return InternalError("Error executing command", err)
  }

  errorData, err := io.ReadAll(stderr)
  if err != nil {
    return InternalError("Error reading the command error output", err)
  }

  Out(StdErr).Print(string(errorData))

	data, err := io.ReadAll(stdout)
	if err != nil {
		return InternalError("Error reading the command output", err)
	}

  response := string(data)

  params.Update(("response"), response)
	Out(StdOut).Print(response)

	if err := cmd.Wait(); err != nil {
		return InternalError("Error waiting the command execute", err)
	}

	return nil
}
