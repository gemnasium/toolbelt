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
	"github.com/wsxiaoys/terminal/color"
	"gopkg.in/yaml.v1"
)

const (
	LIST_PROJECTS_PATH         = "/projects"
	CREATE_PROJECT_PATH        = "/projects"
	SUPPORTED_DEPENDENCY_FILES = `(Gemfile|Gemfile\.lock|.*\.gemspec|package\.json|npm-shrinkwrap\.json|setup\.py|requirements\.txt|requires\.txt|composer\.json|composer\.lock)$`
)

type Project struct {
	Name              string `json:"name,omitempty"`
	Slug              string `json:"slug,omitempty"`
	Description       string `json:"description,omitempty"`
	Origin            string `json:"origin,omitempty"`
	Private           bool   `json:"private,omitempty"`
	Status            string `json:"status,omitempty"`
	Monitored         bool   `json:"monitored,omitempty"`
	UnmonitoredReason string `json:"unmonitored_reason,omitempty"`
}

// List projects on gemnasium
// TODO: Add a flag to display unmonitored projects too
func ListProjects(config *Config) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", config.APIEndpoint+LIST_PROJECTS_PATH, nil)
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
	var private string
	for owner, _ := range projects {
		MonitoredProjectsCount := 0
		if owner != "owned" {
			fmt.Printf("\nShared by: %s\n\n", owner)
		}
		for _, project := range projects[owner] {
			if !project.Monitored {
				continue
			}
			if project.Private {
				private = "[private]"
			} else {
				private = "" // reset
			}
			fmt.Printf("  %s: \"%s\" %s\n", project.Slug, project.Name, private)
			MonitoredProjectsCount += 1
		}
		color.Printf("@{g!}Found %d projects (%d unmonitored are hidden)\n\n", MonitoredProjectsCount, len(projects[owner])-MonitoredProjectsCount)
	}
	return nil
}

// Display project details
// Project is retrieved from its slug
func GetProject(slug string, config *Config) error {
	if slug == "" {
		return errors.New("[slug] can't be empty")
	}
	client := &http.Client{}
	url := fmt.Sprintf("%s/projects/%s", config.APIEndpoint, slug)
	req, err := http.NewRequest("GET", url, nil)
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

// Create a new project on gemnasium.
// The first arg is used as the project name.
// If no arg is provided, the user will be prompted to enter a project name.
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
	var jsonResp map[string]interface{}
	if err := json.Unmarshal(body, &jsonResp); err != nil {
		return err
	}
	fmt.Printf("Project '%s' created! (Remaining private slots: %v)\n", project.Name, jsonResp["remaining_slot_count"])
	fmt.Printf("To configure this project, use the following command:\ngemnasium projects configure %s\n", jsonResp["slug"])
	return nil
}

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

type DependencyFile struct {
	Filename string
	SHA      string
	Content  string
}

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
			deps = append(deps, DependencyFile{Filename: info.Name(), SHA: "sha", Content: "content"})
		}
		return nil
	}
	filepath.Walk(".", searchDeps)
	fmt.Printf("deps %+v\n", deps)
	return nil
}

func Changelog(package_name string) (string, error) {
	changelog := `
		# 1.2.3

		lot's of new features!
		`
	return changelog, nil
}

func statusDots(status string) string {
	var dots string
	switch status {
	case "red":
		dots = "@{k!}\u2B24 @{k!}\u2B24 @{r!}\u2B24"
	case "yellow":
		dots = "@{k!}\u2B24 @{y!}\u2B24 @{k!}\u2B24"
	case "green":
		dots = "@{g!}\u2B24 @{k!}\u2B24 @{k!}\u2B24"
	default:
		dots = "@{k!}\u2B24 @{k!}\u2B24 @{k!}\u2B24"
	}
	return dots

}
