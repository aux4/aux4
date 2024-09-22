package cloud

import (
	"aux4/aux4"
	"aux4/config"
	"aux4/core"
	aux4IO "aux4/io"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
)

const (
	api = "https://dev.api.user.aux4.io/v1"
)

func Login(email string, password string, otp string) (*Aux4Session, error) {
	var authentication = base64.StdEncoding.EncodeToString([]byte(email + ":" + password))

	client := &http.Client{}

  url := api + "/login"

	request, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("User-Agent", aux4.GetUserAgent())
	request.Header.Set("Authorization", "Basic "+authentication)

  aux4Login := Aux4Login{Otp: otp}
  aux4LoginJson, err := json.Marshal(aux4Login)
  if err != nil {
    return nil, core.InternalError("Error parsing login request", err)
  }

  request.Body = io.NopCloser(bytes.NewReader(aux4LoginJson))

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, core.InternalError("Login failed", nil)
	}

	defer response.Body.Close()

	session := Aux4Session{}
	err = json.NewDecoder(response.Body).Decode(&session)
	if err != nil {
		return nil, core.InternalError("Error parsing login response", err)
	}

	credentialsPath := config.GetConfigPath("credentials")

	err = aux4IO.WriteJsonFile(credentialsPath, session)
	if err != nil {
		return nil, core.InternalError("Error saving credentials", err)
	}

	return &session, nil
}

func GetSession() (*Aux4Session, error) {
  credentialsPath := config.GetConfigPath("credentials")

  if _, err := os.Stat(credentialsPath); os.IsNotExist(err) {
    return nil, core.InternalError("Not logged in", nil)
  }

  session := Aux4Session{}

  err := aux4IO.ReadJsonFile(credentialsPath, &session)
  if err != nil {
    return nil, core.InternalError("Error reading credentials", err)
  }

  return &session, nil
}

func Logout() error {
  session, _ := GetSession()
  if session == nil {
    return nil
  }

  credentialsPath := config.GetConfigPath("credentials")
  err := os.Remove(credentialsPath)
  if err != nil {
    return core.InternalError("Error deleting credentials", err)
  }

  url := api + "/users/me/sessions"

  client := &http.Client{}

  request, err := http.NewRequest("DELETE", url, nil)
  if err != nil {
    return err
  } 

  request.Header.Set("User-Agent", aux4.GetUserAgent())
  request.Header.Set("Authorization", "Bearer " + session.AccessToken)

  response, err := client.Do(request)
  if err != nil {
    return err
  }

  if response.StatusCode != 200 {
    return core.InternalError("Logout failed", nil)
  }

  return nil
}

type Aux4Login struct {
	Otp string `json:"otp"`
}

type Aux4Session struct {
	AccessToken string `json:"accessToken"`
	SessionId   string `json:"sessionId"`
}
