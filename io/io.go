package io

import (
	"archive/zip"
	"aux4.dev/aux4/core"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func ReadJsonFile(path string, object any) error {
  file, err := os.ReadFile(path)
  if err != nil {
    return err
  }

  err = json.Unmarshal(file, object)
  if err != nil {
    return err
  }

  return nil
}

func WriteJsonFile(path string, object any) error {
  var content, err = json.Marshal(object)
	if err != nil {
		return err
	}

	err = os.WriteFile(path, content, 0644)
  if err != nil {
    return err
  }

	return nil
}

func GetTemporaryDirectory(path string) (string, error) {
	dir, err := os.MkdirTemp("", path)
	if err != nil {
		return "", err
	}
	return dir, nil
}

func DownloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

  if resp.StatusCode != http.StatusOK {
    return core.InternalError(resp.Status, nil)
  }

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func CopyFile(source string, target string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}

	destinationFile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	return err
}

func GetFileFromZip(zipPath, filename string) (*bytes.Reader, error) {
  reader, err := zip.OpenReader(zipPath)
  if err != nil {
    return nil, err
  }
  defer reader.Close()

  for _, file := range reader.File {
    if strings.HasSuffix(file.Name, filename) {
      fileContent, err := file.Open()
      if err != nil {
        return nil, err
      }

      destination := &bytes.Buffer{}
      _, err = io.Copy(destination, fileContent)
      if err != nil {
        return nil, err
      }

      content := destination.Bytes()

      fileReader := bytes.NewReader(content)
      return fileReader, nil
    }
  }

  return nil, os.ErrNotExist
}

func UnzipFile(zipPath, destDir string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	err = os.MkdirAll(destDir, 0755)
	if err != nil {
		return err
	}

	for _, file := range reader.File {
		err := unzipFileEntry(file, destDir)
		if err != nil {
			return err
		}
	}

	return nil
}

func unzipFileEntry(file *zip.File, destDir string) error {
	filePath := filepath.Join(destDir, file.Name)

	if file.FileInfo().IsDir() {
		err := os.MkdirAll(filePath, file.Mode())
		if err != nil {
			return err
		}
	} else {
		err := os.MkdirAll(filepath.Dir(filePath), 0755)
		if err != nil {
			return err
		}

		outputFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer outputFile.Close()

		sourceFile, err := file.Open()
		if err != nil {
			return err
		}
		defer sourceFile.Close()

		_, err = io.Copy(outputFile, sourceFile)
		if err != nil {
			return err
		}
	}

	return nil
}
