package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Release struct {
	Version string `json:"tag_name"`
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

	return strings.TrimPrefix(release.Version, "v")
}
