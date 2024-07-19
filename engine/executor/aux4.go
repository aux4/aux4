package executor

import (
	"aux4/aux4"
	"aux4/core"
	"aux4/engine"
	"aux4/engine/param"
	"aux4/output"
	"aux4/pkger"
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

type Aux4PkgerListPackagesExecutor struct {
}

func (executor *Aux4PkgerListPackagesExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	var pkger = &pkger.Pkger{}
	packages, dependencies, err := pkger.ListInstalledPackages()
	if err != nil {
		return err
	}

	if len(packages) > 0 {
		output.Out(output.StdOut).Println(output.Gray("Installed packages:"))

		for _, pack := range packages {
			output.Out(output.StdOut).Println(output.Green(" ✓"), output.Yellow(pack.Scope, "/", output.Bold(pack.Name)), output.Gray(pack.Version))
		}
	}

	showDependencies := params.JustGet("show-dependencies")
	if len(dependencies) > 0 && (showDependencies == true || showDependencies == "true") {
		output.Out(output.StdOut).Println(output.Gray("Installed dependencies:"))

		for _, pack := range dependencies {
			output.Out(output.StdOut).Println(" ", output.Cyan("↪"), output.Magenta(pack.Scope, "/", output.Bold(pack.Name)), output.Gray(pack.Version))
		}
	}

	return nil
}

type Aux4PkgerBuildPackageExecutor struct {
}

func (executor *Aux4PkgerBuildPackageExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	pkger := &pkger.Pkger{}
	err := pkger.Build(actions)
	if err != nil {
		return err
	}

	return nil
}

type Aux4PkgerInstallExecutor struct {
}

func (executor *Aux4PkgerInstallExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	var pkger = &pkger.Pkger{}

	fromFile := params.JustGet("from-file")
	if fromFile != nil {
		installedPackages, err := pkger.InstallFromFile(fromFile.(string))
		if err != nil {
			return err
		}

		if len(installedPackages) == 0 {
			return nil
		}

		printInstalledPackages(installedPackages)

		return nil
	}

	scope, name, version, err := getPackage(command, actions, params)
	if err != nil {
		return err
	}

	installedPackages, err := pkger.Install(scope, name, version)
	if err != nil {
		return err
	}

	printInstalledPackages(installedPackages)

	return nil
}

type Aux4PkgerUninstallExecutor struct {
}

func (executor *Aux4PkgerUninstallExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	var scope, name, _, err = getPackage(command, actions, params)
	if err != nil {
		return err
	}

	var pkger = &pkger.Pkger{}
  uninstalledPackages, err := pkger.Uninstall(scope, name)
	if err != nil {
		return err
	}

  printUninstalledPackages(uninstalledPackages)

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
	var scope = pkgParts[0]

	var nameParts = strings.Split(pkgParts[1], "@")
	name := nameParts[0]
	var version = "latest"
	if len(nameParts) > 1 {
		version = nameParts[1]
	}

	return scope, name, version, nil
}

func printInstalledPackages(installedPackages []pkger.Package) {
	output.Out(output.StdOut).Println(output.Gray("Installed packages:"))

	for _, pack := range installedPackages {
		output.Out(output.StdOut).Println(output.Green(" ✓"), output.Yellow(pack.Scope, "/", output.Bold(pack.Name)), output.Magenta(pack.Version))
	}
}

func printUninstalledPackages(uninstalledPackages []pkger.Package) {
	output.Out(output.StdOut).Println(output.Gray("Uninstalled packages:"))

	for _, pack := range uninstalledPackages {
		output.Out(output.StdOut).Println(output.Red(" x"), output.Yellow(pack.Scope, "/", output.Bold(pack.Name)), output.Magenta(pack.Version))
	}
}
