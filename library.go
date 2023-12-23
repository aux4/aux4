package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Package struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	Author      string    `json:"author"`
	Profiles    []Profile `json:"profiles"`
}

func (pack *Package) GetProfile(name string) (*Profile, bool) {
  for _, profile := range pack.Profiles {
    if profile.Name == name {
      return &profile, true
    }
  }
  return nil, false
}

type Profile struct {
	Name     string    `json:"name"`
	Commands []Command `json:"commands"`
}

type Command struct {
	Name    string      `json:"name"`
	Execute []string    `json:"execute"`
	Help    CommandHelp `json:"help"`
}

type CommandHelp struct {
	Text      string                `json:"text"`
	Variables []CommandHelpVariable `json:"variables"`
}

type CommandHelpVariable struct {
	Name    string   `json:"name"`
	Text    string   `json:"text"`
	Default string   `json:"default"`
	Options []string `json:"options"`
}

func LocalLibrary() *Library {
	return &Library{
		packages: make(map[string]*Package),
	}
}

type Library struct {
	packages map[string]*Package
}

func (library *Library) LoadFile(filename string) error {
	path, err := filepath.Abs(filename)

	file, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return library.Load(path, file)
}

func (library *Library) Load(name string, data []byte) error {
	var pack Package

	err := json.Unmarshal(data, &pack)
	if err != nil {
		return err
	}

	if pack.Name == "" {
		pack.Name = name
	}

	_, ok := library.packages[pack.Name]
	if ok {
		return InternalError(fmt.Sprintf("Package %s already exists", pack.Name), nil)
	}

	library.packages[pack.Name] = &pack

	return nil
}

func (library *Library) GetPackage(name string) (*Package, bool) {
	pack, ok := library.packages[name]
	return pack, ok
}
