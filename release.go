package main

import (
  "net/http"
  "encoding/json"
)

type Release struct {
  Name string `json:"name"`
}

func GetLatestRelease() string {
  url := "https://api.github.com/repos/aux4/aux4/releases/latest"

  resp, err := http.Get(url)
  if err != nil {
    return ""
  }

  defer resp.Body.Close()

  var release Release
  err = json.NewDecoder(resp.Body).Decode(&release)
  if err != nil {
    return ""
  }

  return release.Name
}
