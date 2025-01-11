package pkger

import (
	"aux4.dev/aux4/aux4"
	"aux4.dev/aux4/cloud"
	"aux4.dev/aux4/core"
	aux4IO "aux4.dev/aux4/io"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

const (
	registryUrl = "https://dev.api.hub.aux4.io/v1/packages/public"
)

func getPackageSpec(scope string, name string, version string) (Package, error) {
	specUrl := fmt.Sprintf("%s/%s/%s/%s", registryUrl, scope, name, version)

	client := &http.Client{}

	request, err := http.NewRequest("GET", specUrl, nil)
	if err != nil {
		return Package{}, core.InternalError(fmt.Sprintf("Error getting package spec %s/%s", scope, name), err)
	}

	request.Header.Set("User-Agent", aux4.GetUserAgent())

	response, err := client.Do(request)
	if err != nil {
		return Package{}, core.InternalError(fmt.Sprintf("Error getting package spec %s/%s", scope, name), err)
	}

	if response.StatusCode == 404 {
		return Package{}, PackageNotFoundError(scope, name, version)
	} else if response.StatusCode == 409 {
		return Package{}, core.InternalError(fmt.Sprintf("The package %s/%s is not compatible with your platform", scope, name), nil)
	} else if response.StatusCode == 426 {
		return Package{}, core.InternalError("Please upgrade aux4 before installing this package", nil)
	} else if response.StatusCode != 200 {
		return Package{}, core.InternalError(fmt.Sprintf("Error getting package spec %s/%s", scope, name), nil)
	}

	defer response.Body.Close()

	spec := Package{}
	err = json.NewDecoder(response.Body).Decode(&spec)
	if err != nil {
		return Package{}, core.InternalError(fmt.Sprintf("Error parsing package spec %s/%s", scope, name), err)
	}

	return spec, nil
}

func publish(file string) (core.Package, error) {
  aux4File, err := aux4IO.GetFileFromZip(file, ".aux4")
  if err != nil {
    return core.Package{}, core.InternalError(fmt.Sprintf("Error reading package spec %s", file), err)
  }

  spec := core.Package{}
  err = json.NewDecoder(aux4File).Decode(&spec)
  if err != nil {
    return spec, core.InternalError(fmt.Sprintf("Error parsing package spec %s", file), err)
  }

  err = uploadPackage(spec, file)

  return spec, err
}

func uploadPackage(spec core.Package, file string) error {
	client := &http.Client{}

  specUrl := fmt.Sprintf("%s/%s/%s/%s", registryUrl, spec.Scope, spec.Name, spec.Version)

  encodedBody, contentType, err := createMultipartRequestBody(file)
  if err != nil {
    return err
  }

  encodedBodyReader := bytes.NewReader(encodedBody)

  request, err := http.NewRequest("POST", specUrl, encodedBodyReader)
	if err != nil {
    return core.InternalError(fmt.Sprintf("Error publishing package %s: %s", file, err.Error()), err)
	}

  session, err := cloud.GetSession()
  if err != nil {
    return core.InternalError("not logged in", err)
  }

  request.Header.Set("Authorization", "Bearer " + session.AccessToken)
	request.Header.Set("User-Agent", aux4.GetUserAgent())
  request.Header.Set("Content-Type", contentType)

	response, err := client.Do(request)

	if err != nil {
    return core.InternalError(fmt.Sprintf("Error publishing package %s: %s", file, err.Error()), err)
	}

  if response.StatusCode != 200 {
    responseError := ParseHttpResponseError(response)
    return core.InternalError(fmt.Sprintf("Error publishing package %s: %s", file, responseError.Message), nil)
  }

  return nil
}

func createMultipartRequestBody(file string) ([]byte, string, error) {
  fileName := filepath.Base(file)

  var body bytes.Buffer

  multipartWriter := multipart.NewWriter(&body)

  fileWriter, err := multipartWriter.CreateFormFile("file", fileName)
  if err != nil {
    return []byte{}, "", core.InternalError(fmt.Sprintf("Error publishing package %s: %s", file, err.Error()), err)
  } 

  fileReader, err := os.Open(file)
  if err != nil {
    return []byte{}, "", core.InternalError(fmt.Sprintf("Error publishing package %s: %s", file, err.Error()), err)
  }

  _, err = io.Copy(fileWriter, fileReader)
  if err != nil {
    return []byte{}, "", core.InternalError(fmt.Sprintf("Error publishing package %s: %s", file, err.Error()), err)
  }

  multipartWriter.Close()

  encodedBase64Body := base64.StdEncoding.EncodeToString(body.Bytes())
  encodedBodyReader := []byte(encodedBase64Body)

  contentType := multipartWriter.FormDataContentType()

  return encodedBodyReader, contentType, nil
}
