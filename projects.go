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
	"strings"
	"time"

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

func LiveEvaluation(files []string, config *Config) error {
	// Create an array with files content
	depFiles := make([]DependencyFile, len(files))
	for i, file := range files {
		depFile := DependencyFile{Name: file}
		content, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		depFile.Content = content
		depFiles[i] = depFile
	}

	requestDeps := map[string][]DependencyFile{"dependency_files": depFiles}
	depFilesJSON, err := json.Marshal(requestDeps)
	if err != nil {
		return err
	}
	client := &http.Client{}
	req, err := http.NewRequest("POST", config.APIEndpoint+LIVE_EVAL_PATH, bytes.NewReader(depFilesJSON))
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

	// Wait until job is done
	url := fmt.Sprintf("%s%s/%s", config.APIEndpoint, LIVE_EVAL_PATH, jsonResp["job_id"])
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth("x", config.APIKey)
	req.Header.Add("Content-Type", "application/json")
	var response struct {
		Status string `json:"status"`
		Result struct {
			RuntimeStatus     string       `json:"runtime_status"`
			DevelopmentStatus string       `json:"development_status"`
			Dependencies      []Dependency `json:"dependencies"`
		} `json:"result"`
	}
	var iter int // used to display the little dots for each loop bellow
	for {
		// use the same request again and again
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
			response.Status = "error"
		}

		if err = json.Unmarshal(body, &response); err != nil {
			return err
		}

		if !config.RawFormat { // don't display status if RawFormat
			iter += 1
			fmt.Printf("\rJob Status: %s%s", response.Status, strings.Repeat(".", iter))
		}
		if response.Status != "working" && response.Status != "queued" { // Job has completed or failed or whatever
			if config.RawFormat {
				fmt.Printf("%s\n", body)
				return nil
			}
			break
		}
		// Wait 1s before trying again
		time.Sleep(time.Second * 1)
	}

	color.Println(fmt.Sprintf("\n\n%-12.12s %s", "Run. Status", statusDots(response.Result.RuntimeStatus)))
	color.Println(fmt.Sprintf("%-12.12s %s\n\n", "Dev. Status", statusDots(response.Result.DevelopmentStatus)))

	// Display deps in an ascii table
	renderDepsAsTable(response.Result.Dependencies, os.Stdout)
	return nil
}

// Return unicode colorized text dots for each status
// Status is supposed to red|yellow|green otherwise "none" will be returned
func statusDots(status string) string {
	var dots string
	switch status {
	case "red":
		dots = "@k\u2B24 @k\u2B24 @r\u2B24  @{|}(red)"
	case "yellow":
		dots = "@k\u2B24 @y\u2B24 @k\u2B24  @{|}(yellow)"
	case "green":
		dots = "@g\u2B24 @k\u2B24 @k\u2B24  @{|}(green)"
	default:
		dots = "@k\u2B24 @k\u2B24 @k\u2B24  @{|}(none)"
	}
	return dots
}
