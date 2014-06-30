package models

import (
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
)

type Dependency struct {
	Requirement   string     `json:"requirement"`
	LockedVersion string     `json:"locked_version"`
	Package       Package    `json:"package"`
	Type          string     `json:"type"`
	FirstLevel    bool       `json:"first_level"`
	Color         string     `json:"color"`
	Advisories    []Advisory `json:"advisories,omitempty"`
}

// http://docs.gemnasium.apiary.io/#dependencies
func ListDependencies(project *Project) error {
	deps, err := project.Dependencies()
	if err != nil {
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
