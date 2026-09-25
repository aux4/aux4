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
	key := packageKey(pack)

	_, ok := library.packages[key]
	if ok {
		return core.InternalError(fmt.Sprintf("Package %s already exists", key), nil)
	}

	library.orderedPackages = append(library.orderedPackages, key)
	library.packages[key] = &pack

	return nil
}

func (library *Library) GetPackage(name string) (*core.Package, bool) {
	pack, ok := library.packages[name]
	return pack, ok
}

// packageKey returns the identity a package is stored and looked up under in
// the Library. Two packages published under different scopes (e.g.
// aux4/browser and agent/browser) share the same bare Name, so keying by
// Name alone collides ("Package browser already exists") the moment both
// are loaded into one Library -- which happens whenever pkger rebuilds
// global.aux4 from every installed package's own .aux4 file (install,
// uninstall, verify). Keying by "scope/name" instead makes the two
// distinct. Packages with no scope (the aux4 core builtins, and the merged
// global.aux4 blob itself, which carries scope="" and is keyed by its file
// path) keep using the bare Name/path, since those are already unique on
// their own and scope-prefixing them would just be noise.
func packageKey(pack core.Package) string {
	if pack.Scope != "" {
		return pack.Scope + "/" + pack.Name
	}
	return pack.Name
}
