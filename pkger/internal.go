package pkger

import (
	"aux4/aux4"
	"aux4/config"
	"aux4/core"
	"aux4/engine"
	"aux4/io"
	"aux4/output"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var REPO_URL = "https://dev.api.hub.aux4.io/v1/packages/public"

func getPackageSpecFromFile(path string) (Package, error) {
	specFilepath, err := filepath.Abs(path)
	if err != nil {
		return Package{}, core.InternalError("Error getting package spec file path", err)
	}

	reader, err := io.GetFileFromZip(specFilepath, ".aux4")
	if err != nil {
		return Package{}, core.InternalError("Error reading package spec", err)
	}

	spec := Package{}
	err = json.NewDecoder(reader).Decode(&spec)
	if err != nil {
		return Package{}, core.InternalError("Error parsing package spec", err)
	}

	spec.Url = fmt.Sprintf("file://%s", specFilepath)

	return spec, nil
}

func getPackageSpec(scope string, name string, version string) (Package, error) {
	specUrl := fmt.Sprintf("%s/%s/%s/%s", REPO_URL, scope, name, version)

  client := &http.Client{}

  request, err := http.NewRequest("GET", specUrl, nil)
  if err != nil {
    return Package{}, core.InternalError(fmt.Sprintf("Error getting package spec %s/%s", scope, name), err)
  }

  request.Header.Set("User-Agent", getUserAgent())

	response, err := client.Do(request)
	if err != nil {
		return Package{}, core.InternalError(fmt.Sprintf("Error getting package spec %s/%s", scope, name), err)
	}

	if response.StatusCode == 404 {
		return Package{}, PackageNotFoundError(scope, name, version)
  } else if response.StatusCode == 409 {
    return Package{}, core.InternalError(fmt.Sprintf("The package %s/%s is not compatible with your platform", scope, name), nil)
  } else if response.StatusCode == 426 {
    return Package{}, core.InternalError("Please upgrade aux4 before installing this package", nil)
	} else if response.StatusCode != 200 {
		return Package{}, core.InternalError(fmt.Sprintf("Error getting package spec %s/%s", scope, name), nil)
	}

	defer response.Body.Close()

	spec := Package{}
	err = json.NewDecoder(response.Body).Decode(&spec)
	if err != nil {
		return Package{}, core.InternalError(fmt.Sprintf("Error parsing package spec %s/%s", scope, name), err)
	}

	return spec, nil
}

func installPackages(packages []Package) error {
	temporaryDirectory, err := io.GetTemporaryDirectory("aux4-install")
	if err != nil {
		return core.InternalError("Error creating temporary directory", err)
	}

	var packageFolder = config.GetConfigPath("packages")
	err = os.MkdirAll(packageFolder, 0755)
	if err != nil {
		return core.InternalError("Error creating package directory", err)
	}

	var library = engine.LocalLibrary()

	var globalAux4 = config.GetAux4GlobalPath()
	if _, err := os.Stat(globalAux4); err == nil {
		err = library.LoadFile(config.GetAux4GlobalPath())
		if err != nil {
			return err
		}
	}

	for _, pack := range packages {
		var packageFileDownloadPath string
		var downloaded bool

		if strings.HasPrefix(pack.Url, "file://") {
			packageFileDownloadPath = strings.TrimPrefix(pack.Url, "file://")
			downloaded = false
		} else {
			output.Out(output.StdOut).Println("Downloading package", pack.Scope, pack.Name, pack.Version)

			var packageFile = fmt.Sprintf("%s_%s_%s.zip", pack.Scope, pack.Name, pack.Version)
			packageFileDownloadPath = filepath.Join(temporaryDirectory, packageFile)

			err = io.DownloadFile(pack.Url, packageFileDownloadPath)
			if err != nil {
				return core.InternalError(fmt.Sprintf("Error downloading package %s/%s", pack.Scope, pack.Name), err)
			}

			downloaded = true
		}

		output.Out(output.StdOut).Println("Unzipping package", pack.Scope, pack.Name, pack.Version)

		err = io.UnzipFile(packageFileDownloadPath, packageFolder)
		if err != nil {
			return core.InternalError(fmt.Sprintf("Error unzipping package %s/%s", pack.Scope, pack.Name), err)
		}

		output.Out(output.StdOut).Println("Loading package", pack.Scope, pack.Name, pack.Version)

		err = library.LoadFile(filepath.Join(packageFolder, pack.Scope, pack.Name, ".aux4"))
		if err != nil {
			return err
		}

		if downloaded {
			os.RemoveAll(packageFileDownloadPath)
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
		packagePath := filepath.Join(config.GetConfigPath("packages"), pack.Scope, pack.Name)
		err := os.RemoveAll(packagePath)
		if err != nil {
			return core.InternalError(fmt.Sprintf("Error removing package %s/%s", pack.Scope, pack.Name), err)
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
			return core.InternalError(fmt.Sprintf("Error loading dependency %s/%s", pack.Scope, pack.Name), err)
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
			return core.InternalError(fmt.Sprintf("Error loading package %s/%s", pack.Scope, pack.Name), err)
		}
	}

	registry := engine.VirtualExecutorRegisty{}

	env, err := engine.InitializeVirtualEnvironment(library, &registry)
	if err != nil {
		return core.InternalError("Error initializing virtual environment", err)
	}

	err = env.Save(config.GetAux4GlobalPath())
	if err != nil {
		return core.InternalError("Error saving global environment", err)
	}

	return nil
}

func getUserAgent() string {
  return fmt.Sprintf("aux4/%s (%s; %s)", aux4.Version, runtime.GOOS, runtime.GOARCH)
}
