package main

import (
  "os"
)

func main() {
  library := LocalLibrary() 

  if err := library.Load("aux4", []byte(`
    {
      "profiles": [
        {
          "name": "main",
          "commands": [
            {
              "name": "aux4",
              "execute": [
                "profile:aux4"
              ],
              "help": {
                "text": "aux4 utility"
              }
            }
          ]
        },
        {
          "name": "aux4",
          "commands": [
            {
              "name": "man",
              "execute": [
                "set:help=true",
                "profile:main"
              ],
              "help": {
                "text": "Display help for a command"
              }
            }
          ]
        }
      ]
    }
  `)); err != nil {
    Out(StdErr).Println(err) 
    os.Exit(err.(Aux4Error).ExitCode)
  }

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
