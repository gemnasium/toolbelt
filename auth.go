package main

import (
	"encoding/json"

	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	LOGIN_PATH = "/login"
)

// Login with the user email and password
// An entry will be created in ~/.netrc on successful login.
func Login(config *Config) error {
	// Create a function to be overriden in tests
	email, password, err := getCredentials()
	if err != nil {
		return err
	}

	loginAsJson, err := json.Marshal(map[string]string{"email": email, "password": password})
	if err != nil {
		return err
	}
	resp, err := http.Post(config.APIEndpoint+LOGIN_PATH, "application/json", bytes.NewReader(loginAsJson))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Server returned non-200 status: %v\n", resp.Status)
	}

	// Read api token from response
	// body will be of the form:
	// {"api_token": "abcxzy123"}
	body, err := ioutil.ReadAll(resp.Body)

	var response_body map[string]string
	if err := json.Unmarshal(body, &response_body); err != nil {
		//return err
	}
	api_token := response_body["api_token"]

	err = saveCreds(strings.Split(resp.Request.Host, ":")[0], email, api_token)
	if err != nil {
		printFatal("saving new token: " + err.Error())
	}
	fmt.Println("Logged in.")

	return nil
}

// Logout doesn't hit the API of course.
// It simply removes the corresponding entry in ~/.netrc
func Logout(config *Config) error {
	api_url, err := url.Parse(config.APIEndpoint)
	if err != nil {
		return err
	}
	err = removeCreds(api_url.Host)
	if err != nil {
		return err
	}
	fmt.Println("Logged out.")
	return nil
}
