package pkger

import (
	"archive/zip"
	"aux4/core"
	aux4IO "aux4/io"
	"aux4/output"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func build(paths []string) error {
	packageFiles := &[]string{}
	listAllFiles(paths, packageFiles)

	output.Out(output.StdOut).Println("Files", packageFiles)

	aux4Path := getAux4Path(packageFiles)
	if aux4Path == "" {
		return core.InternalError(".aux4 file not found", nil)
	}

	pack := Package{}
	err := aux4IO.ReadJsonFile(aux4Path, &pack)
	if err != nil {
		return core.InternalError("Error parsing .aux4 file", err)
	}

	output.Out(output.StdOut).Println("Building", aux4Path, pack.Scope, pack.Name, pack.Version)

	err = zipPackage(pack, packageFiles)

	return nil
}

func zipPackage(pack Package, packageFiles *[]string) error {
	zipFileName := fmt.Sprintf("%s_%s_%s.zip", pack.Scope, pack.Name, pack.Version)
  prefix := fmt.Sprintf("%s/%s", pack.Scope, pack.Name)

	zipFile, err := os.Create(zipFileName)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, filePath := range *packageFiles {
		err := addFileToZip(zipWriter, prefix, filePath)
		if err != nil {
			return err
		}
	}

	return nil
}

func addFileToZip(zipWriter *zip.Writer, prefix string, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		return err
	}
	header.Name = filepath.Join(prefix, filepath.Base(filePath))
	header.Method = zip.Deflate

  output.Out(output.StdOut).Println("Adding file", header.Name)

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	if err != nil {
		return err
	}

	return nil
}

func getAux4Path(paths *[]string) string {
	for _, path := range *paths {
		if strings.HasSuffix(path, ".aux4") {
			return path
		}
	}
	return ""
}

func listAllFiles(paths []string, allFiles *[]string) {
	for _, path := range paths {
		absolutePath, err := filepath.Abs(path)
		if err != nil {
			continue
		}

		stat, err := os.Stat(absolutePath)
		if err != nil {
			continue
		}

		if stat.IsDir() {
			files, err := os.ReadDir(absolutePath)
			if err != nil {
				continue
			}

			for _, file := range files {
				listAllFiles([]string{filepath.Join(absolutePath, file.Name())}, allFiles)
			}
		} else {
			*allFiles = append(*allFiles, absolutePath)
		}
	}
}
