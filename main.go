package main

import (
	"os"
	"os/signal"
	"path/filepath"
)

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		Out(StdErr).Println("Process aborted")
		os.Exit(130)
	}()

	library := LocalLibrary()
	library.RegisterExecutor("aux4.version", &Aux4VersionExecutor{})

	if err := library.Load("", "aux4", []byte(`
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
              "name": "version",
              "help": {
                "text": "Display the version of aux4"
              }
            },
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

	var aux4Files []string
	listAux4Files(".", &aux4Files)

	for _, aux4File := range aux4Files {
		if err := library.LoadFile(aux4File); err != nil {
			Out(StdErr).Println(Red("Error loading file"), Red(aux4File), Red(err))
		}
	}

	env, err := InitializeVirtualEnvironment(library)
	if err != nil {
		Out(StdErr).Println(err)
		os.Exit(err.(Aux4Error).ExitCode)
	}

	actions, params := ParseArgs(os.Args[1:])

	if err := env.Execute(actions, &params); err != nil {
    Out(StdErr).Println(Red(err))
		os.Exit(err.(Aux4Error).ExitCode)
	}
}

func listAux4Files(path string, list *[]string) {
	dir, err := filepath.Abs(path)
	if err != nil {
		dir = path
	}

	aux4File := filepath.Join(dir, ".aux4")
	if _, err := os.Stat(aux4File); err == nil {
		*list = append([]string{aux4File}, *list...)
	}

	parent := filepath.Dir(dir)
	if parent != dir {
		listAux4Files(parent, list)
	}
}
