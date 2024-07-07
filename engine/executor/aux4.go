package executor

import (
	"aux4/aux4"
	"aux4/core"
	"aux4/engine"
	"aux4/engine/param"
	"aux4/output"
	"aux4/pkger"
	"fmt"
	"strings"
	"time"
)

type Aux4VersionExecutor struct {
}

func (executor *Aux4VersionExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	year := time.Now().Year()

	output.Out(output.StdOut).Println()
	output.Out(output.StdOut).Println("  ", output.Cyan("aux4"), output.Yellow(aux4.Version))
	output.Out(output.StdOut).Println("  ", output.Gray(year, " aux4. aux4 is created and maintained by aux4 community."))
	output.Out(output.StdOut).Println("  ", output.Gray("https://aux4.io"))
	output.Out(output.StdOut).Println()

	latest := aux4.GetLatestRelease()
	if latest != "" && latest != aux4.Version {
		output.Out(output.StdOut).Println("  ", "Latest version:", output.Yellow(latest))
		output.Out(output.StdOut).Println()
	}

	return nil
}

type Aux4PkgerInstallExecutor struct {
}

func (executor *Aux4PkgerInstallExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	var owner, name, version, err = getPackage(command, actions, params)
	if err != nil {
		return err
	}

	output.Out(output.StdOut).Println("Installing package", output.Yellow(fmt.Sprintf("%s/%s", owner, name)), "version", output.Yellow(version))

	var pkger = &pkger.Pkger{}
	err = pkger.Install(owner, name, version)
	if err != nil {
		output.Out(output.StdErr).Println(err)
	}

	return nil
}

type Aux4PkgerUninstallExecutor struct {
}

func (executor *Aux4PkgerUninstallExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	var owner, name, version, err = getPackage(command, actions, params)
	if err != nil {
		return err
	}

	output.Out(output.StdOut).Println("Uninstalling package", output.Yellow(fmt.Sprintf("%s/%s", owner, name)), "version", output.Yellow(version))

	return nil
}

func getPackage(command core.Command, actions []string, params *param.Parameters) (string, string, string, error) {
	var pkg, err = params.Get(command, actions, "package")
	if err != nil {
		return "", "", "", err
	}

	if pkg == nil {
		return "", "", "", core.InternalError("Package name is required", nil)
	}

	var pkgParts = strings.Split(pkg.(string), "/")
	if len(pkgParts) != 2 {
		return "", "", "", core.InternalError("Invalid package name", nil)
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
