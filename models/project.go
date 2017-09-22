package models

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/gemnasium/toolbelt/config"
	"github.com/gemnasium/toolbelt/gemnasium"
	"github.com/gemnasium/toolbelt/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/wsxiaoys/terminal/color"
	"gopkg.in/yaml.v1"
)

type Project struct {
	Name              string `json:"name,omitempty"`
	Slug              string `json:"slug,omitempty"`
	Description       string `json:"description,omitempty"`
	Origin            string `json:"origin,omitempty"`
	Private           bool   `json:"private,omitempty"`
	Color             string `json:"color,omitempty"`
	Monitored         bool   `json:"monitored,omitempty"`
	UnmonitoredReason string `json:"unmonitored_reason,omitempty"`
	CommitSHA         string `json:"commit_sha"`
}

// List projects on gemnasium
// TODO: Add a flag to display unmonitored projects too
func ListProjects(privateProjectsOnly bool) error {
	var projects map[string][]Project
	opts := &gemnasium.APIRequestOptions{
		Method: "GET",
		URI:    "/projects",
		Result: &projects,
	}
	err := gemnasium.APIRequest(opts)
	if err != nil {
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
func (p *Project) Show() error {
	err := p.Fetch()
	if err != nil {
		return err
	}
	if config.RawFormat {
		return nil
	}

	color.Println(fmt.Sprintf("%s: %s\n", p.Name, utils.StatusDots(p.Color)))
	table := tablewriter.NewWriter(os.Stdout)
	table.SetRowLine(true)

	table.Append([]string{"Slug", p.Slug})
	table.Append([]string{"Description", p.Description})
	table.Append([]string{"Origin", p.Origin})
	table.Append([]string{"Private", strconv.FormatBool(p.Private)})
	table.Append([]string{"Monitored", strconv.FormatBool(p.Monitored)})
	if !p.Monitored {
		table.Append([]string{"Unmonitored reason", p.UnmonitoredReason})
	}

	table.Render()
	return nil
}

// Update project details
// http://docs.gemnasium.apiary.io/#patch-%2Fprojects%2F%7Bslug%7D
func (p *Project) Update(name, desc *string, monitored *bool) error {
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
	opts := &gemnasium.APIRequestOptions{
		Method: "PATCH",
		URI:    fmt.Sprintf("/projects/%s", p.Slug),
		Body:   update,
	}
	err := gemnasium.APIRequest(opts)
	if err != nil {
		return err
	}

	color.Printf("@gProject %s updated succesfully\n", p.Slug)
	return nil
}

// Create a new project on gemnasium.
// The first arg is used as the project name.
// If no arg is provided, the user will be prompted to enter a project name.
// http://docs.gemnasium.apiary.io/#post-%2Fprojects
func CreateProject(projectName string, r io.Reader) error {
	project := &Project{Name: projectName}
	if project.Name == "" {
		fmt.Printf("Enter project name: ")
		_, err := fmt.Scanln(&project.Name)
		if err != nil {
			return err
		}
	}
	fmt.Printf("Enter project description: ")
	scanner := bufio.NewScanner(r)
	scanner.Scan()
	project.Description = scanner.Text()
	fmt.Println("") // quickfix for goconvey

	var jsonResp map[string]interface{}
	opts := &gemnasium.APIRequestOptions{
		Method: "POST",
		URI:    "/projects",
		Body:   project,
		Result: &jsonResp,
	}
	err := gemnasium.APIRequest(opts)
	if err != nil {
		return err
	}

	fmt.Printf("Project '%s' created: https://gemnasium.com/%s (Remaining slots: %v)\n", project.Name, jsonResp["slug"], jsonResp["remaining_slot_count"])
	fmt.Printf("To configure this project, use the following command:\ngemnasium configure %s\n", jsonResp["slug"])
	return nil
}

// Create a project config gile (.gemnasium.yml)
func (p *Project) Configure(slug string, r io.Reader, w io.Writer) error {
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
	_, err = w.Write(body)
	if err != nil {
		return err
	}
	color.Println("@gYour .gemnasium.yml was created!")
	return nil
}

// Start project synchronization
// http://docs.gemnasium.apiary.io/#post-%2Fprojects%2F%7Bslug%7D%2Fsync
func (p *Project) Sync() error {
	opts := &gemnasium.APIRequestOptions{
		Method: "POST",
		URI:    fmt.Sprintf("/projects/%s/sync", p.Slug),
	}
	err := gemnasium.APIRequest(opts)
	if err != nil {
		return err
	}

	color.Printf("@gSynchronization started for project %s\n", p.Slug)
	return nil
}

func (p *Project) Fetch() error {
	opts := &gemnasium.APIRequestOptions{
		Method: "GET",
		URI:    fmt.Sprintf("/projects/%s", p.Slug),
		Result: p,
	}
	return gemnasium.APIRequest(opts)
}

func (p *Project) Dependencies() (deps []Dependency, err error) {
	opts := &gemnasium.APIRequestOptions{
		Method: "GET",
		URI:    fmt.Sprintf("/projects/%s/dependencies", p.Slug),
		Result: &deps,
	}
	err = gemnasium.APIRequest(opts)
	return deps, err
}

// Fetch and return the dependency files ([]DependecyFile) for the current project
func (p *Project) DependencyFiles() (dfiles []DependencyFile, err error) {
	opts := &gemnasium.APIRequestOptions{
		Method: "GET",
		URI:    fmt.Sprintf("/projects/%s/dependency_files", p.Slug),
		Result: &dfiles,
	}
	err = gemnasium.APIRequest(opts)
	return dfiles, err
}

// Return a new Project with Slug set.
// The slugs in param are tried in order.
func GetProject(slugs ...string) (*Project, error) {
	slug := config.ProjectSlug
	for _, s := range slugs {
		if s != "" {
			slug = s
		}
	}
	if slug == "" {
		return nil, errors.New("[project slug] can't be empty")
	}
	return &Project{Slug: slug}, nil
}
