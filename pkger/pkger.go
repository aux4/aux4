package pkger

type Pkger struct {
}

func (pkger *Pkger) Install(owner string, name string, version string) error {
  spec := getPackageSpec(owner, name, version)

	packageManager, err := InitPackageManager()
	if err != nil {
		return err
	}

  packagesToInstall := packageManager.Add(spec.Owner, spec.Name, spec.Version, spec.Dependencies)

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

func (pkger *Pkger) Uninstall(owner string, name string) error {
	packageManager, err := InitPackageManager()
	if err != nil {
		return err
	}

  packagesToRemove := packageManager.Remove(owner, name)

	err = packageManager.Save()
	if err != nil {
		return err
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
