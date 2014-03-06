package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/codegangsta/cli"
	"net/http"
	"os"
)

const (
	CREATE_PROJECT_PATH = "/projects"
)

type ProjectParams struct {
	Name   string `json:"name"`
	Branch string `json:"branch"`
}

func CreateProject(ctx *cli.Context, config *Config) error {
	project := ctx.Args().First()
	if project == "" {
		fmt.Printf("Enter project name: ")
		_, err := fmt.Scanln(&project)
		if err != nil {
			return err
		}
	}

	projectParams := &ProjectParams{Name: project, Branch: "master"}
	projectAsJson, err := json.Marshal(projectParams)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	res, err := http.Post(config.APIEndpoint+CREATE_PROJECT_PATH, "application/json", bytes.NewReader(projectAsJson))
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Server returned non-200 status:  %v\n", res.Status)
		os.Exit(1)
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
