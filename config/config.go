package config

import (
	"aux4.dev/aux4/engine/param"
	"os"
	"path/filepath"
)

func GetAux4HomeDirectory() string {
	return filepath.Join(os.Getenv("HOME"), ".aux4.config")
}

func GetConfigPath(path string) string {
	return filepath.Join(GetAux4HomeDirectory(), path)
}

func GetAux4GlobalPath() string {
	return filepath.Join(GetAux4HomeDirectory(), "global.aux4")
}

func ListAux4Files(path string, aux4Params param.Aux4Parameters) []string {
	var aux4Files []string

	listAux4Files(path, &aux4Files, !aux4Params.Local())

	if !aux4Params.NoPackages() {
		var globalAux4 = GetAux4GlobalPath()
		if _, err := os.Stat(globalAux4); err == nil {
			aux4Files = append([]string{globalAux4}, aux4Files...)
		}
	}

	// AUX4_AUX4_FILES points at extra .aux4 files (os.PathListSeparator-separated) to load
	// on top of the discovered ones. It exists so tooling can contribute definitions —
	// notably aux4/mock's command stubs, which register `replace` hooks — without editing
	// the user's own .aux4 or depending on the working directory. Loaded last so those
	// definitions are the most specific; missing paths are skipped silently.
	if extra := os.Getenv("AUX4_AUX4_FILES"); extra != "" {
		for _, extraFile := range filepath.SplitList(extra) {
			if extraFile == "" {
				continue
			}
			if _, err := os.Stat(extraFile); err == nil {
				aux4Files = append(aux4Files, extraFile)
			}
		}
	}

	return aux4Files
}

func listAux4Files(path string, list *[]string, recursive bool) {
	dir, err := filepath.Abs(path)
	if err != nil {
		dir = path
	}

	aux4File := filepath.Join(dir, ".aux4")
	if _, err := os.Stat(aux4File); err == nil {
		*list = append([]string{aux4File}, *list...)
	}

	if !recursive {
		return
	}

	parent := filepath.Dir(dir)
	if parent != dir {
		listAux4Files(parent, list, true)
	}
}
