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

// DefaultRepository is the hub repository a scoped package is assumed to come
// from when its .aux4 does not declare one.
const DefaultRepository = "public"

// packageKey returns the identity a package is stored and looked up under in
// the Library: "<repository>:<scope>/<name>" (e.g. "public:aux4/browser").
//
//   - Different scopes are distinct: aux4/browser and agent/browser share the
//     bare Name "browser" and both load (keying by Name alone made them collide
//     with "Package browser already exists").
//   - The same scope/name from different repositories is distinct
//     (public:aux4/x vs system:aux4/x). A package with no "repository" field is
//     treated as coming from the public repository.
//   - The version is deliberately NOT part of the key: two versions of the same
//     package from the same repository still collide, because only one version
//     of a package can be installed.
//   - Packages with no scope (the aux4 core builtins, a local .aux4 file, and the
//     merged global.aux4 blob, which is keyed by its file path) keep their bare
//     Name/path key, exactly as before.
//
// The key only identifies a package inside one in-memory Library. Load/merge
// order is the insertion order kept in orderedPackages and never depends on the
// key, and the key is never written to disk.
func packageKey(pack core.Package) string {
	if pack.Scope == "" {
		return pack.Name
	}

	repository := pack.Repository
	if repository == "" {
		repository = DefaultRepository
	}

	return repository + ":" + pack.Scope + "/" + pack.Name
}
