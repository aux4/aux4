package main

import (
	"aux4/aux4"
	"aux4/cmd"
  //	"aux4/config"
	"aux4/core"
	"aux4/engine"
	"aux4/engine/executor"
	"aux4/engine/param"
	"aux4/output"
	"os"
)

func main() {
	cmd.AbortOnCtrlC()

	library := engine.LocalLibrary()

	if err := library.Load("", "aux4", []byte(aux4.DefaultAux4())); err != nil {
		output.Out(output.StdErr).Println(err)
		os.Exit(err.(core.Aux4Error).ExitCode)
	}

//	var aux4Files = config.ListAux4Files(".")
//
//	for _, aux4File := range aux4Files {
//		if err := library.LoadFile(aux4File); err != nil {
//			output.Out(output.StdErr).Println(output.Red("Error loading file"), output.Red(aux4File), output.Red(err))
//		}
//	}

	registry := engine.CreateVirtualExecutorRegistry()
	registry.RegisterExecutor("aux4.version", &executor.Aux4VersionExecutor{})
	registry.RegisterExecutor("aux4:pkger.install", &executor.Aux4PkgerInstallExecutor{})
	registry.RegisterExecutor("aux4:pkger.uninstall", &executor.Aux4PkgerUninstallExecutor{})

	env, err := engine.InitializeVirtualEnvironment(library, registry)
	if err != nil {
		output.Out(output.StdErr).Println(err)
		os.Exit(err.(core.Aux4Error).ExitCode)
	}

	actions, params := param.ParseArgs(os.Args[1:])

	if err := executor.Execute(env, actions, &params); err != nil {
		if err, ok := err.(core.Aux4Error); ok {
			output.Out(output.StdErr).Println(err)
			os.Exit(err.ExitCode)
		} else {
			output.Out(output.StdErr).Println(err)
			os.Exit(1)
		}
	}
}
