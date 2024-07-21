package pkger

import (
	"aux4/cmd"
	"aux4/core"
	"aux4/output"
)

func installSystems(systems []System) error {
  for _, system := range systems {
    output.Out(output.StdOut).Println("Installing", system)

    err := install(system)
    if err != nil {
      return err
    }
  }
  return nil
}

func uninstallSystems(systems []System) error {
  for _, system := range systems {
    output.Out(output.StdOut).Println("Uninstalling", system)

    err := uninstall(system)
    if err != nil {
      return err
    }
  }
  return nil
}

func install(system System) error {
  _, _, err := cmd.ExecuteCommandLine("aux4 aux4 pkger system " + system.PackageManager + " install " + system.Package)
  if err != nil {
    return core.InternalError("Error installing package "+system.Package, err)
  }
  return nil
}

func uninstall(system System) error {
  _, _, err := cmd.ExecuteCommandLine("aux4 aux4 pkger system " + system.PackageManager + " uninstall " + system.Package)
  if err != nil {
    return core.InternalError("Error installing package "+system.Package, err)
  }
  return nil
}
