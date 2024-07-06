package main

import (
	"fmt"
	"strings"
	"time"
)

type Aux4VersionExecutor struct {
}

func (executor *Aux4VersionExecutor) GetCommandLine() string {
  return "version"
}

func (executor *Aux4VersionExecutor) Execute(env *VirtualEnvironment, command *VirtualCommand, actions []string, params *Parameters) error {
	year := time.Now().Year()

	Out(StdOut).Println()
	Out(StdOut).Println("  ", Cyan("aux4"), Yellow(Version))
	Out(StdOut).Println("  ", Gray(year, " aux4. aux4 is created and maintained by aux4 community."))
	Out(StdOut).Println("  ", Gray("https://aux4.io"))
	Out(StdOut).Println()

	latest := GetLatestRelease()
	if latest != "" && latest != Version {
		Out(StdOut).Println("  ", "Latest version:", Yellow(latest))
		Out(StdOut).Println()
	}

	return nil
}

type Aux4PkgerInstallExecutor struct {
}

func (executor *Aux4PkgerInstallExecutor) GetCommandLine() string {
  return "install <package>"
}

func (executor *Aux4PkgerInstallExecutor) Execute(env *VirtualEnvironment, command *VirtualCommand, actions []string, params *Parameters) error {
  var owner, name, version, err = getPackage(command, actions, params)
  if err != nil {
    return err
  }

  Out(StdOut).Println("Installing package", Yellow(fmt.Sprintf("%s/%s", owner, name)), "version", Yellow(version))

  var pkger = &Pkger{}
  err = pkger.Install(owner, name, version)
  if err != nil {
    Out(StdErr).Println(err)
  }

  return nil
}

type Aux4PkgerUninstallExecutor struct {
}

func (executor *Aux4PkgerUninstallExecutor) GetCommandLine() string {
  return "uninstall <package>"
}

func (executor *Aux4PkgerUninstallExecutor) Execute(env *VirtualEnvironment, command *VirtualCommand, actions []string, params *Parameters) error {
  var owner, name, version, err = getPackage(command, actions, params)
  if err != nil {
    return err
  }

  Out(StdOut).Println("Uninstalling package", Yellow(fmt.Sprintf("%s/%s", owner, name)), "version", Yellow(version))

  return nil
}

func getPackage(command *VirtualCommand, actions []string, params *Parameters) (string, string, string, error) {
  var pkg, err = params.Get(command, actions, "package")
  if err != nil {
    return "", "", "", err
  }

  if pkg == nil {
    return "", "", "", InternalError("Package name is required", nil)
  }

  var pkgParts = strings.Split(pkg.(string), "/")
  if len(pkgParts) != 2 {
    return "", "", "", InternalError("Invalid package name", nil)
  }
  var owner = pkgParts[0]
  
  var nameParts = strings.Split(pkgParts[1], "@")
  name := nameParts[0]
  var version = "latest"
  if len(nameParts) > 1 {
    version = nameParts[1]
  }

  return owner, name, version, nil
}
