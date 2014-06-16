package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"

	"github.com/codegangsta/cli"
	"github.com/olekukonko/tablewriter"
	"github.com/wsxiaoys/terminal/color"
	"gopkg.in/yaml.v1"
)

const (
	LIST_PROJECTS_PATH         = "/projects"
	CREATE_PROJECT_PATH        = "/projects"
	LIVE_EVAL_PATH             = "/evaluate"
	SUPPORTED_DEPENDENCY_FILES = `(Gemfile|Gemfile\.lock|.*\.gemspec|package\.json|npm-shrinkwrap\.json|setup\.py|requirements\.txt|requires\.txt|composer\.json|composer\.lock)$`
)

// List projects on gemnasium
// TODO: Add a flag to display unmonitored projects too
func ListProjects(config *Config, privateProjectsOnly bool) error {
	client := &http.Client{}
	url := config.APIEndpoint + LIST_PROJECTS_PATH
	req, err := NewAPIRequest("GET", url, config.APIKey, nil)
	if err != nil {
		return err
	}
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

	// if RawFormat flag is set, don't format the output
	if config.RawFormat {
		fmt.Printf("%s", body)
		return nil
	}

	// Parse server response
	var projects map[string][]Project
	if err := json.Unmarshal(body, &projects); err != nil {
		return err
	}
	for owner, _ := range projects {
		MonitoredProjectsCount := 0
		if owner != "owned" {
			fmt.Printf("\nShared by: %s\n\n", owner)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "Slug", "Private"})
		for _, project := range projects[owner] {
			if !project.Monitored || (!project.Private && privateProjectsOnly) {
				continue
			}

			var private string
			if project.Private {
				private = "private"
			} else {
				private = ""
			}
			table.Append([]string{project.Name, project.Slug, private})
			MonitoredProjectsCount += 1
		}
		table.Render()
		color.Printf("@{g!}Found %d projects (%d unmonitored are hidden)\n\n", MonitoredProjectsCount, len(projects[owner])-MonitoredProjectsCount)
	}
	return nil
}

// Display project details
// http://docs.gemnasium.apiary.io/#get-%2Fprojects%2F%7Bslug%7D
func ShowProject(slug string, config *Config) error {
	if slug == "" {
		return errors.New("[slug] can't be empty")
	}
	client := &http.Client{}
	url := fmt.Sprintf("%s/projects/%s", config.APIEndpoint, slug)
	req, err := NewAPIRequest("GET", url, config.APIKey, nil)
	if err != nil {
		return err
	}
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

	// if RawFormat flag is set, don't format the output
	if config.RawFormat {
		fmt.Printf("%s", body)
		return nil
	}

	// Parse server response
	var project Project
	if err := json.Unmarshal(body, &project); err != nil {
		return err
	}
	s := reflect.ValueOf(&project).Elem()
	typeOfT := s.Type()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		if typeOfT.Field(i).Name == "Status" {
			color.Println(fmt.Sprintf("%18.18s: %s", "Status", statusDots(project.Status)))
		} else {
			fmt.Printf("%18.18s: %v\n", typeOfT.Field(i).Name, f.Interface())
		}
	}

	return nil
}

// Update project details
// http://docs.gemnasium.apiary.io/#patch-%2Fprojects%2F%7Bslug%7D
func UpdateProject(slug string, config *Config, name, desc *string, monitored *bool) error {
	if slug == "" {
		return errors.New("[slug] can't be empty")
	}

	if name == nil && desc == nil && monitored == nil {
		return errors.New("Please specify at least one thing to update (name, desc, or monitored")
	}

	update := make(map[string]interface{})
	if name != nil {
		update["name"] = *name
	}
	if desc != nil {
		update["desc"] = *desc
	}
	if monitored != nil {
		update["monitored"] = *monitored
	}
	projectAsJson, err := json.Marshal(update)
	if err != nil {
		return err
	}
	client := &http.Client{}
	url := fmt.Sprintf("%s/projects/%s", config.APIEndpoint, slug)
	req, err := NewAPIRequest("PATCH", url, config.APIKey, bytes.NewReader(projectAsJson))

	if err != nil {
		return err
	}
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

	// if RawFormat flag is set, don't format the output
	if config.RawFormat {
		fmt.Printf("%s", body)
		return nil
	}

	color.Printf("@gProject %s updated succesfully\n", slug)

	return nil
}

