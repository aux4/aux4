package pkger

import (
	"aux4/core"
	"aux4/engine"
)

type Pkger struct {
}

func (pkger *Pkger) ListInstalledPackages() ([]Package, []Package, []System, error) {
	packageManager, err := InitPackageManager()
	if err != nil {
		return nil, nil, nil, err
	}

	packages := []Package{}

	for _, pack := range packageManager.Packages {
		packages = append(packages, pack)
	}

	dependencies := []Package{}

	for _, dependency := range packageManager.Dependencies {
		pack := ParsePackage(dependency.Package)
		dependencies = append(dependencies, pack)
	}

	systemDependencies := []System{}

	for _, systemDependency := range packageManager.SystemDependencies {
		systemDependencies = append(systemDependencies, systemDependency)
	}

	return packages, dependencies, systemDependencies, nil
}

func (pkger *Pkger) Build(files []string) error {
	err := build(files)
	if err != nil {
		return err
	}

	return nil
}

func (pkger *Pkger) Install(env *engine.VirtualEnvironment, scope string, name string, version string) ([]Package, error) {
	spec, err := getPackageSpec(scope, name, version)
	if err != nil {
		return []Package{}, err
	}

	return installFromSpec(env, spec)
}

func (pkger *Pkger) InstallFromFile(env *engine.VirtualEnvironment, filepath string) ([]Package, error) {
	spec, err := getPackageSpecFromFile(filepath)
	if err != nil {
		return []Package{}, err
	}

	return installFromSpec(env, spec)
}

func (pkger *Pkger) Uninstall(env *engine.VirtualEnvironment, scope string, name string) ([]Package, error) {
	packageManager, err := InitPackageManager()
	if err != nil {
		return []Package{}, err
	}

	packagesToRemove, systemDependenciesToRemove, err := packageManager.Remove(scope, name)
	if err != nil {
		return []Package{}, err
	}

	if len(packagesToRemove) == 0 {
		return []Package{}, err
	}

	err = uninstallPackages(packagesToRemove)
	if err != nil {
		return []Package{}, err
	}

  err = uninstallSystems(env, systemDependenciesToRemove)
  if err != nil {
    return []Package{}, err
  }

	err = packageManager.Save()
	if err != nil {
		return []Package{}, err
	}

	err = reloadGlobalPackages(packageManager)
	if err != nil {
		return []Package{}, err
	}

	return packagesToRemove, nil
}

func installFromSpec(env *engine.VirtualEnvironment, spec Package) ([]Package, error) {
	if spec.Scope == "" {
		return []Package{}, core.InternalError("scope is not defined in the package", nil)
	}

	if spec.Name == "" {
		return []Package{}, core.InternalError("name is not defined in the package", nil)
	}

	if spec.Version == "" {
		return []Package{}, core.InternalError("version is not defined in the package", nil)
	}

	if spec.Url == "" {
		return []Package{}, core.InternalError("url is not defined in the package", nil)
	}

	packageManager, err := InitPackageManager()
	if err != nil {
		return []Package{}, err
	}

	packagesToInstall, systemDependenciesToInstall, err := packageManager.Add(spec)
	if err != nil {
		return []Package{}, err
	}

	err = installPackages(packagesToInstall)
	if err != nil {
		return []Package{}, err
	}

	err = installSystems(env, systemDependenciesToInstall)
	if err != nil {
		return []Package{}, err
	}

	err = packageManager.Save()
	if err != nil {
		return []Package{}, err
	}

	return packagesToInstall, nil
}
