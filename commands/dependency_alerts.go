package commands

import (
	"github.com/gemnasium/toolbelt/models"
	"github.com/urfave/cli"
)

func DependencyAlertsList(ctx *cli.Context) error {
	project, err := models.GetProject(ctx.Args().First())
	if err != nil {
		return err
	}

	err = models.ListDependencyAlerts(project)
	return err
}
