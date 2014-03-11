package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/codegangsta/cli"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	CREATE_PROJECT_PATH = "/projects"
	LOGIN_PATH          = "/login"
)

// Login with the user email and password
//
func Login(ctx *cli.Context, config *Config) error {
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

func Logout(ctx *cli.Context, config *Config) error {
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

// Create a new project on gemnasium.
// The first arg is used as the project name.
// If no arg is provided, the user will be prompted to enter a project name.
func CreateProject(ctx *cli.Context, config *Config) error {
	if err := AttemptLogin(ctx, config); err != nil {
		return err
	}
	project := ctx.Args().First()
	if project == "" {
		fmt.Printf("Enter project name: ")
		_, err := fmt.Scanln(&project)
		if err != nil {
			return err
		}
	}

	projectAsJson, err := json.Marshal(&map[string]string{"name": project, "branch": "master"})
	if err != nil {
		return err
	}
	client := &http.Client{}
	req, err := http.NewRequest("POST", config.APIEndpoint+CREATE_PROJECT_PATH, bytes.NewReader(projectAsJson))
	if err != nil {
		return err
	}
	req.SetBasicAuth("x", config.APIKey)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Server returned non-200 status: %v\n", resp.Status)
	}

	fmt.Printf("Project '%s' created!\n", project)
	return nil
}

func Changelog(package_name string) (string, error) {
	changelog := `
		# 1.2.3

		lot's of new features!
		`
	return changelog, nil
}
