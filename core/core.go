package core

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func ReadPackage(path string) (Package, error) {
	var pack Package

	data, err := os.ReadFile(path)
	if err != nil {
		return pack, InternalError("Error loading aux4 file " + path, err)
	}

	err = json.Unmarshal(data, &pack)
	if err != nil {
		return pack, InternalError("Error parsing aux4 file", err)
	}

	return pack, nil
}

func WritePackage(path string, pack Package) error {
	data, err := json.Marshal(pack)
	if err != nil {
		return InternalError("Error marshalling package", err)
	}

	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return InternalError("Error writing package", err)
	}

	return nil
}

type Package struct {
	Path         string
	Scope        string    `json:"scope"`
	Name         string    `json:"name"`
	Dependencies []string  `json:"dependencies"`
	System       []string  `json:"system"`
	Platforms    []string  `json:"platforms"`
	Distribution []string  `json:"dist"`
	Version      string    `json:"version"`
	Description  string    `json:"description"`
	Author       string    `json:"author"`
	Profiles     []Profile `json:"profiles"`
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
	Name    string       `json:"name"`
	Execute []string     `json:"execute"`
	Help    *CommandHelp `json:"help"`
	Ref     CommandRef   `json:"ref"`
}

func (command *Command) SetRef(Path string, Package string, Profile string) {
	command.Ref = CommandRef{
		Path:    Path,
		Dir:     filepath.Dir(Path),
		Package: Package,
		Profile: Profile,
	}
}

type CommandRef struct {
	Path    string `json:"path"`
	Dir     string `json:"dir"`
	Package string `json:"package"`
	Profile string `json:"profile"`
}

type CommandHelp struct {
	Text      string                 `json:"text"`
	Variables []*CommandHelpVariable `json:"variables"`
}

func (help *CommandHelp) GetVariable(name string) (*CommandHelpVariable, bool) {
	if help == nil || help.Variables == nil {
		return nil, false
	}

	for _, variable := range help.Variables {
		if variable.Name == name {
			return variable, true
		}
	}
	return nil, false
}

type CommandHelpVariable struct {
	Name    string   `json:"name"`
	Text    string   `json:"text"`
	Default *string  `json:"default"`
	Arg     bool     `json:"arg"`
	Hide    bool     `json:"hide"`
	Encrypt bool     `json:"encrypt"`
	Env     string   `json:"env"`
	Options []string `json:"options"`
}
