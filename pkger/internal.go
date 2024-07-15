package pkger

import (
	"aux4/config"
	"aux4/engine"
	"aux4/io"
	"aux4/output"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

var REPO_URL = "https://rv8lme69bi.execute-api.us-east-1.amazonaws.com/dev/v1/packages/public"

func getPackageSpec(scope string, name string, version string) (Package, error) {
  specUrl := fmt.Sprintf("%s/%s/%s/%s",REPO_URL, scope, name, version)

  resp, err := http.Get(specUrl)
  if err != nil {
    return Package{}, err
  }

  defer resp.Body.Close()

  spec := Package{}
  err = json.NewDecoder(resp.Body).Decode(&spec)
  if err != nil {
    return Package{}, err
  }

  return spec, nil
}

func installPackages(packages []Package) error {
	temporaryDirectory, err := io.GetTemporaryDirectory("aux4-install")
	if err != nil {
		return err
	}

	var packageFolder = config.GetConfigPath("packages")
	os.MkdirAll(packageFolder, 0755)

	var library = engine.LocalLibrary()

	var globalAux4 = config.GetAux4GlobalPath()
	if _, err := os.Stat(globalAux4); err == nil {
		err = library.LoadFile(config.GetAux4GlobalPath())
		if err != nil {
			return err
		}
	}

	for _, pack := range packages {
    output.Out(output.StdOut).Println("Downloading package", pack.Scope, pack.Name, pack.Version)

		var packageFile = fmt.Sprintf("%s_%s_%s.zip", pack.Scope, pack.Name, pack.Version)
		var packageFileDownloadPath = filepath.Join(temporaryDirectory, packageFile)

		err = io.DownloadFile(pack.Url, packageFileDownloadPath)

		if err != nil {
			return err
		}

    output.Out(output.StdOut).Println("Unzipping package", pack.Scope, pack.Name, pack.Version)
    
		err = io.UnzipFile(packageFileDownloadPath, packageFolder)
		if err != nil {
			return err
		}

    output.Out(output.StdOut).Println("Loading package", pack.Scope, pack.Name, pack.Version)

		err = library.LoadFile(filepath.Join(packageFolder, pack.Scope, pack.Name, ".aux4"))
		if err != nil {
			return err
		}
	}

	registry := engine.VirtualExecutorRegisty{}

	env, err := engine.InitializeVirtualEnvironment(library, &registry)
	if err != nil {
		return err
	}

	err = env.Save(config.GetAux4GlobalPath())
	if err != nil {
		return err
	}

	return nil
}

func uninstallPackages(packages []Package) error {
  for _, pack := range packages {
    output.Out(output.StdOut).Println("Removing package", pack.Scope, pack.Name, pack.Version)

    packagePath := filepath.Join(config.GetConfigPath("packages"), pack.Scope, pack.Name)
    err := os.RemoveAll(packagePath)
    if err != nil {
      return err
    }
  }

  return nil
}

func reloadGlobalPackages(packageManager *PackageManager) error {
	var library = engine.LocalLibrary()

  packagesDirectory := config.GetConfigPath("packages")

  installedPackages := map[string]bool{}

	for _, dependency := range packageManager.Dependencies {
    pack := ParsePackage(dependency.Package)
    packagePath := filepath.Join(packagesDirectory, pack.Scope, pack.Name, ".aux4")

    output.Out(output.StdOut).Println("Loading dependency", pack.Scope, pack.Name, pack.Version)

    err := library.LoadFile(packagePath)
    if err != nil {
      return err
    }

    installedPackages[pack.String()] = true
  }

	for _, pack := range packageManager.Packages {
    if _, ok := installedPackages[pack.String()]; ok {
      continue
    }

		packagePath := filepath.Join(packagesDirectory, pack.Scope, pack.Name, ".aux4")

    output.Out(output.StdOut).Println("Loading package", pack.Scope, pack.Name, pack.Version)

    err := library.LoadFile(packagePath)
		if err != nil {
			return err
		}
	}

	registry := engine.VirtualExecutorRegisty{}

	env, err := engine.InitializeVirtualEnvironment(library, &registry)
	if err != nil {
		return err
	}

	err = env.Save(config.GetAux4GlobalPath())
	if err != nil {
		return err
	}

  return nil
}
