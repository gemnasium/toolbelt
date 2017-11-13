package dependency

import (
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/gemnasium/toolbelt/api"
	"github.com/gemnasium/toolbelt/project"
)

// http://docs.gemnasium.apiary.io/#dependencies
func ListDependencies(p *api.Project) error {
	deps, err := project.ProjectDependencies(p)
	if err != nil {
		return err
	}

	RenderDepsAsTable(deps, os.Stdout)
	return nil
}

// Display deps in an ascii table
func RenderDepsAsTable(deps []api.Dependency, output io.Writer) {
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
