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
  library.RegisterExecutor("aux4:pkger.install", &Aux4PkgerInstallExecutor{})
  library.RegisterExecutor("aux4:pkger.uninstall", &Aux4PkgerUninstallExecutor{})

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
              "name": "pkger",
              "execute": [
                "profile:aux4:pkger"
              ],
              "help": {
                "text": "Manage aux4 packages"
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
        },
        {
          "name": "aux4:pkger",
          "commands": [
            {
              "name": "install",
              "help": {
                "text": "Install a package",
                "variables": [
                  {
                    "name": "package",
                    "text": "the package to install",
                    "arg": true
                  }
                ]
              }
            },
            {
              "name": "uninstall",
              "help": {
                "text": "Uninstall a package",
                "variables": [
                  {
                    "name": "package",
                    "text": "the package to uninstall",
                    "arg": true
                  }
                ]
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

	var globalAux4 = GetAux4GlobalPath()
	if _, err := os.Stat(globalAux4); err == nil {
		aux4Files = append([]string{globalAux4}, aux4Files...)
	}

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
		if err, ok := err.(Aux4Error); ok {
			Out(StdErr).Println(err)
			os.Exit(err.ExitCode)
		} else {
			Out(StdErr).Println(err)
			os.Exit(1)
		}
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
