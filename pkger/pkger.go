package pkger

import (
	"aux4/core"
	"aux4/engine"
	"aux4/io"
	"fmt"
	"path/filepath"
	"runtime"
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

func (pkger *Pkger) InstallFromFile(env *engine.VirtualEnvironment, packageFilePath string) ([]Package, error) {
	spec, err := getPackageSpecFromFile(packageFilePath)
	if err != nil {
		return []Package{}, err
	}

	if len(spec.Distribution) > 0 {
		spec, err = getPackageSpecFromDistribution(spec, packageFilePath)
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

func getPackageSpecFromDistribution(spec Package, packageFilePath string) (Package, error) {
	var platform string

	currentOS := runtime.GOOS
	currentArch := runtime.GOARCH
	currentPlatform := fmt.Sprintf("%s_%s", currentOS, currentArch)

	for _, dist := range spec.Distribution {
		if dist == currentPlatform {
			platform = currentPlatform
			break
		} else if dist == currentOS {
			platform = currentOS
		}
	}

	if platform == "" {
		return Package{}, core.InternalError("No distribution for current platform", nil)
	}

	tmpDir, err := io.GetTemporaryDirectory(fmt.Sprintf("aux4_%s_%s", spec.Scope, spec.Name))
	if err != nil {
		return Package{}, err
	}

	err = io.UnzipFile(packageFilePath, tmpDir)
	if err != nil {
		return Package{}, err
	}

	packagePlatformFileName := fmt.Sprintf("%s/%s/%s_%s_%s_%s.zip", spec.Scope, spec.Name, platform, spec.Scope, spec.Name, spec.Version)
  platformSpec, err := getPackageSpecFromFile(filepath.Join(tmpDir, packagePlatformFileName))
	if err != nil {
		return Package{}, err
	}

	return platformSpec, nil
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
