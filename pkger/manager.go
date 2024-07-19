package pkger

import (
	"aux4/config"
	"aux4/core"
	"aux4/io"
	"fmt"
	"os"
	"strings"

	"github.com/blang/semver/v4"
)

func ParsePackage(definition string) Package {
	var version string
	var packageName string

	parts := strings.Split(definition, "@")
	if len(parts) == 1 {
		packageName = parts[0]
		version = "latest"
	} else {
		packageName = parts[0]
		version = parts[1]
	}

	parts = strings.Split(packageName, "/")
	scope := parts[0]
	name := parts[1]

	return Package{Scope: scope, Name: name, Version: version}
}

type Package struct {
	Scope        string   `json:"scope"`
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Url          string   `json:"url"`
	Dependencies []string `json:"dependencies"`
	Dependency   bool
}

func (pack Package) String() string {
	return pack.Scope + "/" + pack.Name
}

func (pack Package) FullString() string {
	return pack.Scope + "/" + pack.Name + "@" + pack.Version
}

type Dependency struct {
	Package string   `json:"package"`
	UsedBy  []string `json:"usedBy"`
}

type PackageManager struct {
	Packages     map[string]Package    `json:"packages"`
	Dependencies map[string]Dependency `json:"dependencies"`
}

func (packageManager *PackageManager) Add(pack Package) ([]Package, error) {
	packagesToBeInstalled := []Package{}

	err := packageManager.add(pack, &packagesToBeInstalled)

	return packagesToBeInstalled, err
}

func (packageManager *PackageManager) add(pack Package, packagesToBeInstalled *[]Package) error {
	existingPackage, exists := packageManager.Packages[pack.Scope+"/"+pack.Name]
	if exists {
		if existingPackage.Version == pack.Version {
			return PackageAlreadyInstalledError(pack.Scope, pack.Name)
		} else {
			currentVersion, _ := semver.Parse(existingPackage.Version)
			newVersion, _ := semver.Parse(pack.Version)

			if currentVersion.GT(newVersion) {
				return core.InternalError(fmt.Sprintf("The version of %s/%s you are trying to install is older than the current version %s", pack.Scope, pack.Name, existingPackage.Version), nil)
			}
		}
	}

	if !pack.Dependency {
		packageManager.Packages[pack.String()] = pack
	}

	for _, dependencyReference := range pack.Dependencies {
		dependencyPackage := ParsePackage(dependencyReference)

		existingDependency, exists := packageManager.Dependencies[dependencyPackage.String()]
		if !exists {
			existingDependency = Dependency{Package: dependencyReference, UsedBy: []string{}}
			packageManager.Dependencies[dependencyPackage.String()] = existingDependency

			_, existsAsPackage := packageManager.Packages[dependencyPackage.String()]
			if !existsAsPackage {
				dependencySpec, err := getPackageSpec(dependencyPackage.Scope, dependencyPackage.Name, dependencyPackage.Version)
				if err != nil {
					return err
				}

				dependencySpec.Dependency = true
        existingDependency.Package = dependencySpec.FullString()
				err = packageManager.add(dependencySpec, packagesToBeInstalled)
				if err != nil {
					return err
				}
			}
		}

		existingDependency.UsedBy = append(existingDependency.UsedBy, pack.String())
		packageManager.Dependencies[dependencyPackage.String()] = existingDependency
	}

	*packagesToBeInstalled = append(*packagesToBeInstalled, pack)

	return nil
}

func (packageManager *PackageManager) Remove(scope string, name string) ([]Package, error) {
	packagesToRemove := []Package{}

	pack := Package{Scope: scope, Name: name}
	err := packageManager.remove(pack, &packagesToRemove)

	return packagesToRemove, err
}

func (packageManager *PackageManager) remove(pack Package, packagesToRemove *[]Package) error {
	packageName := pack.String()

	pack, exists := packageManager.Packages[packageName]
	if exists {
		*packagesToRemove = append(*packagesToRemove, pack)
	}

	packageAsDependency, existsAsDependency := packageManager.Dependencies[packageName]
	if existsAsDependency {
		if len(packageAsDependency.UsedBy) > 0 {
			delete(packageManager.Packages, packageName)

			return PackageHasDependenciesError(pack.Scope, pack.Name, packageAsDependency.UsedBy)
		} else {
      dependencyPackage := ParsePackage(packageAsDependency.Package)
      dependencyPackage.Dependency = true
      *packagesToRemove = append(*packagesToRemove, dependencyPackage)
		}
	}

	if len(pack.Dependencies) > 0 {
		for _, dependencyReference := range pack.Dependencies {
			dependencyPackage := ParsePackage(dependencyReference)

			existingDependency := packageManager.Dependencies[dependencyPackage.String()]
			usedBy := existingDependency.UsedBy
			for index, usedByPackage := range usedBy {
				if usedByPackage == packageName {
					usedBy = append(usedBy[:index], usedBy[index+1:]...)
				}
			}
			existingDependency.UsedBy = usedBy
			packageManager.Dependencies[dependencyPackage.String()] = existingDependency

			if len(existingDependency.UsedBy) == 0 {
				dependencyPackage := ParsePackage(existingDependency.Package)

				_, existsAsPackage := packageManager.Packages[dependencyPackage.String()]
				if !existsAsPackage {
					dependencyPackage.Dependency = true
					err := packageManager.remove(dependencyPackage, packagesToRemove)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	delete(packageManager.Packages, packageName)
	delete(packageManager.Dependencies, packageName)

	return nil
}

func (packageManager *PackageManager) Save() error {
	packagesDirectory := config.GetConfigPath("packages")
	err := os.MkdirAll(packagesDirectory, os.ModePerm)

	configPath := config.GetConfigPath("packages/all.json")
	err = io.WriteJsonFile(configPath, packageManager)
	if err != nil {
		return core.InternalError("Failed to save package manager configuration", err)
	}

	return nil
}

func InitPackageManager() (*PackageManager, error) {
	configPath := config.GetConfigPath("packages/all.json")

	if _, err := os.Stat(configPath); err != nil {
		return &PackageManager{
			Packages:     make(map[string]Package),
			Dependencies: make(map[string]Dependency),
		}, nil
	}

	var packageManager PackageManager
	err := io.ReadJsonFile(configPath, &packageManager)
	if err != nil {
		return nil, core.InternalError("Failed to read package manager configuration", err)
	}

	return &packageManager, nil
}
