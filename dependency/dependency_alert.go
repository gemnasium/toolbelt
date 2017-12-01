package dependency

import (
	"os"
	"strconv"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/gemnasium/toolbelt/api"
)

func ListDependencyAlerts(p *api.Project) error {
	// V1 and V2 return different informations
	switch a := api.APIImpl.(type) {
	case *api.APIv1:
		alerts, err := api.APIImpl.DependencyAlertsGet(p)
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
	case *api.V2ToV1:
		v2p := api.V2Project{}
		api.V1ProjectToV2(p, &v2p)
		alerts, err := a.APIv2.DependencyAlertsGet(&v2p)
		if err != nil {
			return err
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Advisory", "Date", "Status"})

		table.SetAlignment(tablewriter.ALIGN_LEFT) // table is lost when ID have 2 or 3 digits...
		for _, alert := range alerts {
			table.Append([]string{alert.Advisory.Identifier, alert.Advisory.Date.Format(time.RFC822), alert.Status})
		}
		table.Render() // Send output
	}

	return nil
}
