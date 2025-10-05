package io

import (
	"archive/zip"
	"aux4.dev/aux4/core"
	"bytes"
	"encoding/json"
	"fmt"
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

type OrderedMap struct {
	keys   []string
	values map[string]interface{}
}

func NewOrderedMap() *OrderedMap {
	return &OrderedMap{
		keys:   make([]string, 0),
		values: make(map[string]interface{}),
	}
}

func (om *OrderedMap) Set(key string, value interface{}) {
	if om == nil {
		return
	}
	if om.values == nil {
		om.values = make(map[string]interface{})
	}
	if om.keys == nil {
		om.keys = make([]string, 0)
	}
	if _, exists := om.values[key]; !exists {
		om.keys = append(om.keys, key)
	}
	om.values[key] = value
}

func (om *OrderedMap) Get(key string) (interface{}, bool) {
	if om == nil {
		return nil, false
	}
	if om.values == nil {
		return nil, false
	}
	value, exists := om.values[key]
	return value, exists
}

func (om *OrderedMap) Keys() []string {
	if om == nil || om.keys == nil {
		return []string{}
	}
	return om.keys
}

func (om *OrderedMap) Values() map[string]interface{} {
	if om == nil || om.values == nil {
		return make(map[string]interface{})
	}
	return om.values
}

func (om *OrderedMap) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("{")
	for i, key := range om.keys {
		if i > 0 {
			buf.WriteString(",")
		}
		keyBytes, err := json.Marshal(key)
		if err != nil {
			return nil, err
		}
		valueBytes, err := json.Marshal(om.values[key])
		if err != nil {
			return nil, err
		}
		buf.Write(keyBytes)
		buf.WriteString(":")
		buf.Write(valueBytes)
	}
	buf.WriteString("}")
	return buf.Bytes(), nil
}

func (om *OrderedMap) UnmarshalJSON(data []byte) error {
	keys, objmap, err := parseJSONWithOrder(data)
	if err != nil {
		return err
	}

	om.keys = keys
	om.values = make(map[string]interface{}, len(objmap))

	for _, key := range om.keys {
		value, err := unmarshalWithOrderPreservation(*objmap[key])
		if err != nil {
			return err
		}
		om.values[key] = value
	}

	return nil
}

func parseJSONWithOrder(data []byte) ([]string, map[string]*json.RawMessage, error) {
	var keys []string
	objmap := make(map[string]*json.RawMessage)
	
	decoder := json.NewDecoder(bytes.NewReader(data))
	
	t, err := decoder.Token()
	if err != nil {
		return nil, nil, err
	}
	
	if delim, ok := t.(json.Delim); !ok || delim != '{' {
		return nil, nil, fmt.Errorf("expected object start")
	}
	
	for decoder.More() {
		keyToken, err := decoder.Token()
		if err != nil {
			return nil, nil, err
		}
		
		key, ok := keyToken.(string)
		if !ok {
			return nil, nil, fmt.Errorf("expected string key")
		}
		
		var rawValue json.RawMessage
		if err := decoder.Decode(&rawValue); err != nil {
			return nil, nil, err
		}
		
		keys = append(keys, key)
		objmap[key] = &rawValue
	}
	
	t, err = decoder.Token()
	if err != nil {
		return nil, nil, err
	}
	
	if delim, ok := t.(json.Delim); !ok || delim != '}' {
		return nil, nil, fmt.Errorf("expected object end")
	}
	
	return keys, objmap, nil
}

func unmarshalWithOrderPreservation(data json.RawMessage) (interface{}, error) {
	var temp interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return nil, err
	}

	if _, ok := temp.(map[string]interface{}); ok {
		orderedMap := NewOrderedMap()
		if err := orderedMap.UnmarshalJSON(data); err != nil {
			return nil, err
		}
		return orderedMap, nil
	}

	if tempSlice, ok := temp.([]interface{}); ok {
		for i, item := range tempSlice {
			if _, ok := item.(map[string]interface{}); ok {
				itemBytes, err := json.Marshal(item)
				if err != nil {
					continue
				}
				orderedItem, err := unmarshalWithOrderPreservation(itemBytes)
				if err != nil {
					continue
				}
				tempSlice[i] = orderedItem
			}
		}
		return tempSlice, nil
	}

	return temp, nil
}

func (om *OrderedMap) String() string {
	jsonBytes, err := om.MarshalJSON()
	if err != nil {
		return fmt.Sprintf("OrderedMap{error: %v}", err)
	}
	return string(jsonBytes)
}
