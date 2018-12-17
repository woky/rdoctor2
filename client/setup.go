package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func promptForIdentity() string {
	SayOut("*** One-time setup ***")
	SayOut("Please fill in the following entries to get access to RDoctor API.")
	for {
		var name string
		for {
			name = Prompt("Your name or nick")
			if len(name) > 0 {
				break
			}
			SayOut("Name cannot be empty.")
		}
		email := Prompt("Your email (optional)")
		identity := fmt.Sprintf("%s <%s>", name, email)
		SayOut("Your API key will be tied to:")
		SayOut("    %s", identity)
		response := Prompt("Edit again? [y/N]")
		if response == "" || response == "n" || response == "N" {
			return identity
		}
	}
}

func requestNewApiKey(config *Config, identity string) string {
	PrintOut("Requesting new API key... ")
	response, err := http.PostForm(config.GetNewKeyUrl(identity), nil)
	fmt.Println("Done.")
	if err != nil {
		Die(err.Error())
	}
	statusCategory := response.StatusCode / 100
	if statusCategory != 2 {
		Die("Unsuccessful response: %s", response.Status)
	}
	bodyData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		Die("Failure while reading response body: %s", err)
	}
	return string(bodyData)
}

func ping(config *Config) bool {
	if !config.HasApiKey() {
		return false
	}
	response, err := http.PostForm(config.GetPingUrl(), nil)
	if err != nil {
		Die(err.Error())
	}
	status := response.StatusCode
	if status == 401 {
		return false
	}
	if status/100 == 2 {
		return true
	}
	Die("Unsuccessful service response: %s", response.Status)
	panic("")
}

func RunSetup(config *Config) {
	if ping(config) {
		return
	}

	identity := promptForIdentity()
	apiKey := requestNewApiKey(config, identity)

	SayOut("")
	SayOut("Please confirm your new API key in web browser at URL below.")
	SayOut("")
	SayOut("    %s", config.GetConfirmKeyUrl(apiKey))
	SayOut("")
	PrintOut("Waiting for confirmation... ")

	config.SetApiKey(apiKey)
	done := make(chan struct{})
	go func() {
		for {
			if ping(config) {
				done <- struct{}{}
				close(done)
				return
			}
			time.Sleep(3 * time.Second)
		}
	}()

	spinnerChars := []byte{0x2f, 0x2d, 0x5c, 0x7c}
	i := 0
	for {
		select {
		case <-done:
			fmt.Println("\b Done.")
			config.Save()
			SayOut("*** Setup finished ***")
			return
		default:
			fmt.Print("\b")
			fmt.Print(string(spinnerChars[i]))
			i = (i + 1) % len(spinnerChars)
			time.Sleep(500 * time.Millisecond)
		}
	}
}
