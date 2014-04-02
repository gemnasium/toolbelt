package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/codegangsta/cli"
)

const (
	CREATE_PROJECT_PATH = "/projects"
)

// Create a new project on gemnasium.
// The first arg is used as the project name.
// If no arg is provided, the user will be prompted to enter a project name.
func CreateProject(ctx *cli.Context, config *Config, r io.Reader) error {
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
	var description string
	fmt.Printf("Enter project description: ")
	_, err := fmt.Fscanf(r, "%s", &description)
	if err != nil {
		return err
	}
	fmt.Println("") // quickfix for goconvey

	projectAsJson, err := json.Marshal(&map[string]string{"name": project, "description": description})
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
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Server returned non-200 status: %v\n", resp.Status)
	}

	// Parse server response
	var proj map[string]interface{}
	if err := json.Unmarshal(body, &proj); err != nil {
		return err
	}
	fmt.Printf("Project '%s' created! (Remaining private slots: %v)\n", project, proj["remaining_slot_count"])
	fmt.Printf("Slug: %s\n", proj["slug"])
	return nil
}

func Changelog(package_name string) (string, error) {
	changelog := `
		# 1.2.3

		lot's of new features!
		`
	return changelog, nil
}
