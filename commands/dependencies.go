package commands

import (
	"github.com/gemnasium/toolbelt/models"
	"github.com/urfave/cli"
)

func DependenciesList(ctx *cli.Context) error {
	project, err := models.GetProject(ctx.Args().First())
	if err != nil {
		return err
	}
	err = models.ListDependencies(project)
	return err
}
