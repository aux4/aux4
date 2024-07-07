package pkger

import (
	"aux4/config"
	"aux4/io"
	"os"
)

type Package struct {
	Owner   string `json:"owner"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

type PackageManager struct {
  Packages []Package `json:"packages"`
}

func (packageManager *PackageManager) Add(owner string, name string, version string) {
  packageManager.Packages = append(packageManager.Packages, Package{Owner: owner, Name: name, Version: version})
}

func (packageManager *PackageManager) Remove(owner string, name string) {
  var newPackages []Package
  for _, pack := range packageManager.Packages {
    if pack.Owner == owner && pack.Name == name {
      continue
    }
    newPackages = append(newPackages, pack)
  }
  packageManager.Packages = newPackages
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
    return &PackageManager{}, nil
  }

  var packageManager PackageManager
  err := io.ReadJsonFile(configPath, &packageManager)
  if err != nil {
    return nil, err
  }

  return &packageManager, nil
}
