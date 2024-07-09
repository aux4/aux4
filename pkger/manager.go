package pkger

import (
	"aux4/config"
	"aux4/io"
	"os"
	"strings"
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
	ownerName := parts[0]
	name := parts[1]

	return Package{Owner: ownerName, Name: name, Version: version}
}

type Package struct {
	Owner        string   `json:"owner"`
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Dependencies []string `json:"dependencies"`
}

func (pack Package) String() string {
	return pack.Owner + "/" + pack.Name
}

type Dependency struct {
	Package string   `json:"package"`
	UsedBy  []string `json:"usedBy"`
}

type PackageManager struct {
	Packages     map[string]Package    `json:"packages"`
	Dependencies map[string]Dependency `json:"dependencies"`
}

func (packageManager *PackageManager) Add(owner string, name string, version string, dependencies []string) ([]Package, error) {
  _, exists := packageManager.Packages[owner + "/" + name]
  if exists {
    return []Package{}, PackageAlreadyInstalledError(owner, name)
  }

	pack := Package{Owner: owner, Name: name, Version: version, Dependencies: dependencies}
	packageManager.Packages[pack.String()] = pack

	packagesToBeInstalled := []Package{}

	for _, dependencyReference := range dependencies {
		dependencyPackage := ParsePackage(dependencyReference)

		existingDependency, exists := packageManager.Dependencies[dependencyPackage.String()]
		if !exists {
			existingDependency = Dependency{Package: dependencyReference, UsedBy: []string{}}
			packageManager.Dependencies[dependencyPackage.String()] = existingDependency

			_, existsAsPackage := packageManager.Packages[dependencyPackage.String()]
			if !existsAsPackage {
				packagesToBeInstalled = append(packagesToBeInstalled, dependencyPackage)
			}
		}

		existingDependency.UsedBy = append(existingDependency.UsedBy, pack.String())
		packageManager.Dependencies[dependencyPackage.String()] = existingDependency
	}
  
  packagesToBeInstalled = append(packagesToBeInstalled, pack)

	return packagesToBeInstalled, nil
}

func (packageManager *PackageManager) Remove(owner string, name string) ([]Package, error) {
	packageName := owner + "/" + name
  pack, exists := packageManager.Packages[packageName]

	dependenciesToRemove := []Package{}

  if exists {
    dependenciesToRemove = append(dependenciesToRemove, pack)
  }

	packageAsDependency, existsAsDependency := packageManager.Dependencies[packageName]
	if existsAsDependency && len(packageAsDependency.UsedBy) > 0 {
    delete(packageManager.Packages, packageName)

		return []Package{}, PackageHasDependenciesError(owner, name, packageAsDependency.UsedBy)
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
				delete(packageManager.Dependencies, dependencyPackage.String())

				_, existsAsPackage := packageManager.Packages[dependencyPackage.String()]
				if !existsAsPackage {
					dependenciesToRemove = append(dependenciesToRemove, dependencyPackage)
				}
			}
		}
	}

  if !exists {
    return []Package{}, PackageNotFoundError(owner, name)
  }

	delete(packageManager.Packages, packageName)
	delete(packageManager.Dependencies, packageName)

	return dependenciesToRemove, nil
}

func (packageManager *PackageManager) Save() error {
	packagesDirectory := config.GetConfigPath("packages")
	err := os.MkdirAll(packagesDirectory, os.ModePerm)

	configPath := config.GetConfigPath("packages/all.json")
	err = io.WriteJsonFile(configPath, packageManager)
	if err != nil {
		return err
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
		return nil, err
	}

	return &packageManager, nil
}
