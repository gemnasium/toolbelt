package commands

import (
	"github.com/codegangsta/cli"
	"github.com/gemnasium/toolbelt/models"
	"github.com/gemnasium/toolbelt/utils"
)

func DependencyAlertsList(ctx *cli.Context) {
	project, err := models.GetProject(ctx.Args().First())
	utils.ExitIfErr(err)
	err = models.ListDependencyAlerts(project)
	utils.ExitIfErr(err)
}
