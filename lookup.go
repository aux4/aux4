package main

import (
  "strings"
  "os"
  "bufio"
)

func ParameterLookups() []ParameterLookup {
  return []ParameterLookup{
    &DefaultLookup{},
    &ArgLookup{},
    &PromptLookup{},
  }
}

type DefaultLookup struct {
}

func (l DefaultLookup) Get(parameters *Parameters, command *VirtualCommand, actions []string, name string) any {
  variable, ok := command.Help.GetVariable(name)
  if !ok {
    return nil
  }

  if variable.Default == "" {
    return nil
  }

  return variable.Default
}

type ArgLookup struct {
}

func (l ArgLookup) Get(parameters *Parameters, command *VirtualCommand, actions []string, name string) any {
  variable, ok := command.Help.GetVariable(name)
  if !ok {
    return nil
  }

  if !variable.Arg {
    return nil
  }

  if len(actions) != 2 {
    return nil
  }

  return actions[1]
}

type PromptLookup struct {
}

func (l PromptLookup) Get(parameters *Parameters, command *VirtualCommand, action []string, name string) any {
  variable, ok := command.Help.GetVariable(name)
  if !ok {
    return nil
  }

  if variable.Default != "" {
    return nil
  }

  Out(StdOut).Print(variable.Text + ": ")
  reader := bufio.NewReader(os.Stdin)
  text, _ := reader.ReadString('\n')

  return strings.TrimSpace(text)
}
