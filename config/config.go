package config

import (
	"os"
	"path/filepath"
)

func GetConfigDirectory() (string) {
	return filepath.Join(os.Getenv("HOME"), ".aux4.config")
}

func GetConfigPath(path string) (string) {
  return filepath.Join(GetConfigDirectory(), path)
}

func GetAux4GlobalPath() (string) {
  return filepath.Join(GetConfigDirectory(), "global.aux4")
}

func ListAux4Files(path string) []string {
  var aux4Files []string

  listAux4Files(path, &aux4Files)

	var globalAux4 = GetAux4GlobalPath()
	if _, err := os.Stat(globalAux4); err == nil {
		aux4Files = append([]string{globalAux4}, aux4Files...)
	}

  return aux4Files
}

func listAux4Files(path string, list *[]string) {
	dir, err := filepath.Abs(path)
	if err != nil {
		dir = path
	}

	aux4File := filepath.Join(dir, ".aux4")
	if _, err := os.Stat(aux4File); err == nil {
		*list = append([]string{aux4File}, *list...)
	}

	parent := filepath.Dir(dir)
	if parent != dir {
		listAux4Files(parent, list)
	}
}
