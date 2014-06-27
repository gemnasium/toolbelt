package commands

import (
	"github.com/codegangsta/cli"
	"github.com/gemnasium/toolbelt/models"
	"github.com/gemnasium/toolbelt/utils"
)

func DependenciesList(ctx *cli.Context) {
	project, err := models.GetProject(ctx.Args().First())
	utils.ExitIfErr(err)
	err = models.ListDependencies(project)
	utils.ExitIfErr(err)
}
