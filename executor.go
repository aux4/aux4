package main

import (
	"io"
	"os/exec"
	"strings"
)

func VirtualCommandExecutorFactory(command string) VirtualCommandExecutor {
	if strings.HasPrefix(command, "profile:") {
		return &ProfileCommandExecutor{Command: command}
	} else if strings.HasPrefix(command, "set:") {
		return &SetCommandExecutor{Command: command}
	} else if strings.HasPrefix(command, "log:") {
		return &LogCommandExecutor{Command: command}
	} else if strings.HasPrefix(command, "debug:") {
		return &DebugCommandExecutor{Command: command}
	}
	return &CommandLineExecutor{Command: command}
}

type ProfileCommandExecutor struct {
	Command string
}

func (executor *ProfileCommandExecutor) Execute(env *VirtualEnvironment, command *VirtualCommand, actions []string, params *Parameters) error {
	profileNameExpression := strings.TrimPrefix(executor.Command, "profile:")
	profileName := InjectParameters(command, profileNameExpression, params)
	env.SetProfile(profileName)
	return env.Execute(actions[1:], params)
}

type DebugCommandExecutor struct {
	Command string
}

func (executor *DebugCommandExecutor) Execute(env *VirtualEnvironment, command *VirtualCommand, actions []string, params *Parameters) error {
	instruction := InjectParameters(command, executor.Command, params)
	Out(Debug).Println(instruction)
	return nil
}

type LogCommandExecutor struct {
	Command string
}

func (executor *LogCommandExecutor) Execute(env *VirtualEnvironment, command *VirtualCommand, actions []string, params *Parameters) error {
	instruction := InjectParameters(command, executor.Command, params)
	Out(StdOut).Println(instruction)
	return nil
}

type SetCommandExecutor struct {
	Command string
}

func (executor *SetCommandExecutor) Execute(env *VirtualEnvironment, command *VirtualCommand, actions []string, params *Parameters) error {
	expression := strings.TrimPrefix(executor.Command, "set:")
	multuple := strings.Split(expression, ";")
	for _, pair := range multuple {
		parts := strings.Split(pair, "=")
		name := parts[0]
		valueExpression := parts[1]
		value := InjectParameters(command, valueExpression, params)
		params.Set(name, value)
	}
	return nil
}

type CommandLineExecutor struct {
	Command string
}

func (executor *CommandLineExecutor) Execute(env *VirtualEnvironment, command *VirtualCommand, actions []string, params *Parameters) error {
	instruction := InjectParameters(command, executor.Command, params)

	cmd := exec.Command("bash", "-c", instruction)

	stdout, err := cmd.StdoutPipe()

	if err != nil {
		return InternalError("Error getting stdout from command", err)
	}

  if err := cmd.Start(); err != nil {
    return InternalError("Error executing command", err)
  }

	data, err := io.ReadAll(stdout)

	if err != nil {
		return InternalError("Error reading the command output", err)
	}

	Out(StdOut).Println(string(data))

	if err := cmd.Wait(); err != nil {
		return InternalError("Error waiting the command execute", err)
	}

	return nil
}
