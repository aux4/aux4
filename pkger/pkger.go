package pkger

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

func (pkger *Pkger) Install(scope string, name string, version string) error {
	spec, err := getPackageSpec(scope, name, version)
	if err != nil {
		return err
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
