package main

import (
  "strings"
  "os"
)

type Aux4VersionExecutor struct {
}

func (executor *Aux4VersionExecutor) Execute(env *VirtualEnvironment, command *VirtualCommand, actions []string, params *Parameters) error {
  version, err := os.ReadFile("version")
  if err != nil {
    return InternalError("Unable to read version file", err)
  }

  Out(StdOut).Println("aux4 version", strings.TrimSpace(string(version)))
  
  return nil
}
