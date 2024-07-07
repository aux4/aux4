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

func (packageManager *PackageManager) Add(owner string, name string, version string, dependencies []string) []string {
	pack := Package{Owner: owner, Name: name, Version: version}
	packageManager.Packages[pack.String()] = pack

	packagesToBeInstalled := []string{pack.String()}

	for _, dependencyReference := range dependencies {
		dependencyPackage := ParsePackage(dependencyReference)

		existingDependency, exists := packageManager.Dependencies[dependencyPackage.String()]
		if !exists {
			existingDependency = Dependency{Package: dependencyPackage.String(), UsedBy: []string{}}
			packageManager.Dependencies[dependencyPackage.String()] = existingDependency

			packagesToBeInstalled = append(packagesToBeInstalled, dependencyPackage.String())
		}

		existingDependency.UsedBy = append(existingDependency.UsedBy, pack.String())
	}

	return packagesToBeInstalled
}

func (packageManager *PackageManager) Remove(owner string, name string) []string {
  pack := packageManager.Packages[owner + "/" + name]
  packageName := pack.String()

	dependenciesToRemove := []string{packageName}

	for _, dependency := range packageManager.Dependencies {
		usedBy := dependency.UsedBy
		for index, usedByPackage := range usedBy {
			if usedByPackage == packageName {
				dependenciesToRemove = append(dependenciesToRemove, dependency.Package)
				usedBy = append(usedBy[:index], usedBy[index+1:]...)
			}
		}
		dependency.UsedBy = usedBy
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

      if len(existingDependency.UsedBy) == 0 {
        dependenciesToRemove = append(dependenciesToRemove, existingDependency.Package)
        delete(packageManager.Dependencies, existingDependency.Package)
      }
    }
  }

	delete(packageManager.Packages, packageName)
	delete(packageManager.Dependencies, packageName)

	return dependenciesToRemove
}

func (packageManager *PackageManager) Save() error {
	configPath := config.GetConfigPath("packages/all.json")
	err := io.WriteJsonFile(configPath, packageManager)
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
