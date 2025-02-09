package engine

import (
	"aux4.dev/aux4/core"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func LocalLibrary() *Library {
	return &Library{
		packages:  make(map[string]*core.Package),
	}
}

type Library struct {
	orderedPackages []string
	packages        map[string]*core.Package
}

func (library *Library) LoadFile(filename string) error {
	path, err := filepath.Abs(filename)
  if err != nil {
    return core.InternalError("Error loading aux4 file: " + filename, err)
  }

  file, err := os.Open(path)
	if err != nil {
    return core.InternalError("Error reading aux4 file: " + path, err)
	}

  var pack core.Package
	err = json.NewDecoder(file).Decode(&pack)
	if err != nil {
    return core.InternalError("Error parsing aux4 file: " + path, err)
	}

	pack.Path = path

  if pack.Name == "" {
    pack.Name = path
  }

	return library.load(pack)
}

func (library *Library) Load(path string, name string, data []byte) error {
	var pack core.Package

	err := json.Unmarshal(data, &pack)
	if err != nil {
    return core.InternalError("Error parsing aux4 file: " + path, err)
	}

	pack.Path = path

  if pack.Name == "" {
    pack.Name = name
  }

  return library.load(pack)
}

func (library *Library) load(pack core.Package) error {
	_, ok := library.packages[pack.Name]
	if ok {
		return core.InternalError(fmt.Sprintf("Package %s already exists", pack.Name), nil)
	}

	library.orderedPackages = append(library.orderedPackages, pack.Name)
	library.packages[pack.Name] = &pack

	return nil
}

func (library *Library) GetPackage(name string) (*core.Package, bool) {
	pack, ok := library.packages[name]
	return pack, ok
}
