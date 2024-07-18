package pkger

import (
	"aux4/core"
	"aux4/output"
)

type Pkger struct {
}

func (pkger *Pkger) ListInstalledPackages() ([]Package, []Package, error) {
	packageManager, err := InitPackageManager()
	if err != nil {
		return nil, nil, err
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

	return packages, dependencies, nil
}

func (pkger *Pkger) Build(files []string) error {
  err := build(files)
  if err != nil {
    return err
  }

  return nil
}

func (pkger *Pkger) Install(scope string, name string, version string) error {
	spec, err := getPackageSpec(scope, name, version)
	if err != nil {
		return err
	}

  return installFromSpec(spec)
}

func (pkger *Pkger) InstallFromFile(filepath string) error {
  spec, err := getPackageSpecFromFile(filepath)
  if err != nil {
    return err
  } 

  output.Out(output.StdOut).Println("Installing package", spec.Scope, spec.Name, spec.Version, spec.Url)

  return installFromSpec(spec)
}

func (pkger *Pkger) Uninstall(scope string, name string) error {
	packageManager, err := InitPackageManager()
	if err != nil {
		return err
	}

	packagesToRemove, err := packageManager.Remove(scope, name)
	if err != nil {
		return err
	}

	err = packageManager.Save()
	if err != nil {
		return err
	}

	if len(packagesToRemove) == 0 {
		return nil
	}

	err = uninstallPackages(packagesToRemove)
	if err != nil {
		return err
	}

	err = reloadGlobalPackages(packageManager)
	if err != nil {
		return err
	}

	return nil
}

func installFromSpec(spec Package) error {
  if spec.Scope == "" {
    return core.InternalError("scope is not defined in the package", nil)
  }

  if spec.Name == "" {
    return core.InternalError("name is not defined in the package", nil)
  }

  if spec.Version == "" {
    return core.InternalError("version is not defined in the package", nil)
  }

  if spec.Url == "" {
    return core.InternalError("url is not defined in the package", nil)
  }

	packageManager, err := InitPackageManager()
	if err != nil {
		return err
	}

	packagesToInstall, err := packageManager.Add(spec)
	if err != nil {
		return err
	}

	err = packageManager.Save()
	if err != nil {
		return err
	}

	err = installPackages(packagesToInstall)
	if err != nil {
		return err
	}

	return nil
}

