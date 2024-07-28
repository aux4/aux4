package pkger

import (
	"aux4/cmd"
	"aux4/core"
	"aux4/engine"
	"aux4/output"
)

func installSystems(env *engine.VirtualEnvironment, systems []System) error {
	for _, system := range systems {
		output.Out(output.StdOut).Println("Installing system package", system.Id)

		err := install(env, system)
		if err != nil {
			return err
		}
	}
	return nil
}

func uninstallSystems(env *engine.VirtualEnvironment, systems []System) error {
	for _, system := range systems {
		output.Out(output.StdOut).Println("Uninstalling system package", system.Id)

		err := uninstall(env, system)
		if err != nil {
			return err
		}
	}
	return nil
}

func install(env *engine.VirtualEnvironment, system System) error {
	availablePackageSystems := GetAvailablePackageSystems(env)

	for _, systemPackage := range system.Packages {
		if _, ok := availablePackageSystems[systemPackage.PackageManager]; ok {
			_, _, err := cmd.ExecuteCommandLine("aux4 aux4 pkger system " + systemPackage.PackageManager + " install " + systemPackage.Package)
			if err != nil {
				return core.InternalError("Error installing package "+systemPackage.Package, err)
			}
			return nil
		}
	}

	return core.InternalError("Cannot install "+system.Id+". No package manager available.", nil)
}

func uninstall(env *engine.VirtualEnvironment, system System) error {
  availablePackageSystems := GetAvailablePackageSystems(env)

  for _, systemPackage := range system.Packages {
    if _, ok := availablePackageSystems[systemPackage.PackageManager]; ok {
      _, _, err := cmd.ExecuteCommandLine("aux4 aux4 pkger system " + systemPackage.PackageManager + " uninstall " + systemPackage.Package)
      if err != nil {
        return core.InternalError("Error installing package "+systemPackage.Package, err)
      }
      return nil
    }
  }

  return core.InternalError("Cannot uninstall "+system.Id+". No package manager available.", nil)
}

func GetAvailablePackageSystems(env *engine.VirtualEnvironment) map[string]bool {
	availableSystems := map[string]bool{}

	commands := env.ListCommandsAvailable("aux4:pkger:system")
	for _, command := range commands {
		availableSystems[command] = true
	}

	return availableSystems
}
