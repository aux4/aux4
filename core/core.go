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
		return pack, InternalError("Error loading aux4 file "+path, err)
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
	Path         string     `json:"-"`
	Repository   string     `json:"repository,omitempty"`
	Scope        string     `json:"scope"`
	Name         string     `json:"name"`
	Version      string     `json:"version"`
	License      string     `json:"license"`
	Description  string     `json:"description"`
	Git          string     `json:"git"`
	Website      string     `json:"website"`
	Dependencies []string   `json:"dependencies"`
	System       [][]string `json:"system"`
	Platforms    []string   `json:"platforms"`
	Distribution []string   `json:"dist"`
	Tags         []string   `json:"tags"`
	Profiles     []Profile  `json:"profiles"`
	Hooks        []Hook     `json:"hooks,omitempty"`
	// Schemes declares the remote-file-path schemes (e.g. "cloud") this package
	// provides, mapping a scheme name to the get/put commands that implement it.
	// See SchemeProvider.
	Schemes map[string]SchemeProvider `json:"schemes,omitempty"`
}

// SchemeProvider is the pair of aux4 commands that implement a <scheme>://
// remote file path: Get fetches the remote file to a local path, Put writes a
// local file back to the remote path with optimistic-concurrency (CAS)
// semantics. Both are plain aux4 command lines (without the "aux4 " prefix),
// e.g. {"get": "cloud files get", "put": "cloud files upload"}.
//
//   - Get is run as: aux4 <Get> --uri <uri> --to <localFile>
//     and must print {"exists":true|false,"version":"<etag>"} on stdout. A
//     missing remote file is exists:false with exit code 0 (not an error).
//   - Put is run as: aux4 <Put> --uri <uri> --file <localFile> --ifMatch <etag|none>
//     ("none" means create-only, i.e. If-None-Match). Exit code 3 means a
//     conflict (the remote changed since it was read); any other non-zero exit
//     is a regular error.
type SchemeProvider struct {
	Get string `json:"get"`
	Put string `json:"put"`
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

func (profile *Profile) GetCommand(name string) (*Command, bool) {
	for _, command := range profile.Commands {
		if command.Name == name {
			return &command, true
		}
	}
	return nil, false
}

type Command struct {
	Name    string        `json:"name"`
	Execute []string      `json:"execute"`
	Help    *CommandHelp  `json:"help"`
	Private bool          `json:"private"`
	NoHooks bool          `json:"noHooks,omitempty"`
	Render  CommandRender `json:"render,omitempty"`
	Ref     CommandRef    `json:"ref"`
}

type Hook struct {
	Command string            `json:"command"`
	Order   int               `json:"order,omitempty"`
	Params  map[string]string `json:"params,omitempty"`
	Before  []string          `json:"before,omitempty"`
	Replace []string          `json:"replace,omitempty"`
	After   []string          `json:"after,omitempty"`
	Error   []string          `json:"error,omitempty"`
}

type CommandRender map[string]string

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
	Text       string                 `json:"text"`
	Variables  []*CommandHelpVariable `json:"variables"`
	HasMan     bool                   `json:"hasMan"`
	HasExample bool                   `json:"hasExample"`
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
	Name     string   `json:"name"`
	Text     string   `json:"text"`
	Default  *string  `json:"default"`
	Type     string   `json:"type,omitempty"`
	Arg      bool     `json:"arg"`
	Multiple bool     `json:"multiple"`
	Hide     bool     `json:"hide"`
	Encrypt  bool     `json:"encrypt"`
	Private  bool     `json:"private"`
	Env      string   `json:"env"`
	Options  []string `json:"options"`
	// File declares this variable as a remote-file-path slot: "read", "write" or
	// "readwrite". It is additive and empty by default, meaning "not a file
	// variable" — the value is passed through unchanged exactly as before.
	// When set, a value shaped like "<scheme>://..." (e.g. "cloud://my-pkg/a.txt")
	// is staged to/from a local temp file around the command's execution using
	// the scheme registered by an installed package (see core.SchemeProvider). A
	// value without "://" is still a plain local path and is never touched.
	File string `json:"file,omitempty"`
}

// VisibleVariables returns the variables a user should be shown. A private
// variable still resolves like any other -- it can be passed, has its default,
// and binds to env and config -- it is simply not advertised.
//
// This is not the same as hide, which masks a variable's value when prompting
// and leaves the variable itself listed in the help.
func (help *CommandHelp) VisibleVariables() []*CommandHelpVariable {
	if help == nil || help.Variables == nil {
		return nil
	}

	visible := make([]*CommandHelpVariable, 0, len(help.Variables))
	for _, variable := range help.Variables {
		if !variable.Private {
			visible = append(visible, variable)
		}
	}
	return visible
}
