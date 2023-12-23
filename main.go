package main

import (
  "os"
)

func main() {
  library := LocalLibrary() 
  if err := library.LoadFile(".aux4"); err != nil {
    Out(StdErr).Println(err) 
    os.Exit(err.(Aux4Error).ExitCode)
  }

  env, err := InitializeVirtualEnvironment(library)
  if err != nil {
    Out(StdErr).Println(err) 
    os.Exit(err.(Aux4Error).ExitCode)
  }

  actions, params := ParseArgs(os.Args[1:])

  if err := env.Execute(actions, &params); err != nil {
    Out(StdErr).Println(err) 
    os.Exit(err.(Aux4Error).ExitCode)
  }
}
