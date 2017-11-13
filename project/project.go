package project

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/gemnasium/toolbelt/config"
	"github.com/gemnasium/toolbelt/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/wsxiaoys/terminal/color"
	"gopkg.in/yaml.v1"
	"github.com/gemnasium/toolbelt/api"
	"bufio"
)

// List projects on gemnasium
// TODO: Add a flag to display unmonitored projects too
func ListProjects(privateProjectsOnly bool) (err error) {
	var projects map[string][]api.Project
	projects, err = api.APIImpl.ProjectList(privateProjectsOnly)
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
func ProjectShow(p *api.Project) error {
	err := ProjectFetch(p)
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
	table.Append([]string{"Name", p.Name})
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
func ProjectUpdate(p *api.Project, name, desc *string, monitored *bool) error {
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
	err := api.APIImpl.ProjectUpdate(p, update)
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
	project := &api.Project{Name: projectName}
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

	jsonResp, err := api.APIImpl.ProjectCreate(project)
	if err != nil {
		return err
	}

	fmt.Printf("Project '%s' created: https://gemnasium.com/%s (Remaining slots: %v)\n", project.Name, jsonResp["slug"], jsonResp["remaining_slot_count"])
	fmt.Printf("To configure this project, use the following command:\ngemnasium configure %s\n", jsonResp["slug"])
	return nil
}

// Create a project config gile (.gemnasium.yml)
func ProjectConfigure(p *api.Project, slug string, r io.Reader, w io.Writer) error {
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
func ProjectSync(p *api.Project) (err error) {
	err = api.APIImpl.ProjectSync(p)
	if err != nil {
		return err
	}

	color.Printf("@gSynchronization started for project %s\n", p.Slug)
	return nil
}

func ProjectFetch(p *api.Project) (err error) {
	err = api.APIImpl.ProjectFetch(p)
	return err
}

func ProjectDependencies(p *api.Project) (deps []api.Dependency, err error) {
	deps, err = api.APIImpl.ProjectGetDependencies(p)
	return deps, err
}

// Fetch and return the dependency files ([]DependecyFile) for the current project
func ProjectDependencyFiles(p *api.Project) (dfiles []api.DependencyFile, err error) {
	dfiles, err = api.APIImpl.ProjectGetDependencyFiles(p)
	return dfiles, err
}

// Return a new Project with Slug set.
// The slugs in param are tried in order.
func GetProject(slugs ...string) (*api.Project, error) {
	slug := config.ProjectSlug
	for _, s := range slugs {
		if s != "" {
			slug = s
		}
	}
	if slug == "" {
		return nil, errors.New("[project slug] can't be empty")
	}
	return &api.Project{Slug: slug}, nil
}
