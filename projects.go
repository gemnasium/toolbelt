package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/codegangsta/cli"
)

const (
	CREATE_PROJECT_PATH = "/projects"
)

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
