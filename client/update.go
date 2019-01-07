package main

import (
	"io/ioutil"
	"net/http"
	"strings"

	version "github.com/hashicorp/go-version"
)

func needUpdate(latestVer string) bool {
	v1, err := version.NewVersion(Version)
	if err != nil {
		SayErr("Could not parse current version: %s", err)
		return false
	}
	v2, err := version.NewVersion(latestVer)
	if err != nil {
		SayErr("Could not parse latest version: %s", err)
		return false
	}
	return v1.LessThan(v2)
}

func CheckForUpdate(config *Config) {
	if Version == DevVersion {
		return
	}
	response, err := http.Get(config.GetLatestVersionUrl())
	if err != nil {
		SayErr("Could not check for new version: %s", err)
		return
	}
	statusCategory := response.StatusCode / 100
	if statusCategory != 2 {
		SayErr("Could not check for new version: %s", response.Status)
		return
	}
	bodyData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		SayErr("Could not check for new version: %s", err)
		return
	}
	latestVer := strings.TrimSpace(string(bodyData))
	if needUpdate(latestVer) {
		Warn("Your version of rdoctor %s is older than latest version %s.",
			Version, latestVer)
		Warn("Please get the latest version from the URL below.")
		Warn("")
		Warn("    %s", config.GetLatestClientUrl())
		Warn("")
	}
}