// Create a new project on gemnasium.
// The first arg is used as the project name.
// If no arg is provided, the user will be prompted to enter a project name.
// http://docs.gemnasium.apiary.io/#post-%2Fprojects
func CreateProject(projectName string, config *Config, r io.Reader) error {
	project := &Project{Name: projectName}
	if project.Name == "" {
		fmt.Printf("Enter project name: ")
		_, err := fmt.Scanln(&project.Name)
		if err != nil {
			return err
		}
	}
	fmt.Printf("Enter project description: ")
	_, err := fmt.Fscanf(r, "%s", &project.Description)
	if err != nil {
		return err
	}
	fmt.Println("") // quickfix for goconvey

	projectAsJson, err := json.Marshal(project)
	if err != nil {
		return err
	}
	client := &http.Client{}
	req, err := NewAPIRequest("POST", config.APIEndpoint+CREATE_PROJECT_PATH, config.APIKey, bytes.NewReader(projectAsJson))
	if err != nil {
		return err
	}
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
	var jsonResp map[string]interface{}
	if err := json.Unmarshal(body, &jsonResp); err != nil {
		return err
	}
	fmt.Printf("Project '%s' created! (Remaining private slots: %v)\n", project.Name, jsonResp["remaining_slot_count"])
	fmt.Printf("To configure this project, use the following command:\ngemnasium projects configure %s\n", jsonResp["slug"])
	return nil
}

// Create a project config gile (.gemnasium.yml)
func ConfigureProject(slug string, config *Config, r io.Reader, f *os.File) error {

	if slug == "" {
		fmt.Printf("Enter project slug: ")
		_, err := fmt.Scanln(&slug)
		if err != nil {
			return err
		}
	}

	// We just create a file with project_config for now.
	projectConfig := &map[string]string{"project_slug": slug}
	body, err := yaml.Marshal(&projectConfig)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	// write content to the file
	_, err = f.Write(body)
	if err != nil {
		return err
	}
	// Issue a Sync to flush writes to stable storage.
	f.Sync()
	return nil
}

// Push project dependencies
// Not yet implemented and WIP
func PushDependencies(ctx *cli.Context, config *Config) error {
	deps := []DependencyFile{}
	searchDeps := func(path string, info os.FileInfo, err error) error {

		// Skip excluded paths
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}
		matched, err := regexp.MatchString(SUPPORTED_DEPENDENCY_FILES, info.Name())
		if err != nil {
			return err
		}

		if matched {
			fmt.Printf("[debug] Found: %s\n", info.Name())
			deps = append(deps, DependencyFile{Name: info.Name(), SHA: "sha", Content: []byte("content")})
		}
		return nil
	}
	filepath.Walk(".", searchDeps)
	fmt.Printf("deps %+v\n", deps)
	return nil
}

// Start project synchronization
// http://docs.gemnasium.apiary.io/#post-%2Fprojects%2F%7Bslug%7D%2Fsync
func SyncProject(projectSlug string, config *Config) error {
	if projectSlug == "" {
		return errors.New("[projectSlug] can't be empty")
	}
	client := &http.Client{}
	url := fmt.Sprintf("%s/projects/%s/sync", config.APIEndpoint, projectSlug)
	req, err := NewAPIRequest("POST", url, config.APIKey, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("Server returned non-200 status: %v\n", resp.Status)
	}

	color.Printf("@gSynchronization started for project %s\n", projectSlug)
	return nil
}
