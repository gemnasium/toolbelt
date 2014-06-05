package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
)

// http://docs.gemnasium.apiary.io/#dependencies
func ListDependencies(projectSlug string, config *Config) error {
	if projectSlug == "" {
		return errors.New("[projectSlug] can't be empty")
	}
	client := &http.Client{}
	url := fmt.Sprintf("%s/projects/%s/dependencies", config.APIEndpoint, projectSlug)
	req, err := NewAPIRequest("GET", url, config.APIKey, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Server returned non-200 status: %v\n", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// if RawFormat flag is set, don't format the output
	if config.RawFormat {
		fmt.Printf("%s", body)
		return nil
	}

	// Parse server response
	var deps []Dependency
	if err := json.Unmarshal(body, &deps); err != nil {
		return err
	}

	renderDepsAsTable(deps, os.Stdout)
	return nil
}

// Display deps in an ascii table
func renderDepsAsTable(deps []Dependency, output io.Writer) {
	// Display deps in an ascii table
	table := tablewriter.NewWriter(output)
	// TODO: Add a "type" header in deps have more than 1 type
	table.SetHeader([]string{"Dependencies", "Requirements", "Locked", "Status", "Advisories"})

	for _, dep := range deps {
		// transform dep.Advisories to []string
		advisories := make([]string, len(dep.Advisories))
		for i, adv := range dep.Advisories {
			advisories[i] = strconv.Itoa(adv.ID)
		}
		sort.Strings(advisories)

		var levelPrefix string
		if !dep.FirstLevel {
			levelPrefix = "+-- "
		}
		table.Append([]string{levelPrefix + dep.Package.Name, dep.Requirement, dep.LockedVersion, dep.Color, strings.Join(advisories, ", ")})
	}
	table.Render() // Send output
}
