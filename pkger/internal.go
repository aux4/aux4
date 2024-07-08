package pkger

import (
	"aux4/config"
	"aux4/engine"
	"aux4/io"
	"aux4/output"
	"fmt"
	"os"
	"path/filepath"
)

var REPO_URL = "/Users/davidsg/Public/repo"

func getPackageSpec(owner string, name string, version string) Package {
  specPath := filepath.Join(REPO_URL, "spec", fmt.Sprintf("%s/%s/%s.json", owner, name, version))
  spec := Package{}
  io.ReadJsonFile(specPath, &spec)
  return spec
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
    output.Out(output.StdOut).Println("Downloading package", pack.Owner, pack.Name, pack.Version)

		var packageFile = fmt.Sprintf("%s_%s_%s.zip", pack.Owner, pack.Name, pack.Version)
		var packageFileDownloadPath = filepath.Join(temporaryDirectory, packageFile)

		err = io.CopyFile(filepath.Join(REPO_URL, "dist", packageFile), packageFileDownloadPath)

		if err != nil {
			return err
		}

    output.Out(output.StdOut).Println("Unzipping package", pack.Owner, pack.Name, pack.Version)
    
		err = io.UnzipFile(packageFileDownloadPath, packageFolder)
		if err != nil {
			return err
		}

    output.Out(output.StdOut).Println("Loading package", pack.Owner, pack.Name, pack.Version)

		err = library.LoadFile(filepath.Join(packageFolder, pack.Owner, pack.Name, ".aux4"))
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
    output.Out(output.StdOut).Println("Removing package", pack.Owner, pack.Name, pack.Version)

    packagePath := filepath.Join(config.GetConfigPath("packages"), pack.Owner, pack.Name)
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
    packagePath := filepath.Join(packagesDirectory, pack.Owner, pack.Name, ".aux4")

    output.Out(output.StdOut).Println("Loading dependency", pack.Owner, pack.Name, pack.Version)

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

		packagePath := filepath.Join(packagesDirectory, pack.Owner, pack.Name, ".aux4")

    output.Out(output.StdOut).Println("Loading package", pack.Owner, pack.Name, pack.Version)

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
