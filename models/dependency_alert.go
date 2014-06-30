package models

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gemnasium/toolbelt/gemnasium"
	"github.com/olekukonko/tablewriter"
)

type DependencyAlert struct {
	ID       int       `json:"id"`
	Advisory Advisory  `json:"advisory"`
	OpenAt   time.Time `json:"open_at"`
	Status   string    `json:"status"`
}

func ListDependencyAlerts(project *Project) error {
	var alerts []DependencyAlert
	opts := &gemnasium.APIRequestOptions{
		Method: "GET",
		URI:    fmt.Sprintf("/projects/%s/alerts", project.Slug),
		Result: &alerts,
	}
	err := gemnasium.APIRequest(opts)
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Advisory", "Date", "Status"})

	table.SetAlignment(tablewriter.ALIGN_LEFT) // table is lost when ID have 2 or 3 digits...
	for _, alert := range alerts {
		table.Append([]string{strconv.Itoa(alert.Advisory.ID), alert.OpenAt.Format(time.RFC822), alert.Status})
	}
	table.Render() // Send output
	return nil
}
